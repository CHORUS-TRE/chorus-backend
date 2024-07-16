package converter

import (
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToProtoTimestamp(t time.Time) (*timestamp.Timestamp, error) {
	if t.IsZero() {
		return nil, nil
	}

	return timestamppb.New(t), nil
}

func FromProtoBoolValue(b *wrappers.BoolValue) *bool {
	if b == nil {
		return nil
	}
	return &b.Value
}

func PointerToProtoTimestamp(t *time.Time) (*timestamp.Timestamp, error) {
	if t == nil || t.IsZero() {
		return nil, nil
	}

	return timestamppb.New(*t), nil
}

func FromProtoTimestamp(t *timestamp.Timestamp) (time.Time, error) {
	if t == nil {
		return time.Time{}, nil
	}

	return t.AsTime(), nil
}

func FromProtoStringDate(t string) (time.Time, error) {
	if t == "" {
		return time.Time{}, nil
	}

	date, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "unable to convert proto string date to time.Time")
	}
	return date, nil
}

func ToProtoStringDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func SortFromBusiness(bSort model.Sort) *chorus.Sort {
	return &chorus.Sort{
		Order: bSort.SortOrder,
		Type:  bSort.SortType,
	}
}

func SortToBusiness(aSort *chorus.Sort) model.Sort {
	if aSort == nil {
		return model.Sort{}
	}
	return model.Sort{
		SortOrder: aSort.Order,
		SortType:  aSort.Type,
	}
}

func PaginationFromBusiness(bPagination model.Pagination) *chorus.PaginationQuery {
	return &chorus.PaginationQuery{
		Offset: uint32(bPagination.Offset),
		Limit:  uint32(bPagination.Limit),
		Sort:   SortFromBusiness(bPagination.Sort),
	}
}

func PaginationToBusiness(aPagination *chorus.PaginationQuery) model.Pagination {
	if aPagination == nil {
		return model.Pagination{}
	}
	return model.Pagination{
		Offset: uint64(aPagination.Offset),
		Limit:  uint64(aPagination.Limit),
		Sort:   SortToBusiness(aPagination.Sort),
	}
}
