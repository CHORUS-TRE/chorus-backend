package pagination

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
)

type cursorData struct {
	A int
	B string
}

func TestRequestCursor_RequestCursorFromPb(t *testing.T) {
	sampleCursorData := &cursorData{
		A: 123,
		B: "fubar",
	}
	marshaled, err := json.Marshal(sampleCursorData)
	require.NoError(t, err)
	pbCurrentPage := base64.StdEncoding.EncodeToString(marshaled)

	tests := []struct {
		name                string
		pbCursor            *chorus.RequestCursor
		expectedCursorData  *cursorData
		expectedPageSize    uint64
		expectedPageRequest PageRequest
		expectsError        bool
	}{
		{
			name:                "With nil protobuf cursor, returns nil cursor data",
			pbCursor:            nil,
			expectedCursorData:  nil,
			expectedPageSize:    0,
			expectedPageRequest: PageFirst,
			expectsError:        false,
		},
		{
			name:                "With empty protobuf cursor, returns a nil cursor data",
			pbCursor:            &chorus.RequestCursor{},
			expectedCursorData:  nil,
			expectedPageSize:    0,
			expectedPageRequest: PageFirst,
			expectsError:        false,
		},
		{
			name: "With valid protobuf cursor, returns the proper cursor",
			pbCursor: &chorus.RequestCursor{
				CurrentPage: pbCurrentPage,
				PageRequest: "LAST",
				PageSize:    10,
			},
			expectedCursorData:  sampleCursorData,
			expectedPageSize:    10,
			expectedPageRequest: PageLast,
			expectsError:        false,
		},
		{
			name: "With invalid protobuf currentPage, returns an error",
			pbCursor: &chorus.RequestCursor{
				CurrentPage: "invalid_serialized_cursor_data",
				PageRequest: "LAST",
				PageSize:    10,
			},
			expectsError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cursor, err := RequestCursorFromPb[cursorData](test.pbCursor)
			if test.expectsError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				if test.expectedCursorData != nil {
					require.Equal(t, *test.expectedCursorData, *cursor.CursorData)
				} else {
					require.Nil(t, cursor.CursorData)
				}
				require.Equal(t, test.expectedPageSize, cursor.PageSize)
				require.Equal(t, test.expectedPageRequest, cursor.PageRequest)
			}
		})
	}
}

func TestResponseCursor_ToPb(t *testing.T) {
	sampleCursorData := &cursorData{
		A: 123,
		B: "fubar",
	}
	marshaled, err := json.Marshal(sampleCursorData)
	require.NoError(t, err)
	pbCurrentPage := base64.StdEncoding.EncodeToString(marshaled)

	tests := []struct {
		name                string
		responseCursor      *ResponseCursor[cursorData]
		expectedCurrentPage string
		expectedHasPrevious bool
		expectedHasNext     bool
		expectsError        bool
	}{
		{
			name:                "With nil response cursor, returns a default protobuf cursor",
			responseCursor:      nil,
			expectedCurrentPage: "",
			expectedHasPrevious: false,
			expectedHasNext:     false,
			expectsError:        false,
		},
		{
			name:                "With empty response cursor, returns a default protobuf cursor", // Typically when the API returns no data
			responseCursor:      &ResponseCursor[cursorData]{},
			expectedCurrentPage: "",
			expectedHasPrevious: false,
			expectedHasNext:     false,
			expectsError:        false,
		},
		{
			name: "With valid response cursor, returns the proper protobuf cursor",
			responseCursor: &ResponseCursor[cursorData]{
				CursorData:  sampleCursorData,
				HasNext:     true,
				HasPrevious: true,
			},
			expectedCurrentPage: pbCurrentPage,
			expectedHasPrevious: true,
			expectedHasNext:     true,
			expectsError:        false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pbCursor, err := test.responseCursor.ToPb()
			if test.expectsError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				require.Equal(t, test.expectedCurrentPage, pbCursor.CurrentPage)
				require.Equal(t, test.expectedHasPrevious, pbCursor.HasPrevious)
				require.Equal(t, test.expectedHasNext, pbCursor.HasNext)
			}
		})
	}
}
