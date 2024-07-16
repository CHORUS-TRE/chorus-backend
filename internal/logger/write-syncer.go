package logger

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/retry.v1"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
)

//go:generate ../../scripts/tools/$GOOS/bin/mockery --name Flusher --disable-version-string

const (
	defaultQueueSize      = 100_000
	defaultPoolBufferSize = 256                    // 256 B
	defaultSize           = 2 * 1024 * 1024 * 1024 // 2 GiB
	defaultTickerInterval = 100 * time.Millisecond
)

var (
	ErrOverflow = errors.New("encountered a writer overflow")
)

type Flusher interface {
	Flush(entries [][]byte) error
}

type writeSyncerOptions struct {
	signalCh     chan<- os.Signal
	disallowDrop bool
	ticker       time.Duration
}

type WriteSyncerOption func(*writeSyncerOptions)

func WithWriteSyncerSignal(c chan<- os.Signal) WriteSyncerOption {
	return func(opts *writeSyncerOptions) {
		opts.signalCh = c
	}
}

func WithNoDrop() WriteSyncerOption {
	return func(wso *writeSyncerOptions) {
		wso.disallowDrop = true
	}
}

func WithTicker(t time.Duration) WriteSyncerOption {
	return func(wso *writeSyncerOptions) {
		wso.ticker = t
	}
}

type BufferedWriteSyncer struct {
	flusher Flusher

	pool   *bufferPool
	ticker *time.Ticker
	// fallback is the writer where log entries will be dumped when the buffer
	// is full.
	fallback io.Writer

	// bufferedLogEntries is used as a buffered channel to allow the log
	// entries to be written with a minimal time overhead.
	bufferedLogEntries chan []byte
	// force is used to force a flushing of the log entries.
	force chan struct{}
	// stop is used to stop the writer such that it refuses any new log entry.
	stop chan struct{}
	// done is used to indicate the closure of the writer.
	done chan struct{}
	// signal is used to send a SIGINT when the writer is configured to stop on
	// overflow.
	signal chan<- os.Signal

	size    int64
	maxSize int64
	stopped atomic.Value
	backoff retry.Strategy
	// allowDrop indicates when true that log entries can be dropped when the
	// buffer is full. Otherwise it dumps the buffer and stops the writer.
	allowDrop bool
}

func NewBufferedWriteSyncer(cfg *config.Logger, flusher Flusher, opts ...WriteSyncerOption) *BufferedWriteSyncer {

	o := writeSyncerOptions{ticker: defaultTickerInterval}
	for _, opt := range opts {
		opt(&o)
	}

	if cfg.BufferSize <= 0 {
		cfg.BufferSize = defaultSize
	}

	queueSize := defaultQueueSize
	if queueSize > cfg.BufferSize/defaultPoolBufferSize {
		// If the queue is bigger than the maximum buffer size, it could
		// overflow quickly by accepting all incoming log entries. The queue
		// size is thus reduced to allow flush to happen early.
		queueSize = cfg.BufferSize / defaultPoolBufferSize / 2
	}

	w := &BufferedWriteSyncer{
		flusher: flusher,

		pool:     newBufferPool(),
		ticker:   time.NewTicker(o.ticker),
		fallback: os.Stderr,

		bufferedLogEntries: make(chan []byte, queueSize),
		force:              make(chan struct{}),
		stop:               make(chan struct{}),
		done:               make(chan struct{}),
		signal:             o.signalCh,

		size:    0,
		maxSize: int64(cfg.BufferSize),
		stopped: atomic.Value{},
		backoff: retry.Exponential{
			Initial:  100 * time.Millisecond,
			Factor:   2,
			MaxDelay: 10 * time.Second,
		},
		allowDrop: !o.disallowDrop,
	}

	w.stopped.Store(false)

	go w.innerLoop()

	return w
}

func (ws *BufferedWriteSyncer) Write(data []byte) (int, error) {

	// Start by doing a copy of the data as it will be free at the end of the
	// function.
	copy := ws.pool.Get()
	copy = append(copy, data...)

	// atomically increase the current size of the buffer. Here we use the
	// capacity as the actual buffer may be bigger as it comes from the pool.
	// This allows for an accurate check of the memory used.
	currentSize := atomic.AddInt64(&ws.size, int64(cap(copy)))

	if currentSize >= ws.maxSize {
		// the system cannot follow the flux of log entries, we need to either
		// drop or stop the daemon.
		atomic.AddInt64(&ws.size, -int64(cap(copy)))

		ws.doOverflow(copy)

		return 0, ErrOverflow
	}

	select {

	case <-ws.stop:
		// the stop channel must be the highest priority.

		atomic.AddInt64(&ws.size, -int64(cap(copy)))

		return 0, io.EOF

	default:
		select {
		case ws.bufferedLogEntries <- copy:
			// log entry went through, so all good.

		default:
			// the buffered channel is full, so we force the flushing to try to
			// absorb the load.
			ws.force <- struct{}{}
			ws.bufferedLogEntries <- copy
		}
	}

	return len(data), nil
}

