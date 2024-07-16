package logger

import (
	"fmt"
	"os"

	"github.com/go-redis/redis"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
)

type RedisFlusher struct {
	client *redis.Client
	key    string
}

func (f *RedisFlusher) Flush(entries [][]byte) error {
	payload := []any{}
	for _, entry := range entries {
		payload = append(payload, entry)
	}
	return f.client.RPush(f.key, payload...).Err()
}

func NewRedisWriteSyncer(cfg *config.Logger, signalCh chan<- os.Signal) *BufferedWriteSyncer {

	flusher := &RedisFlusher{
		client: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
			DB:       cfg.Database,
			Password: string(cfg.Password),
		}),
		key: cfg.Key,
	}

	opts := []WriteSyncerOption{
		WithWriteSyncerSignal(signalCh),
	}

	if cfg.DisallowDropLog {
		opts = append(opts, WithNoDrop())
	}

	return NewBufferedWriteSyncer(cfg, flusher, opts...)
}

// LogstashEncoder is an extension of the JSONEncoder that will, combined with a
// redisCore, format the resulting log entry to the Logstash format.
//
// {"level":"...", "@timestamp":"...","@message":"...","@fields":{}}
type logstashEncoder struct {
	zapcore.Encoder
}

func NewLogstashEncoder(cfg zapcore.EncoderConfig) zapcore.Encoder {

	// Override the configuration to use the logstash JSON keys.
	cfg.TimeKey = "@timestamp"
	cfg.MessageKey = "@message"

	// Disable the caller so we can include it into the @fields namespace.
	cfg.CallerKey = ""

	return &logstashEncoder{
		Encoder: zapcore.NewJSONEncoder(cfg),
	}
}

func (e *logstashEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {

	fields = append(fields, zap.String("caller", ent.Caller.String()))

	return e.Encoder.EncodeEntry(ent, fields)
}

type redisCore struct {
	zapcore.Core
	w *BufferedWriteSyncer
}

func NewRedisCore(cfg zapcore.EncoderConfig, logger config.Logger, level zapcore.LevelEnabler, signal chan<- os.Signal) *redisCore {
	w := NewRedisWriteSyncer(&logger, signal)

	return &redisCore{
		Core: zapcore.NewCore(NewLogstashEncoder(getEncoderConfig()), w, level),
		w:    w,
	}
}

func (c *redisCore) With(fields []zapcore.Field) zapcore.Core {

	logstashFields := []zapcore.Field{
		zap.Int("@version", 1),
		zap.Namespace("@fields"),
	}

	return c.Core.With(append(logstashFields, fields...))
}

func (c *redisCore) Stop() error {
	return c.w.Stop()
}
