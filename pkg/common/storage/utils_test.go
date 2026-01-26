package storage

import (
	"testing"
	"time"

	audit_model "github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	"github.com/stretchr/testify/assert"
)

func TestBuildAuditFilterClause(t *testing.T) {
	fromTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	toTime := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name           string
		filter         *audit_model.AuditFilter
		initialArgs    []interface{}
		expectedClause string
		expectedArgs   []interface{}
	}{
		{
			name:           "nil filter returns empty string and no args",
			filter:         nil,
			initialArgs:    []interface{}{},
			expectedClause: "",
			expectedArgs:   []interface{}{},
		},
		{
			name: "filter with only TenantID",
			filter: &audit_model.AuditFilter{
				TenantID: 123,
			},
			initialArgs:    []interface{}{},
			expectedClause: "tenantid = $1",
			expectedArgs:   []interface{}{uint64(123)},
		},
		{
			name: "filter with only UserID",
			filter: &audit_model.AuditFilter{
				UserID: 456,
			},
			initialArgs:    []interface{}{},
			expectedClause: "userid = $1",
			expectedArgs:   []interface{}{uint64(456)},
		},
		{
			name: "filter with only Action",
			filter: &audit_model.AuditFilter{
				Action: audit_model.AuditActionUserLogin,
			},
			initialArgs:    []interface{}{},
			expectedClause: "action = $1",
			expectedArgs:   []interface{}{audit_model.AuditActionUserLogin},
		},
		{
			name: "filter with only WorkspaceID",
			filter: &audit_model.AuditFilter{
				WorkspaceID: 100,
			},
			initialArgs:    []interface{}{},
			expectedClause: "workspaceid = $1",
			expectedArgs:   []interface{}{uint64(100)},
		},
		{
			name: "filter with only WorkbenchID",
			filter: &audit_model.AuditFilter{
				WorkbenchID: 200,
			},
			initialArgs:    []interface{}{},
			expectedClause: "workbenchid = $1",
			expectedArgs:   []interface{}{uint64(200)},
		},
		{
			name: "filter with only FromTime",
			filter: &audit_model.AuditFilter{
				FromTime: fromTime,
			},
			initialArgs:    []interface{}{},
			expectedClause: "createdat >= $1",
			expectedArgs:   []interface{}{fromTime},
		},
		{
			name: "filter with only ToTime",
			filter: &audit_model.AuditFilter{
				ToTime: toTime,
			},
			initialArgs:    []interface{}{},
			expectedClause: "createdat <= $1",
			expectedArgs:   []interface{}{toTime},
		},
		{
			name: "filter with TenantID and UserID",
			filter: &audit_model.AuditFilter{
				TenantID: 123,
				UserID:   456,
			},
			initialArgs:    []interface{}{},
			expectedClause: "tenantid = $1 AND userid = $2",
			expectedArgs:   []interface{}{uint64(123), uint64(456)},
		},
		{
			name: "filter with time range",
			filter: &audit_model.AuditFilter{
				FromTime: fromTime,
				ToTime:   toTime,
			},
			initialArgs:    []interface{}{},
			expectedClause: "createdat >= $1 AND createdat <= $2",
			expectedArgs:   []interface{}{fromTime, toTime},
		},
		{
			name: "filter with multiple fields",
			filter: &audit_model.AuditFilter{
				TenantID: 123,
				UserID:   456,
				Action:   audit_model.AuditActionWorkspaceCreate,
			},
			initialArgs:    []interface{}{},
			expectedClause: "tenantid = $1 AND userid = $2 AND action = $3",
			expectedArgs: []interface{}{
				uint64(123),
				uint64(456),
				audit_model.AuditActionWorkspaceCreate,
			},
		},
		{
			name: "filter with all fields set",
			filter: &audit_model.AuditFilter{
				TenantID:    1,
				UserID:      2,
				Action:      audit_model.AuditActionWorkbenchCreate,
				WorkspaceID: 4,
				WorkbenchID: 5,
				FromTime:    fromTime,
				ToTime:      toTime,
			},
			initialArgs:    []interface{}{},
			expectedClause: "tenantid = $1 AND userid = $2 AND action = $3 AND workspaceid = $4 AND workbenchid = $5 AND createdat >= $6 AND createdat <= $7",
			expectedArgs: []interface{}{
				uint64(1),
				uint64(2),
				audit_model.AuditActionWorkbenchCreate,
				uint64(4),
				uint64(5),
				fromTime,
				toTime,
			},
		},
		{
			name:           "filter with all zero values returns empty string and no args",
			filter:         &audit_model.AuditFilter{},
			initialArgs:    []interface{}{},
			expectedClause: "",
			expectedArgs:   []interface{}{},
		},
		{
			name: "filter appends to existing args with correct placeholder numbers",
			filter: &audit_model.AuditFilter{
				TenantID: 123,
				UserID:   456,
			},
			initialArgs:    []interface{}{"existing_arg_1", "existing_arg_2"},
			expectedClause: "tenantid = $3 AND userid = $4",
			expectedArgs:   []interface{}{"existing_arg_1", "existing_arg_2", uint64(123), uint64(456)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := make([]interface{}, len(tt.initialArgs))
			copy(args, tt.initialArgs)

			result := BuildAuditFilterClause(tt.filter, &args)

			assert.Equal(t, tt.expectedClause, result)
			assert.Equal(t, len(tt.expectedArgs), len(args), "args length mismatch")

			for i, expectedArg := range tt.expectedArgs {
				assert.Equal(t, expectedArg, args[i], "arg at index %d mismatch", i)
			}
		})
	}
}