// Sync implements the zapcore.WriteSyncer interface. It will attempt to flush
// the log entries to the sink only once, and return an error if it fails.
func (ws *BufferedWriteSyncer) Sync() error {
	return ws.sync(false)
}

func (ws *BufferedWriteSyncer) sync(allowRetry bool) error {

	var queue [][]byte
	var attempt *retry.Attempt

	for {
		select {

		// Collect all log entries available.
		case log := <-ws.bufferedLogEntries:
			queue = append(queue, log)

		default:
			// no more messages, so we start to flush.
			if len(queue) == 0 {
				// .. no message to flush.
				return nil
			}

			err := ws.flush(queue)
			if err == nil || !allowRetry {
				return err
			}

			if ws.dumpOnStop(queue) {
				return io.EOF
			}

			if attempt == nil {
				attempt = retry.StartWithCancel(ws.backoff, nil, ws.stop)
			}

			// We must warn somehow that the sink is failing.
			fmt.Fprintf(ws.fallback, "unable to write logs to sink: %v\n", err)

			// return is not checked as we want to backoff indefinitely until
			// either the writer is stopped, or the sink is available again.
			attempt.Next()

			// if flushing fails, we gather again the log entries and retry
			// until it succeeds, or until the buffer is full.
			ws.drainForceChannel()
		}
	}
}

func (ws *BufferedWriteSyncer) flush(queue [][]byte) error {

	if err := ws.flusher.Flush(queue); err != nil {
		return err
	}

	size := int64(0)

	for _, buffer := range queue {

		size += int64(cap(buffer))

		// Free the buffers.
		ws.pool.Put(buffer)
	}

	atomic.AddInt64(&ws.size, -size)

	return nil
}

func (ws *BufferedWriteSyncer) Stop() error {

	if !ws.tryStop() {
		return nil
	}

	return ws.sync(true)
}

func (ws *BufferedWriteSyncer) tryStop() bool {

	if !ws.stopped.CompareAndSwap(false, true) {
		// already stopped.
		return false
	}

	ws.ticker.Stop()
	close(ws.stop)

	<-ws.done

	return true
}

func (ws *BufferedWriteSyncer) dumpOnStop(queue [][]byte) bool {
	select {

	case <-ws.stop:
		// Unable to flush but the writer must stops, so we dump the
		// queue and stop trying to sync.
		for _, buffer := range queue {
			_, _ = ws.fallback.Write(buffer)
		}

		return true

	default:
		return false
	}
}

func (ws *BufferedWriteSyncer) innerLoop() {
	defer close(ws.done)

	for {
		select {
		case <-ws.stop:
			ws.drainForceChannel()
			return

		case <-ws.ticker.C:
			// start flushing..

		case <-ws.force:
			// absorb a high load of log entries by flushing early.
		}

		// Here the error is ignored such that if redis is unavailable, the
		// buffer will be filled until the limit (or redis coming back) and
		// stopping the writer will return the error.
		_ = ws.sync(true)

		// avoid to trigger unnecessary forced flush.
		ws.drainForceChannel()
	}
}

func (ws *BufferedWriteSyncer) drainForceChannel() {
	for {
		select {
		case <-ws.force:
		default:
			return
		}
	}
}

func (ws *BufferedWriteSyncer) doOverflow(data []byte) {

	defer func() {
		_, _ = ws.fallback.Write(data)
	}()

	if ws.allowDrop {
		// Log is drop and dump to the fallback.
		return
	}

	ws.trySendSignal()

	// Stop and dump the remaining log entries to the fallback writer.
	_ = ws.Stop()
}

func (ws *BufferedWriteSyncer) trySendSignal() {
	select {
	case ws.signal <- os.Interrupt:
		// signal has been sent to the daemon.
	default:
		// either we don't have a channel configured, or the daemon is already
		// stopping.
	}
}

type bufferPool struct {
	p *sync.Pool
}

func (p *bufferPool) Get() []byte {
	buffer := p.p.Get().([]byte)
	buffer = buffer[:0]

	return buffer
}

func (p *bufferPool) Put(buffer interface{}) {
	p.p.Put(buffer)
}

func newBufferPool() *bufferPool {
	return &bufferPool{
		p: &sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, defaultPoolBufferSize)
			},
		},
	}
}
