package logger

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger/mocks"
)

func TestBufferedWriteSyncer_Write(t *testing.T) {
	closed := make(chan struct{})
	close(closed)

	vectors := []struct {
		Data          string
		MaxSize       int64
		Ch            chan []byte
		Before        func(*BufferedWriteSyncer)
		AssertN       require.ValueAssertionFunc
		AssertErr     require.ErrorAssertionFunc
		AssertStopErr require.ErrorAssertionFunc
	}{
		{
			Data:          "simple log with an empty buffer",
			Ch:            make(chan []byte, 1),
			AssertN:       require.NotZero,
			AssertErr:     require.NoError,
			AssertStopErr: require.NoError,
		},
		{
			Data:          "simple log with a blocking queue",
			Ch:            make(chan []byte),
			AssertN:       require.NotZero,
			AssertErr:     require.NoError,
			AssertStopErr: require.NoError,
		},
		{
			Data: "already stopped",
			Ch:   make(chan []byte),
			Before: func(w *BufferedWriteSyncer) {
				//nolint:errcheck
				w.Stop()
			},
			AssertN:       require.Zero,
			AssertErr:     require.Error,
			AssertStopErr: require.NoError,
		},
		{
			Data:          "writer overflow",
			MaxSize:       defaultPoolBufferSize,
			Ch:            make(chan []byte, 1),
			AssertN:       require.Zero,
			AssertErr:     require.Error,
			AssertStopErr: require.NoError,
		},
	}

	for _, v := range vectors {

		t.Run(v.Data, func(t *testing.T) {
			flusher := &mocks.Flusher{}

			w := NewBufferedWriteSyncer(&config.Logger{BufferSize: int(v.MaxSize)}, flusher)
			w.bufferedLogEntries = v.Ch
			w.fallback = io.Discard

			if v.Before != nil {
				v.Before(w)
			}
			//nolint:errcheck
			defer w.Stop()

			flusher.On("Flush", [][]byte{[]byte(v.Data)}).Return(nil).Once()

			n, err := w.Write([]byte(v.Data))
			v.AssertN(t, n)
			v.AssertErr(t, err)

			v.AssertStopErr(t, w.Stop())

			require.EqualValues(t, 0, w.size)
		})
	}
}

func TestBufferedWriteSyncer_Write_OverflowAndSignal(t *testing.T) {

	flusher := &mocks.Flusher{}

	signalCh := make(chan os.Signal, 1)
	buf := new(bytes.Buffer)

	w := NewBufferedWriteSyncer(&config.Logger{}, flusher, WithWriteSyncerSignal(signalCh), WithNoDrop(), WithTicker(time.Hour))
	w.maxSize = 256 * 10
	w.fallback = buf

	//nolint:errcheck
	defer w.Stop()

	flusher.On("Flush", mock.Anything).Return(errors.New("fake"))

	for i := 0; i < 20; i++ {
		//nolint:errcheck
		w.Write([]byte("test\n"))
	}

	signal := <-signalCh
	require.Equal(t, os.Interrupt, signal)

	// Make sure that the logs are written into the fallback as the sink is
	// always failing.
	for i := 0; i < 20; i++ {
		_, err := buf.ReadString('\n')
		require.NoError(t, err)
	}

	require.Equal(t, 0, buf.Len())
}

func TestBufferWriteSyncer_Write_temporarySinkError(t *testing.T) {
	flusher := &mocks.Flusher{}

	w := NewBufferedWriteSyncer(&config.Logger{}, flusher, WithTicker(time.Hour))

	//nolint:errcheck
	defer w.Stop()

	flusher.On("Flush", mock.Anything).Return(errors.New("fake")).Once()
	flusher.On("Flush", mock.Anything).Return(nil)

	_, err := w.Write([]byte("let's if it goes through"))
	require.NoError(t, err)

	err = w.sync(true)
	require.NoError(t, err)
}
