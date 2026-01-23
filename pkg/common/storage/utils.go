package storage

import (
	"fmt"
	"strings"

	audit_model "github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	common "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	user_service "github.com/CHORUS-TRE/chorus-backend/pkg/user/service"
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

func BuildUserFilterClause(filter *user_service.UserFilter, args *[]interface{}) string {
	var clauses []string

	if filter != nil && len(filter.IDsIn) > 0 {
		clauses = append(clauses, fmt.Sprintf("id = ANY($%d)", len(*args)+1))
		*args = append(*args, pq.Int64Array(Uint64ToPqInt64(filter.IDsIn)))
	}

	if filter != nil && len(filter.WorkspaceIDs) > 0 {
		clauses = append(clauses, fmt.Sprintf(`
		id IN (
			SELECT userid FROM user_role WHERE id IN (
				SELECT userroleid FROM user_role_context WHERE contextdimension='workspace' AND value = ANY($%d)
			)
		)`, len(*args)+1))
		*args = append(*args, pq.Int64Array(Uint64ToPqInt64(filter.WorkspaceIDs)))
	}

	if filter != nil && len(filter.WorkbenchIDs) > 0 {
		clauses = append(clauses, fmt.Sprintf(`
		id IN (
			SELECT userid FROM user_role WHERE id IN (
				SELECT userroleid FROM user_role_context WHERE contextdimension='workbench' AND value = ANY($%d)
			)
		)`, len(*args)+1))
		*args = append(*args, pq.Int64Array(Uint64ToPqInt64(filter.WorkbenchIDs)))
	}

	if filter != nil && filter.Search != nil && *filter.Search != "" {
		clauses = append(clauses, fmt.Sprintf("(LOWER(firstname) LIKE LOWER($%d) OR LOWER(lastname) LIKE LOWER($%d) OR LOWER(username) LIKE LOWER($%d))", len(*args)+1, len(*args)+1, len(*args)+1))
		*args = append(*args, "%"+*filter.Search+"%")
	}

	return strings.Join(clauses, " AND ")
}

func BuildAuditFilterClause(filter *audit_model.AuditFilter, args *[]interface{}) string {
	var clauses []string

	if filter != nil && *filter.TenantID != 0 {
		clauses = append(clauses, fmt.Sprintf("tenantid = $%d", len(*args)+1))
		*args = append(*args, filter.TenantID)
	}

	if filter != nil && *filter.UserID != 0 {
		clauses = append(clauses, fmt.Sprintf("userid = $%d", len(*args)+1))
		*args = append(*args, filter.UserID)
	}

	if filter != nil && filter.Action != nil {
		clauses = append(clauses, fmt.Sprintf("action = $%d", len(*args)+1))
		*args = append(*args, filter.Action)
	}

	if filter != nil && *filter.WorkspaceID != 0 {
		clauses = append(clauses, fmt.Sprintf("workspaceid = $%d", len(*args)+1))
		*args = append(*args, filter.WorkspaceID)
	}

	if filter != nil && *filter.WorkbenchID != 0 {
		clauses = append(clauses, fmt.Sprintf("workbenchid = $%d", len(*args)+1))
		*args = append(*args, filter.WorkbenchID)
	}

	if filter != nil && filter.FromTime != nil {
		clauses = append(clauses, fmt.Sprintf("createdat >= $%d", len(*args)+1))
		*args = append(*args, filter.FromTime)
	}

	if filter != nil && filter.ToTime != nil {
		clauses = append(clauses, fmt.Sprintf("createdat <= $%d", len(*args)+1))
		*args = append(*args, filter.ToTime)
	}

	return strings.Join(clauses, " AND ")
}
