package audit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"

	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/correlation"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
)

func TestNewEntry(t *testing.T) {
	tests := []struct {
		name           string
		ctx            context.Context
		action         model.AuditAction
		opts           []Option
		expectedTenant uint64
		expectedUser   uint64
		expectedCID    string
		expectedDesc   string
		expectedDetail map[string]any
	}{
		{
			name:           "basic entry with no context",
			ctx:            context.Background(),
			action:         model.AuditActionUserLogin,
			opts:           nil,
			expectedTenant: 0,
			expectedUser:   0,
			expectedCID:    "",
			expectedDesc:   "",
			expectedDetail: map[string]any{},
		},
		{
			name: "entry with JWT claims in context",
			ctx: context.WithValue(context.Background(), jwt_model.JWTClaimsContextKey, &jwt_model.JWTClaims{
				TenantID: 123,
				ID:       456,
				Username: "testuser",
			}),
			action:         model.AuditActionUserCreate,
			opts:           nil,
			expectedTenant: 123,
			expectedUser:   456,
			expectedCID:    "",
			expectedDesc:   "",
			expectedDetail: map[string]any{},
		},
		{
			name: "entry with correlation ID in context",
			ctx: context.WithValue(context.Background(), correlation.CorrelationIDContextKey{},
				"test-correlation-id"),
			action:         model.AuditActionWorkspaceCreate,
			opts:           nil,
			expectedTenant: 0,
			expectedUser:   0,
			expectedCID:    "test-correlation-id",
			expectedDesc:   "",
			expectedDetail: map[string]any{},
		},
		{
			name: "entry with full context and options",
			ctx: func() context.Context {
				ctx := context.WithValue(context.Background(), jwt_model.JWTClaimsContextKey, &jwt_model.JWTClaims{
					TenantID: 999,
					ID:       888,
					Username: "admin",
				})
				return context.WithValue(ctx, correlation.CorrelationIDContextKey{}, "full-cid")
			}(),
			action: model.AuditActionAppInstanceCreate,
			opts: []Option{
				WithDescription("Test description"),
				WithWorkspaceID(111),
				WithWorkbenchID(222),
				WithDetail("app_id", uint64(333)),
				WithDetail("status", "running"),
			},
			expectedTenant: 999,
			expectedUser:   888,
			expectedCID:    "full-cid",
			expectedDesc:   "Test description",
			expectedDetail: map[string]any{
				"app_id": uint64(333),
				"status": "running",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			entry := NewEntry(tt.ctx, tt.action, tt.opts...)
			after := time.Now()

			assert.Equal(t, tt.action, entry.Action)
			assert.Equal(t, tt.expectedTenant, entry.TenantID)
			assert.Equal(t, tt.expectedUser, entry.UserID)
			assert.Equal(t, tt.expectedCID, entry.CorrelationID)
			assert.Equal(t, tt.expectedDesc, entry.Description)
			assert.Equal(t, tt.expectedDetail, map[string]any(entry.Details))
			assert.True(t, entry.CreatedAt.After(before) || entry.CreatedAt.Equal(before))
			assert.True(t, entry.CreatedAt.Before(after) || entry.CreatedAt.Equal(after))
		})
	}
}

func TestWithWorkspaceID(t *testing.T) {
	entry := &model.AuditEntry{}
	opt := WithWorkspaceID(12345)
	opt(entry)

	assert.Equal(t, uint64(12345), entry.WorkspaceID)
}

func TestWithWorkbenchID(t *testing.T) {
	entry := &model.AuditEntry{}
	opt := WithWorkbenchID(67890)
	opt(entry)

	assert.Equal(t, uint64(67890), entry.WorkbenchID)
}

func TestWithDescription(t *testing.T) {
	entry := &model.AuditEntry{}
	description := "User created successfully"
	opt := WithDescription(description)
	opt(entry)

	assert.Equal(t, description, entry.Description)
}

func TestWithDetail(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    any
		expected any
	}{
		{
			name:     "string value",
			key:      "username",
			value:    "testuser",
			expected: "testuser",
		},
		{
			name:     "uint64 value",
			key:      "user_id",
			value:    uint64(123),
			expected: uint64(123),
		},
		{
			name:     "int value",
			key:      "status_code",
			value:    200,
			expected: 200,
		},
		{
			name:     "bool value",
			key:      "success",
			value:    true,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &model.AuditEntry{
				Details: model.AuditDetails{},
			}
			opt := WithDetail(tt.key, tt.value)
			opt(entry)

			assert.Equal(t, tt.expected, entry.Details[tt.key])
		})
	}
}

func TestWithDetail_MultipleDetails(t *testing.T) {
	entry := &model.AuditEntry{
		Details: model.AuditDetails{},
	}

	WithDetail("key1", "value1")(entry)
	WithDetail("key2", uint64(42))(entry)
	WithDetail("key3", true)(entry)

	assert.Len(t, entry.Details, 3)
	assert.Equal(t, "value1", entry.Details["key1"])
	assert.Equal(t, uint64(42), entry.Details["key2"])
	assert.Equal(t, true, entry.Details["key3"])
}

func TestWithGRPCMethod(t *testing.T) {
	entry := &model.AuditEntry{
		Details: model.AuditDetails{},
	}
	method := "/chorus.UserService/CreateUser"
	opt := WithGRPCMethod(method)
	opt(entry)

	assert.Equal(t, method, entry.Details["grpc_method"])
}

