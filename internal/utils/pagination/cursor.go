package pagination

import (
	"encoding/base64"
	"encoding/json"

	"github.com/pkg/errors"

	pb "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
)

type PageRequest string

// Note that these are taken from the cursor proto definition.
// If you change the definitions, you must change these accordingly.
// We are not taking the String() values because it would not make
// these constant.
const (
	PageFirst    PageRequest = "FIRST"
	PagePrevious PageRequest = "PREVIOUS"
	PageNext     PageRequest = "NEXT"
	PageLast     PageRequest = "LAST"
)

type SortOrder string

// Same comment as above holds here as well
const (
	ASC  SortOrder = "ASC"
	DESC SortOrder = "DESC"
)

func (so *SortOrder) Inverse() SortOrder {
	if *so == ASC {
		return DESC
	} else {
		return ASC
	}
}

type RequestCursor[T any] struct {
	// The page to request, w.r.t the current page
	PageRequest PageRequest `validate:"omitempty,oneof=FIRST PREVIOUS NEXT LAST"`
	// The size (i.e., maximum number of items) of the page to request. This needs
	// to be properly validated by the inner service.
	PageSize uint64 `validate:"required,min=1"`
	// Service-specific (inner) cursor data
	// Will be nil if there is no cursor to pass
	CursorData *T `validate:"cursorData"`
}

func (src *RequestCursor[T]) HasCurrentPage() bool {
	return src.CursorData != nil
}

// NewRequestCursor returns a request cursor of given page size for the first page
func NewRequestCursor[T any](pageSize uint64) *RequestCursor[T] {
	return &RequestCursor[T]{
		PageRequest: PageFirst,
		PageSize:    pageSize,
	}
}

// NextRequestCursor returns a request cursor of given page size for the next page
func NextRequestCursor[T any](pageSize uint64, cursor *ResponseCursor[T]) *RequestCursor[T] {
	return &RequestCursor[T]{
		PageRequest: PageNext,
		PageSize:    pageSize,
		CursorData:  cursor.CursorData,
	}
}

func RequestCursorFromPb[T any](rcPb *pb.RequestCursor) (*RequestCursor[T], error) {
	if rcPb == nil {
		rcPb = &pb.RequestCursor{}
	}

	pageRequest := PageRequest(rcPb.PageRequest)
	if pageRequest == "" {
		// FIRST is default (value is 0) in the Cursor proto
		pageRequest = PageFirst
	}

	var cursorData *T
	if rcPb.CurrentPage != "" {
		cursorData = new(T)
		currentPage, err := base64.StdEncoding.DecodeString(rcPb.CurrentPage)
		if err != nil {
			return nil, errors.Wrap(err, "not a valid base64-encoded string")
		}
		if err := json.Unmarshal(currentPage, cursorData); err != nil {
			return nil, errors.Wrap(err, "failed to deserialize data into the target struct")
		}
	}

	serviceCursor := RequestCursor[T]{
		PageRequest: pageRequest,
		PageSize:    rcPb.PageSize,
		CursorData:  cursorData,
	}

	return &serviceCursor, nil
}

type ResponseCursor[T any] struct {
	CursorData  *T
	HasPrevious bool
	HasNext     bool
}

func (cursor *ResponseCursor[T]) ToPb() (*pb.ResponseCursor, error) {
	responseCursor := &pb.ResponseCursor{}

	if cursor == nil {
		cursor = &ResponseCursor[T]{
			CursorData:  nil,
			HasPrevious: false,
			HasNext:     false,
		}
	}

	currentPage := ""
	if cursor.CursorData != nil {
		currentPageBytes, err := json.Marshal(cursor.CursorData)
		if err != nil {
			return responseCursor, errors.Wrap(err, "failed to marshal cursor data as JSON")
		}

		currentPage = base64.StdEncoding.EncodeToString(currentPageBytes)
	}

	return &pb.ResponseCursor{
		CurrentPage: currentPage,
		HasPrevious: cursor.HasPrevious,
		HasNext:     cursor.HasNext,
	}, nil
}

type Sorting struct {
	// Handling service should define this field internally
	// with proper column value validation.
	// sortBy    []string
	SortOrder SortOrder `validate:"omitempty,oneof=ASC DESC"`
}
