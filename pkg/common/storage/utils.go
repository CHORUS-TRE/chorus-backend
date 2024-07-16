package storage

import (
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

func Rollback(tx *sqlx.Tx, txErr error) error {
	rollErr := tx.Rollback()
	if rollErr != nil {
		return errors.Wrap(rollErr, txErr.Error())
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

func SortOrderToString(sortOrder string) string {
	if sortOrder != "DESC" && sortOrder != "ASC" {
		return "ASC"
	}
	return sortOrder
}
