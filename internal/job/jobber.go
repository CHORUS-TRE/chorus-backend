package job

import (
	"context"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"

	"go.uber.org/zap"
)

type Jobber interface {
	Register(name string, reg Registration)
	Run(ctx context.Context)
}

type jobber struct {
	ownerID       string
	enabled       bool
	running       bool
	lockStore     LockStore
	checkInterval time.Duration
	jitter        float64
	log           *logger.ContextLogger

	mu            sync.Mutex
	registrations map[string]Registration
	lastRun       map[string]time.Time
}

func NewJobber(ownerID string, enabled bool, lockStore LockStore, checkInterval time.Duration, jitter float64, log *logger.ContextLogger) Jobber {
	return &jobber{
		ownerID:       ownerID,
		enabled:       enabled,
		lockStore:     lockStore,
		checkInterval: checkInterval,
		jitter:        jitter,
		log:           log,
		registrations: make(map[string]Registration),
		lastRun:       make(map[string]time.Time),
	}
}

func (j *jobber) Register(name string, reg Registration) {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.registrations[name] = reg
}

func (j *jobber) Run(ctx context.Context) {
	if !j.enabled {
		j.log.Info(ctx, "jobber is disabled, not starting")
		return
	}

	if j.running {
		j.log.Warn(ctx, "jobber is already running")
		return
	}

	j.running = true

	j.log.Info(ctx, "jobber starting",
		zap.String("owner", j.ownerID),
		zap.Duration("check_interval", j.checkInterval),
		zap.Float64("jitter", j.jitter))

	// Initial random delay to desynchronize backends started simultaneously.
	initialDelay := time.Duration(float64(j.checkInterval) * j.jitter * rand.Float64())
	select {
	case <-time.After(initialDelay):
	case <-ctx.Done():
		return
	}

	ticker := time.NewTicker(j.checkInterval)
	defer ticker.Stop()

	for {
		j.tick(ctx)

		select {
		case <-ctx.Done():
			j.log.Info(ctx, "jobber stopping")
			return
		case <-ticker.C:
		}
	}
}

func (j *jobber) tick(ctx context.Context) {
	j.mu.Lock()
	names := make([]string, 0, len(j.registrations))
	for name := range j.registrations {
		names = append(names, name)
	}
	j.mu.Unlock()

	// Clean up expired locks.
	if n, err := j.lockStore.ReleaseExpired(ctx); err != nil {
		j.log.Error(ctx, "failed to release expired locks", zap.Error(err))
	} else if n > 0 {
		j.log.Info(ctx, "released expired job locks", zap.Int64("count", n))
	}

	for _, name := range names {
		j.mu.Lock()
		reg := j.registrations[name]
		last := j.lastRun[name]
		j.mu.Unlock()

		// Add per-job jitter (±jitter*interval) to spread load.
		jitterDuration := time.Duration(float64(reg.Interval) * j.jitter * (rand.Float64()*2 - 1))
		nextRun := last.Add(reg.Interval + jitterDuration)

		if time.Now().Before(nextRun) {
			continue
		}

		acquired, err := j.lockStore.TryAcquire(ctx, name, j.ownerID, reg.Timeout)
		if err != nil {
			j.log.Error(ctx, "failed to acquire job lock",
				zap.String("job", name),
				zap.Error(err))
			continue
		}

		if !acquired {
			continue
		}

		j.runJob(ctx, name, reg)
	}
}

func (j *jobber) runJob(ctx context.Context, name string, reg Registration) {
	j.log.Info(ctx, "running job", zap.String("job", name))

	start := time.Now()

	var jobCtx context.Context
	var cancel context.CancelFunc
	if reg.Timeout > 0 {
		jobCtx, cancel = context.WithTimeout(ctx, reg.Timeout)
		defer cancel()
	} else {
		jobCtx = ctx
	}

	msg, err := reg.Job.Do(jobCtx)
	status := StatusSuccess
	if err != nil {
		status = StatusFailure
		msg = err.Error()
	}

	elapsed := time.Since(start)
	j.log.Info(ctx, "job completed",
		zap.String("job", name),
		zap.String("status", status.String()),
		zap.String("message", msg),
		zap.Duration("duration", elapsed))

	j.mu.Lock()
	j.lastRun[name] = time.Now()
	j.mu.Unlock()

	// Use a context detached from the parent so the lock can be released
	// even when the backend is shutting down and ctx has been cancelled.
	releaseCtx := context.WithoutCancel(ctx)

	if err := j.lockStore.Release(releaseCtx, name, j.ownerID, status, msg); err != nil {
		j.log.Error(ctx, "failed to release job lock",
			zap.String("job", name),
			zap.Error(err))
	}
}
