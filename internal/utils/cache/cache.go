package cache

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"sort"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/trace"

	"github.com/coocood/freecache"
	"go.uber.org/zap"
)

// Manual function

func Decode(v []byte, t interface{}) error {
	dec := gob.NewDecoder(bytes.NewBuffer(v))
	return dec.Decode(t)
}

func Encode(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(v)
	return buf.Bytes(), err
}

func GenKey(args ...interface{}) ([]byte, error) {
	key, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(key)
	return hash[:], nil
}

// Middleware functions

// Key is the option type to generate a unique key for the cache.
type Key func(w io.Writer)

// WithString is an option to write a string to the key entropy.
func WithString(v string) Key {
	return func(w io.Writer) {
		_, _ = w.Write([]byte(v))
	}
}

// WithStrings is an option to write a list of strings sorted lexicographically
// to the key entropy.
func WithStrings(values []string) Key {
	return func(w io.Writer) {
		c := make([]string, len(values))
		copy(c, values)
		sort.Strings(c)
		j, _ := json.Marshal(c)
		_, _ = w.Write(j)
	}
}

// WithInt is an option to write a integer to the key entropy.
func WithInt(v int) Key {
	return func(w io.Writer) {
		data := make([]byte, binary.MaxVarintLen64)
		n := binary.PutVarint(data, int64(v))
		_, _ = w.Write(data[:n])
	}
}

func WithInt64(v int64) Key {
	return func(w io.Writer) {
		data := make([]byte, binary.MaxVarintLen64)
		n := binary.PutVarint(data, v)
		_, _ = w.Write(data[:n])
	}
}

func WithUint64(v uint64) Key {
	return func(w io.Writer) {
		data := make([]byte, binary.MaxVarintLen64)
		n := binary.PutUvarint(data, v)
		_, _ = w.Write(data[:n])
	}
}

func WithInterface(obj interface{}) Key {
	return func(w io.Writer) {
		enc := gob.NewEncoder(w)
		_ = enc.Encode(obj)
	}
}

type Cache struct {
	fc  *freecache.Cache
	log *logger.ContextLogger
}

func NewCache(fc *freecache.Cache, log *logger.ContextLogger) *Cache {
	if fc == nil {
		panic("cache cannot be nil")
	}
	if log == nil {
		panic("log cannot be nil")
	}

	return &Cache{
		fc:  fc,
		log: log,
	}
}

func (c Cache) NewEntry(keys ...Key) CacheEntry {
	caller := trace.Caller()

	h := sha256.New()
	_, _ = h.Write([]byte(caller))
	for _, key := range keys {
		key(h)
	}

	return CacheEntry{
		fc:     c.fc,
		log:    c.log,
		key:    h.Sum([]byte{0}),
		caller: caller,
	}
}

type CacheEntry struct {
	fc     *freecache.Cache
	log    *logger.ContextLogger
	key    []byte
	caller string
}

func (e CacheEntry) Get(ctx context.Context, dest ...interface{}) bool {

	for i, obj := range dest {
		e.key[0] = byte(i)

		value, err := e.fc.Get(e.key)
		if errors.Is(err, freecache.ErrNotFound) {
			e.log.Debug(ctx, "cache miss", zap.String(logger.LoggerKeyParentCaller, e.caller))
			return false
		}
		if err != nil {
			e.log.Error(ctx, "cache: unable to read", zap.Error(err), zap.String(logger.LoggerKeyParentCaller, e.caller))
			return false
		}

		dec := gob.NewDecoder(bytes.NewBuffer(value))

		err = dec.Decode(obj)
		if err != nil {
			e.log.Error(ctx, "cache: unable to decode", zap.Error(err), zap.String(logger.LoggerKeyParentCaller, e.caller))
			return false
		}
	}

	e.log.Debug(ctx, "cache hit", zap.String(logger.LoggerKeyParentCaller, e.caller))
	return true
}

// Set writes the values to the cache. When expiration is above zero, it
// indicates the expiration time in seconds for the values.
func (e CacheEntry) Set(ctx context.Context, expiration int, values ...interface{}) {
	for i, obj := range values {
		// Set the destination index so that each of them has a different
		// key.
		e.key[0] = byte(i)

		data := new(bytes.Buffer)
		enc := gob.NewEncoder(data)

		err := enc.Encode(obj)
		if err != nil {
			e.log.Error(ctx, "cache: unable to encode", zap.Error(err), zap.Stringer(logger.LoggerKeyObjectType, reflect.TypeOf(obj)))
			return
		}

		err = e.fc.Set(e.key, data.Bytes(), expiration)
		if err != nil {
			if errors.Is(err, freecache.ErrLargeEntry) {
				e.log.Warn(ctx, "cache: unable to set value", zap.Error(err), zap.Stringer(logger.LoggerKeyObjectType, reflect.TypeOf(obj)))
				return
			}
			e.log.Error(ctx, "cache: unable to set value", zap.Error(err), zap.Stringer(logger.LoggerKeyObjectType, reflect.TypeOf(obj)))
			return
		}
	}
}
