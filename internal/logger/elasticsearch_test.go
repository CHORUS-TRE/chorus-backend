package logger

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger/mocks"
)

func TestOpenSearchBatcher_Flush(t *testing.T) {
	tests := []struct {
		Name         string
		Entries      [][]byte
		FlusherCalls func(flusher *mocks.Flusher)
	}{
		{
			Name:         "empty",
			Entries:      nil,
			FlusherCalls: func(flusher *mocks.Flusher) {},
		},
		{
			Name:    "one part",
			Entries: [][]byte{{1}},
			FlusherCalls: func(flusher *mocks.Flusher) {
				flusher.On("Flush", [][]byte{{1}}).Return(nil).Once()
			},
		},
		{
			Name:    "two parts",
			Entries: [][]byte{{1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}},
			FlusherCalls: func(flusher *mocks.Flusher) {
				flusher.On("Flush", [][]byte{{1}, {2}, {3}, {4}}).Return(nil).Once()
				flusher.On("Flush", [][]byte{{5}, {6}, {7}, {8}}).Return(nil).Once()
			},
		},
		{
			Name:    "three parts",
			Entries: [][]byte{{1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}},
			FlusherCalls: func(flusher *mocks.Flusher) {
				flusher.On("Flush", [][]byte{{1}, {2}, {3}, {4}}).Return(nil).Once()
				flusher.On("Flush", [][]byte{{5}, {6}, {7}, {8}}).Return(nil).Once()
				flusher.On("Flush", [][]byte{{9}}).Return(nil).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			f, err := NewOpenSearchFlusher(&config.Logger{})
			require.NoError(t, err)

			flusher := mocks.NewFlusher(t)

			f.Flusher = flusher
			f.maxSize = 4 // Bytes

			tt.FlusherCalls(flusher)

			err = f.Flush(tt.Entries)
			require.NoError(t, err)
		})
	}
}

func TestEsBulkBody_Read(t *testing.T) {

	const expect = `{"index": {"_index": "index", "_id": "random"} }
line 1
{"index": {"_index": "index", "_id": "random"} }
line 2
{"index": {"_index": "index", "_id": "random"} }
line 3
`

	vectors := []struct {
		TestName string
		Dest     []byte
	}{
		{
			TestName: "Read (1)",
			Dest:     make([]byte, 1),
		},
		{
			TestName: "Read (2)",
			Dest:     make([]byte, 2),
		},
		{
			TestName: "Read (3)",
			Dest:     make([]byte, 3),
		},
		{
			TestName: "Read (1024)",
			Dest:     make([]byte, 1024),
		},
	}

	for _, v := range vectors {

		t.Run(v.TestName, func(t *testing.T) {

			entries := [][]byte{
				[]byte("line 1"),
				[]byte("line 2"),
				[]byte("line 3"),
			}

			entriesCopy := [][]byte{
				[]byte("line 1"),
				[]byte("line 2"),
				[]byte("line 3"),
			}

			b := newBulkBody(entries, "index")
			b.randomFn = func() string { return "random" }

			res := &strings.Builder{}

			var err error
			var n int
			for ; err != io.EOF; n, err = b.Read(v.Dest) {
				_, e := res.Write(v.Dest[:n])
				require.NoError(t, e)
			}

			require.Equal(t, entriesCopy, entries) // expect no alteration.
			require.Equal(t, expect, res.String())
		})
	}
}

func TestPrepareIndex(t *testing.T) {

	tests := []struct {
		Index  string
		Expect string
	}{
		{
			Index:  "simple-index",
			Expect: "simple-index",
		},
		{
			Index:  "simple-index-%Y",
			Expect: "simple-index-2022",
		},
		{
			Index:  "simple-index-%y",
			Expect: "simple-index-22",
		},
		{
			Index:  "simple-index-%m",
			Expect: "simple-index-06",
		},
		{
			Index:  "simple-index-%d",
			Expect: "simple-index-04",
		},
		{
			Index:  "simple-index-%H.%M.%S-suffix",
			Expect: "simple-index-12.30.55-suffix",
		},
		{
			Index:  "simple-index-%%",
			Expect: "simple-index-%%",
		},
		{
			Index:  "simple-index-%YY",
			Expect: "simple-index-2022Y",
		},
	}

	for _, tt := range tests {

		currentTime := time.Date(2022, 6, 4, 12, 30, 55, 0, time.UTC)

		t.Run(tt.Index, func(t *testing.T) {
			require.Equal(t, tt.Expect, prepareIndex(tt.Index, currentTime))
		})
	}
}
