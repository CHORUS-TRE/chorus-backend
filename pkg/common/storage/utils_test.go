//go:build unit

package storage

import (
	"testing"
	"time"

	audit_model "github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	"github.com/stretchr/testify/assert"
)

func ptr[T any](v T) *T {
	return &v
}

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
				TenantID:     ptr(uint64(123)),
				UserID:       ptr(uint64(0)),
				ResourceID:   ptr(uint64(0)),
				WorkspaceID:  ptr(uint64(0)),
				WorkbenchID:  ptr(uint64(0)),
				ResourceType: nil,
				Action:       nil,
				FromTime:     nil,
				ToTime:       nil,
			},
			initialArgs:    []interface{}{},
			expectedClause: "tenantid = $1",
			expectedArgs:   []interface{}{ptr(uint64(123))},
		},
		{
			name: "filter with only UserID",
			filter: &audit_model.AuditFilter{
				TenantID:     ptr(uint64(0)),
				UserID:       ptr(uint64(456)),
				ResourceID:   ptr(uint64(0)),
				WorkspaceID:  ptr(uint64(0)),
				WorkbenchID:  ptr(uint64(0)),
				ResourceType: nil,
				Action:       nil,
				FromTime:     nil,
				ToTime:       nil,
			},
			initialArgs:    []interface{}{},
			expectedClause: "userid = $1",
			expectedArgs:   []interface{}{ptr(uint64(456))},
		},
		{
			name: "filter with only ResourceType",
			filter: &audit_model.AuditFilter{
				TenantID:     ptr(uint64(0)),
				UserID:       ptr(uint64(0)),
				ResourceType: ptr(audit_model.AuditResourceUser),
				ResourceID:   ptr(uint64(0)),
				WorkspaceID:  ptr(uint64(0)),
				WorkbenchID:  ptr(uint64(0)),
				Action:       nil,
				FromTime:     nil,
				ToTime:       nil,
			},
			initialArgs:    []interface{}{},
			expectedClause: "resourcetype = $1",
			expectedArgs:   []interface{}{ptr(audit_model.AuditResourceUser)},
		},
		{
			name: "filter with only ResourceID",
			filter: &audit_model.AuditFilter{
				TenantID:     ptr(uint64(0)),
				UserID:       ptr(uint64(0)),
				ResourceType: nil,
				ResourceID:   ptr(uint64(789)),
				WorkspaceID:  ptr(uint64(0)),
				WorkbenchID:  ptr(uint64(0)),
				Action:       nil,
				FromTime:     nil,
				ToTime:       nil,
			},
			initialArgs:    []interface{}{},
			expectedClause: "resourceid = $1",
			expectedArgs:   []interface{}{ptr(uint64(789))},
		},
		{
			name: "filter with only Action",
			filter: &audit_model.AuditFilter{
				TenantID:     ptr(uint64(0)),
				UserID:       ptr(uint64(0)),
				ResourceType: nil,
				ResourceID:   ptr(uint64(0)),
				Action:       ptr(audit_model.AuditActionUserLogin),
				WorkspaceID:  ptr(uint64(0)),
				WorkbenchID:  ptr(uint64(0)),
				FromTime:     nil,
				ToTime:       nil,
			},
			initialArgs:    []interface{}{},
			expectedClause: "action = $1",
			expectedArgs:   []interface{}{ptr(audit_model.AuditActionUserLogin)},
		},
		{
			name: "filter with only WorkspaceID",
			filter: &audit_model.AuditFilter{
				TenantID:     ptr(uint64(0)),
				UserID:       ptr(uint64(0)),
				ResourceType: nil,
				ResourceID:   ptr(uint64(0)),
				Action:       nil,
				WorkspaceID:  ptr(uint64(100)),
				WorkbenchID:  ptr(uint64(0)),
				FromTime:     nil,
				ToTime:       nil,
			},
			initialArgs:    []interface{}{},
			expectedClause: "workspaceid = $1",
			expectedArgs:   []interface{}{ptr(uint64(100))},
		},
		{
			name: "filter with only WorkbenchID",
			filter: &audit_model.AuditFilter{
				TenantID:     ptr(uint64(0)),
				UserID:       ptr(uint64(0)),
				ResourceType: nil,
				ResourceID:   ptr(uint64(0)),
				Action:       nil,
				WorkspaceID:  ptr(uint64(0)),
				WorkbenchID:  ptr(uint64(200)),
				FromTime:     nil,
				ToTime:       nil,
			},
			initialArgs:    []interface{}{},
			expectedClause: "workbenchid = $1",
			expectedArgs:   []interface{}{ptr(uint64(200))},
		},
		{
			name: "filter with only FromTime",
			filter: &audit_model.AuditFilter{
				TenantID:     ptr(uint64(0)),
				UserID:       ptr(uint64(0)),
				ResourceType: nil,
				ResourceID:   ptr(uint64(0)),
				Action:       nil,
				WorkspaceID:  ptr(uint64(0)),
				WorkbenchID:  ptr(uint64(0)),
				FromTime:     &fromTime,
				ToTime:       nil,
			},
			initialArgs:    []interface{}{},
			expectedClause: "createdat >= $1",
			expectedArgs:   []interface{}{&fromTime},
		},
		{
			name: "filter with only ToTime",
			filter: &audit_model.AuditFilter{
				TenantID:     ptr(uint64(0)),
				UserID:       ptr(uint64(0)),
				ResourceType: nil,
				ResourceID:   ptr(uint64(0)),
				Action:       nil,
				WorkspaceID:  ptr(uint64(0)),
				WorkbenchID:  ptr(uint64(0)),
				FromTime:     nil,
				ToTime:       &toTime,
			},
			initialArgs:    []interface{}{},
			expectedClause: "createdat <= $1",
			expectedArgs:   []interface{}{&toTime},
		},
		{
			name: "filter with TenantID and UserID",
			filter: &audit_model.AuditFilter{
				TenantID:     ptr(uint64(123)),
				UserID:       ptr(uint64(456)),
				ResourceType: nil,
				ResourceID:   ptr(uint64(0)),
				Action:       nil,
				WorkspaceID:  ptr(uint64(0)),
				WorkbenchID:  ptr(uint64(0)),
				FromTime:     nil,
				ToTime:       nil,
			},
			initialArgs:    []interface{}{},
			expectedClause: "tenantid = $1 AND userid = $2",
			expectedArgs:   []interface{}{ptr(uint64(123)), ptr(uint64(456))},
		},
		{
			name: "filter with time range",
			filter: &audit_model.AuditFilter{
				TenantID:     ptr(uint64(0)),
				UserID:       ptr(uint64(0)),
				ResourceType: nil,
				ResourceID:   ptr(uint64(0)),
				Action:       nil,
				WorkspaceID:  ptr(uint64(0)),
				WorkbenchID:  ptr(uint64(0)),
				FromTime:     &fromTime,
				ToTime:       &toTime,
			},
			initialArgs:    []interface{}{},
			expectedClause: "createdat >= $1 AND createdat <= $2",
			expectedArgs:   []interface{}{&fromTime, &toTime},
		},
		{
			name: "filter with multiple fields",
			filter: &audit_model.AuditFilter{
				TenantID:     ptr(uint64(123)),
				UserID:       ptr(uint64(456)),
				ResourceType: ptr(audit_model.AuditResourceWorkspace),
				ResourceID:   ptr(uint64(789)),
				Action:       ptr(audit_model.AuditActionWorkspaceCreate),
				WorkspaceID:  ptr(uint64(0)),
				WorkbenchID:  ptr(uint64(0)),
				FromTime:     nil,
				ToTime:       nil,
			},
			initialArgs:    []interface{}{},
			expectedClause: "tenantid = $1 AND userid = $2 AND resourcetype = $3 AND resourceid = $4 AND action = $5",
			expectedArgs: []interface{}{
				ptr(uint64(123)),
				ptr(uint64(456)),
				ptr(audit_model.AuditResourceWorkspace),
				ptr(uint64(789)),
				ptr(audit_model.AuditActionWorkspaceCreate),
			},
		},
		{
			name: "filter with all fields set",
			filter: &audit_model.AuditFilter{
				TenantID:     ptr(uint64(1)),
				UserID:       ptr(uint64(2)),
				ResourceType: ptr(audit_model.AuditResourceWorkbench),
				ResourceID:   ptr(uint64(3)),
				Action:       ptr(audit_model.AuditActionWorkbenchCreate),
				WorkspaceID:  ptr(uint64(4)),
				WorkbenchID:  ptr(uint64(5)),
				FromTime:     &fromTime,
				ToTime:       &toTime,
			},
			initialArgs:    []interface{}{},
			expectedClause: "tenantid = $1 AND userid = $2 AND resourcetype = $3 AND resourceid = $4 AND action = $5 AND workspaceid = $6 AND workbenchid = $7 AND createdat >= $8 AND createdat <= $9",
			expectedArgs: []interface{}{
				ptr(uint64(1)),
				ptr(uint64(2)),
				ptr(audit_model.AuditResourceWorkbench),
				ptr(uint64(3)),
				ptr(audit_model.AuditActionWorkbenchCreate),
				ptr(uint64(4)),
				ptr(uint64(5)),
				&fromTime,
				&toTime,
			},
		},
		{
			name: "filter with all zero/nil values returns empty string and no args",
			filter: &audit_model.AuditFilter{
				TenantID:     ptr(uint64(0)),
				UserID:       ptr(uint64(0)),
				ResourceType: nil,
				ResourceID:   ptr(uint64(0)),
				Action:       nil,
				WorkspaceID:  ptr(uint64(0)),
				WorkbenchID:  ptr(uint64(0)),
				FromTime:     nil,
				ToTime:       nil,
			},
			initialArgs:    []interface{}{},
			expectedClause: "",
			expectedArgs:   []interface{}{},
		},
		{
			name: "filter appends to existing args with correct placeholder numbers",
			filter: &audit_model.AuditFilter{
				TenantID:     ptr(uint64(123)),
				UserID:       ptr(uint64(456)),
				ResourceType: nil,
				ResourceID:   ptr(uint64(0)),
				Action:       nil,
				WorkspaceID:  ptr(uint64(0)),
				WorkbenchID:  ptr(uint64(0)),
				FromTime:     nil,
				ToTime:       nil,
			},
			initialArgs:    []interface{}{"existing_arg_1", "existing_arg_2"},
			expectedClause: "tenantid = $3 AND userid = $4",
			expectedArgs:   []interface{}{"existing_arg_1", "existing_arg_2", ptr(uint64(123)), ptr(uint64(456))},
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
