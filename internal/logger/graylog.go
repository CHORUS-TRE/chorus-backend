package logger

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
)

// NewGraylogWriteSyncer is a writer for a core. It writes log entries into
// a Graylog collector.
//
// We support only the Graylog input of type "GELF HTTP"
// If clients have a recent enough version of Graylog, we strongly recommend that "Enable Bulk Receiving" be activated on the input
//
// CONFIG:
//   - `grayloghost` needs to be exact address of the graylog ingester (which ofter is at `/gelf`). e.g. http://localhost:12201/gelf
//   - `graylogbulkreceiving` should be set to true/false, in sync with "Enable Bulk Receiving" on the graylog GELF HTTP input config side. If those options are not in sync, the logger won't work. Defaults to false
//   - `graylogtimeout` is the timeout to use for http connections sending data to graylog (e.g. 5s). Defaults to 5s.
//   - `graylogauthorizeselfsignedcertificate` to authorize self signed certificate with TLS/HTTPS communication
func NewGraylogWriteSyncer(cfg config.Logger, signalCh chan<- os.Signal) (*BufferedWriteSyncer, error) {

	flusher, err := NewGraylogFlusher(&cfg)
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

type gelfFlusher struct {
	client   *http.Client
	url      string
	bulkMode bool
}

func NewGraylogFlusher(cfg *config.Logger) (*gelfFlusher, error) {
	timeout := cfg.GraylogTimeout
	if timeout.Nanoseconds() == 0 {
		timeout, _ = time.ParseDuration("5s")
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.GraylogAuthorizeSelfSignedCertificate}, //nolint:gosec
	}
	return &gelfFlusher{
		client: &http.Client{
			Transport: tr,
			Timeout:   timeout,
		},
		url:      cfg.GraylogHost,
		bulkMode: cfg.GraylogBulkReceiving,
	}, nil
}

// Flush sends the message to an http Graylog server
func (f *gelfFlusher) Flush(entries [][]byte) error {
	if f.bulkMode {
		err := f.flush(entries)
		if err != nil {
			return err
		}
	} else {
		for _, entry := range entries {
			err := f.flush([][]byte{entry})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *gelfFlusher) flush(entries [][]byte) error {
	// https://docs.graylog.org/v1/docs/ingest-gelf
	body := make([][]byte, 0, len(entries))

	for _, entry := range entries {
		gelfEntry := newGelfData(entry)
		bytes.ReplaceAll(gelfEntry, []byte("\n"), []byte("")) // Need to remove newlines inside individual messages. Necessary for bulk ingest
		body = append(body, gelfEntry)
	}

	payload := bytes.Join(body, []byte("\n"))

	var buf bytes.Buffer
	g := gzip.NewWriter(&buf)
	if _, err := g.Write(payload); err != nil {
		return errors.New("ERROR - can't gzip the log payload for Graylog (Write)")
	}
	if err := g.Close(); err != nil {
		return errors.New("ERROR - can't gzip the log payload for Graylog (Close)")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.url, &buf)
	if err != nil {
		return err
	}

	req.ContentLength = -1 // Chunking
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := f.client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	return nil
}

// newGelfData transforms a zap serialized message into a GELF serialized payload
func newGelfData(entry []byte) []byte {
	var payload map[string]interface{}
	gelfData := []byte{}

	err := json.Unmarshal(entry, &payload)
	if err != nil {
		return gelfData
	}
	gelfPayload := toGelfPayload(payload)
	gelfData, err = json.Marshal(gelfPayload)
	if err != nil {
		return []byte{}
	}

	return gelfData
}

// toGelfPayload takes a zap JSON and turns it into a GELF compatible payload
// e.g. { "version": "1.1", "host": "example.org", "short_message": "A short message", "level": 5, "_some_info": "foo" }
// see https://docs.graylog.org/docs/gelf
func toGelfPayload(payload map[string]interface{}) map[string]interface{} {
	gelfPayload := map[string]interface{}{"version": "1.1"}

	level := payload["level"]
	delete(payload, "level")
	gelfLevel := toGelfLevel(level)
	gelfPayload["level"] = gelfLevel

	ts := payload["ts"]
	delete(payload, "ts")
	gelfTs := toGelfTs(ts)
	if gelfTs != 0 {
		// If we could not determine a timestamp, graylog will put one on message reception
		gelfPayload["timestamp"] = gelfTs
	}

	gelfPayload["short_message"] = payload["msg"]
	delete(payload, "msg")

	gelfPayload["host"] = payload["cmp_id"]
	delete(payload, "cmp_id")

	// Add all remaining fields
	addCustomFields(gelfPayload, payload)

	return gelfPayload
}

func addCustomFields(gelfPayload map[string]interface{}, zapPayload map[string]interface{}) {
	for key, value := range zapPayload {
		gelfPayload["_"+key] = value
	}
}

// toGelfTs returns 0 when it fails
func toGelfTs(iTimestamp interface{}) int64 {
	strTs, ok := iTimestamp.(string)
	if !ok {
		return 0
	}
	layout := "2006-01-02T15:04:05.999999999Z0700" // based on RFC3339Nano, but without the ":" in the timezone
	ts, err := time.Parse(layout, strTs)
	if err != nil {
		return 0
	}
	return ts.Unix()
}

func toGelfLevel(iLevel interface{}) int {
	if iLevel == nil {
		return 3
	}

	strLevel, ok := iLevel.(string)
	if !ok {
		return 3
	}

	switch strLevel {
	case "debug":
		return 7
	case "info":
		return 6
	case "warning":
		return 4
	case "error":
		return 3
	case "fatal":
		return 2
	default:
		return 3
	}
}
