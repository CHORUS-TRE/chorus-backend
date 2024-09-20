package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"time"

	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/uuid"
)

var timeRegex = regexp.MustCompile(`%[YydmHMS]`)

// NewOpenSearchWriteSyncer is a writer for a core. It writes log entries into
// an OpenSearch collector. Index names support date/time formatting using a
// subset of the strftime format, namely %Y, %y, %m, %d, %H, %M, %S (see the
// strftime doc).
func NewOpenSearchWriteSyncer(cfg config.Logger, signalCh chan<- os.Signal) (*BufferedWriteSyncer, error) {

	flusher, err := NewOpenSearchFlusher(&cfg)
	if err != nil {
		return nil, err
	}

	opts := []WriteSyncerOption{
		WithWriteSyncerSignal(signalCh),
	}

	if cfg.DisallowDropLog {
		opts = append(opts, WithNoDrop())
	}

	return NewBufferedWriteSyncer(&cfg, flusher, opts...), nil
}

// batcher is a wrapper of a flusher to allow a heavy number of entries to be
// flushed by batch and thus avoid HTTP body above the limit.
type batcher struct {
	Flusher

	// maxSize defines the maximum amount of bytes the entries can take in the
	// http body. Note that the default maximum size for OpenSearch is 100MB for
	// the content-length of the request.
	maxSize int
}

func (b batcher) Flush(entries [][]byte) error {
	if len(entries) == 0 {
		return nil
	}

	currentSize := 0
	offset := 0

	for i, e := range entries {
		currentSize += len(e)

		if currentSize > b.maxSize {
			// Flush the previous, so the current loop iteration is not included
			// as it goes above the limit.
			if err := b.Flusher.Flush(entries[offset:i]); err != nil {
				return err
			}

			offset = i
			currentSize = len(e)
		}
	}

	// Flush the remaining batch.
	if err := b.Flusher.Flush(entries[offset:]); err != nil {
		return err
	}

	return nil
}

type osFlusher struct {
	indexName string
	client    *opensearch.Client
}

func NewOpenSearchFlusher(cfg *config.Logger) (*batcher, error) {

	oscfg := opensearch.Config{
		Addresses: cfg.OpenSearchAddresses,
		Username:  cfg.OpenSearchUsername,
		Password:  string(cfg.OpenSearchPassword),
	}

	client, err := opensearch.NewClient(oscfg)
	if err != nil {
		return nil, err
	}

	f := &osFlusher{
		client:    client,
		indexName: cfg.OpenSearchIndexName,
	}

	b := &batcher{
		maxSize: 50 * 1024 * 1024, // 50MB
		Flusher: f,
	}

	return b, nil
}

func (f *osFlusher) Flush(entries [][]byte) error {

	index := prepareIndex(f.indexName, time.Now())

	req := opensearchapi.BulkRequest{
		Index: index,
		Body:  newBulkBody(entries, index),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Perform the request with the client.
	var res, err = req.Do(ctx, f.client)
	if err != nil {
		return fmt.Errorf("could not perform opensearch request: %w", err)
	}

	defer func() {
		_ = res.Body.Close()
	}()

	if res.IsError() {
		_, _ = io.Copy(os.Stderr, res.Body)
		return fmt.Errorf("bulk operation failed: %s", res.Status())
	}

	return nil
}

// bulkBody implements the io.Reader interface to sent a bulk request to
// ElasticSearch / OpenSearch.
type bulkBody struct {
	index        string
	offset       int
	writeEntry   bool
	writeLineEnd bool
	cmd          []byte
	entries      [][]byte
	randomFn     func() string
}

func newBulkBody(entries [][]byte, index string) *bulkBody {
	body := &bulkBody{
		entries:  make([][]byte, len(entries)),
		index:    index,
		randomFn: uuid.Next,
	}

	// the input of entries *must* not be altered thus a copy of the array is
	// made.
	copy(body.entries, entries)

	return body
}

func (r *bulkBody) Read(dest []byte) (int, error) {

	if !r.writeLineEnd && r.offset > len(r.entries)-1 {
		return 0, io.EOF
	}

	n := 0

	if !r.writeEntry && !r.writeLineEnd {
		if r.cmd == nil {
			r.cmd = []byte(fmt.Sprintf(`{"index": {"_index": "%s", "_id": "%s"} }%c`, r.index, r.randomFn(), '\n'))
		}

		var src []byte
		if len(r.cmd) > len(dest) {
			src = r.cmd[:len(dest)]
			r.cmd = r.cmd[len(dest):]
		} else {
			src = r.cmd
			r.cmd = nil
			r.writeLineEnd = false
			r.writeEntry = true
		}

		copy(dest, src)
		n += len(src)
	} else if !r.writeLineEnd {
		entry := r.entries[r.offset]
		var src []byte

		if len(entry) > len(dest) {
			// Note: here we change the entry slice but it does not change the
			// initial entry from the input.
			src = entry[:len(dest)]
			r.entries[r.offset] = entry[len(dest):]
		} else {
			src = entry
			r.offset++
			r.writeEntry = false
			r.writeLineEnd = true
		}

		copy(dest, src)
		n += len(src)
	} else {
		if len(dest) == 0 {
			return 0, nil
		}

		r.writeEntry = false
		r.writeLineEnd = false

		dest[0] = '\n'
		return 1, nil
	}

	return n, nil
}

// timeLayout converts the strftime layout into the respective Go layout.
var timeLayout = map[string]string{
	"%y": "06",
	"%Y": "2006",
	"%m": "01",
	"%d": "02",
	"%H": "15",
	"%M": "04",
	"%S": "05",
}

// prepareIndex takes an index and format any time key into the current UTC
// time.
func prepareIndex(index string, t time.Time) string {
	t = t.UTC()

	// a regexp is used because we cannot simply replace the time placeholder in
	// the index name as it could clash with part of the index.
	return timeRegex.ReplaceAllStringFunc(index, func(s string) string {
		return t.Format(timeLayout[s])
	})
}
