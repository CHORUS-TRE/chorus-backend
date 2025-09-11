package storage

import (
	"fmt"
	"strings"

	common "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

func Rollback(tx *sqlx.Tx, txErr error) error {
	rollErr := tx.Rollback()
	if rollErr != nil {
		return fmt.Errorf("%s: %w", txErr.Error(), rollErr)
	}
	return txErr
}

func PqInt64ToUint64(array pq.Int64Array) []uint64 {
	output := make([]uint64, len(array))
	for i, element := range array {
		output[i] = uint64(element)
	}
	return output
}

func Uint64ToPqInt64(array []uint64) pq.Int64Array {
	output := make(pq.Int64Array, len(array))
	for i, element := range array {
		output[i] = int64(element)
	}
	return output
}

func SortOrderToString(sortOrder string) string {
	order := strings.ToUpper(sortOrder)
	if order != "DESC" && order != "ASC" {
		return "ASC"
	}
	return order
}

func SortTypeToString(sortType string, model common.Sortable) string {
	if ok := model.IsValidSortType(sortType); !ok {
		return ""
	}

	return sortType
}

// BuildPaginationClause returns the SQL clause string for ORDER BY, LIMIT and OFFSET and the effective pagination object
func BuildPaginationClause(pagination *common.Pagination, model common.Sortable) (string, *common.Pagination) {
	var clause string

	// Check if pagination is empty
	if pagination.Limit == 0 && pagination.Offset == 0 && pagination.Sort.SortType == "" {
		return "", nil
	}

	// Add ORDER BY clause
	sortType := SortTypeToString(pagination.Sort.SortType, model)
	sortOrder := ""
	if sortType != "" {
		sortOrder = SortOrderToString(pagination.Sort.SortOrder)
		clause = fmt.Sprintf(" ORDER BY %s %s", sortType, sortOrder)
	}

	// Add LIMIT clause
	limit := pagination.Limit
	if pagination.Limit <= 0 || pagination.Limit > common.MAX_LIMIT {
		limit = common.DEFAULT_LIMIT
	}
	clause += fmt.Sprintf(" LIMIT %d", limit)

	// Add OFFSET clause
	offset := uint64(0)
	if pagination.Offset > 0 {
		offset = pagination.Offset
		clause = fmt.Sprintf("%s OFFSET %d", clause, pagination.Offset)
	}

	return clause, &common.Pagination{
		Offset: offset,
		Limit:  limit,
		Sort: common.Sort{
			SortType:  sortType,
			SortOrder: sortOrder,
		},
	}
}