func TestWithGRPCStatusCode(t *testing.T) {
	tests := []struct {
		name     string
		code     codes.Code
		expected int
	}{
		{
			name:     "OK status",
			code:     codes.OK,
			expected: 0,
		},
		{
			name:     "NotFound status",
			code:     codes.NotFound,
			expected: 5,
		},
		{
			name:     "PermissionDenied status",
			code:     codes.PermissionDenied,
			expected: 7,
		},
		{
			name:     "Internal status",
			code:     codes.Internal,
			expected: 13,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &model.AuditEntry{
				Details: model.AuditDetails{},
			}
			opt := WithGRPCStatusCode(tt.code)
			opt(entry)

			assert.Equal(t, tt.expected, entry.Details["grpc_status_code"])
		})
	}
}

func TestWithErrorMessage(t *testing.T) {
	entry := &model.AuditEntry{
		Details: model.AuditDetails{},
	}
	errMsg := "failed to create user: database connection lost"
	opt := WithErrorMessage(errMsg)
	opt(entry)

	assert.Equal(t, errMsg, entry.Details["error_message"])
}

func TestMultipleOptions(t *testing.T) {
	ctx := context.WithValue(context.Background(), jwt_model.JWTClaimsContextKey, &jwt_model.JWTClaims{
		TenantID: 1,
		ID:       2,
		Username: "testuser",
	})

	entry := NewEntry(ctx, model.AuditActionWorkbenchCreate,
		WithDescription("Created workbench"),
		WithWorkspaceID(100),
		WithWorkbenchID(200),
		WithDetail("name", "Test Workbench"),
		WithDetail("status", "active"),
		WithGRPCMethod("/chorus.WorkbenchService/CreateWorkbench"),
		WithGRPCStatusCode(codes.OK),
	)

	assert.Equal(t, model.AuditActionWorkbenchCreate, entry.Action)
	assert.Equal(t, uint64(1), entry.TenantID)
	assert.Equal(t, uint64(2), entry.UserID)
	assert.Equal(t, "testuser", entry.Username)
	assert.Equal(t, "Created workbench", entry.Description)
	assert.Equal(t, uint64(100), entry.WorkspaceID)
	assert.Equal(t, uint64(200), entry.WorkbenchID)
	assert.Equal(t, "Test Workbench", entry.Details["name"])
	assert.Equal(t, "active", entry.Details["status"])
	assert.Equal(t, "/chorus.WorkbenchService/CreateWorkbench", entry.Details["grpc_method"])
	assert.Equal(t, 0, entry.Details["grpc_status_code"])
}

func TestNewEntry_PreservesUsername(t *testing.T) {
	ctx := context.WithValue(context.Background(), jwt_model.JWTClaimsContextKey, &jwt_model.JWTClaims{
		TenantID: 1,
		ID:       2,
		Username: "john.doe@example.com",
	})

	entry := NewEntry(ctx, model.AuditActionUserLogin)

	assert.Equal(t, "john.doe@example.com", entry.Username)
}

func TestRecord_CreatesEntryWithoutCancel(t *testing.T) {
	// Create a mock writer to capture the entry
	mockWriter := &mockAuditWriter{
		recordFunc: func(ctx context.Context, entry *model.AuditEntry) (*model.AuditEntry, error) {
			// Verify that the entry was created with expected values
			assert.Equal(t, model.AuditActionUserLogin, entry.Action)
			assert.Equal(t, "User logged in", entry.Description)
			assert.Equal(t, "testuser", entry.Username)
			return entry, nil
		},
	}

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, jwt_model.JWTClaimsContextKey, &jwt_model.JWTClaims{
		TenantID: 1,
		ID:       2,
		Username: "testuser",
	})

	// Record an audit entry
	Record(ctx, mockWriter, model.AuditActionUserLogin,
		WithDescription("User logged in"),
	)

	// Cancel the original context immediately
	cancel()

	// Give the goroutine time to complete
	time.Sleep(50 * time.Millisecond)

	// Verify the record function was called despite the parent context being cancelled
	require.True(t, mockWriter.called, "Record should have been called even after parent context cancellation")
}

// mockAuditWriter is a mock implementation of AuditWriter for testing
type mockAuditWriter struct {
	recordFunc func(ctx context.Context, entry *model.AuditEntry) (*model.AuditEntry, error)
	called     bool
}

func (m *mockAuditWriter) Record(ctx context.Context, entry *model.AuditEntry) (*model.AuditEntry, error) {
	m.called = true
	if m.recordFunc != nil {
		return m.recordFunc(ctx, entry)
	}
	return entry, nil
}

func TestWithDetail_InitializedDetails(t *testing.T) {
	// NewEntry always initializes Details, so WithDetail expects it to be initialized
	entry := &model.AuditEntry{
		Details: model.AuditDetails{},
	}

	// This should not panic
	WithDetail("key", "value")(entry)

	// The detail should be set
	assert.NotNil(t, entry.Details)
	assert.Equal(t, "value", entry.Details["key"])
}

func TestNewEntry_EmptyDetails(t *testing.T) {
	entry := NewEntry(context.Background(), model.AuditActionUserRead)

	// Details should be initialized but empty
	assert.NotNil(t, entry.Details)
	assert.Len(t, entry.Details, 0)
}