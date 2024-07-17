package logger

import (
	"time"

	"go.uber.org/zap"
)

const (
	LoggerKeyLayer string = "layer"

	LoggerKeyTenantID      string = "tenant_id"
	LoggerKeyCorrelationID string = "correlation_id"
	LoggerKeyUserID        string = "user_id"

	LoggerKeyAppID         string = "app_id"
	LoggerKeyAppInstanceID string = "app_instance_id"
	LoggerKeyWorkbenchID   string = "workbench_id"
	LoggerKeyWorkspaceID   string = "workspace_id"

	LoggerKeyMethod     string = "method"
	LoggerKeyCount      string = "count"
	LoggerKeyTotalItems string = "total_items"
	LoggerKeyElapsedMs  string = "elapsed-ms"
	LoggerKeyService    string = "service"
	LoggerKeyClient     string = "client"
	LoggerKeyStartTime  string = "start_time"
	LoggerKeyEndTime    string = "end_time"

	LoggerKeyRoles        string = "roles"
	LoggerKeyCaller       string = "caller"
	LoggerKeyParentCaller string = "parent_caller"
	LoggerKeyObjectType   string = "object_type"

	LoggerMessageRequestStarted   string = "request started"
	LoggerMessageRequestFailed    string = "request failed"
	LoggerMessageRequestCompleted string = "request completed"
)

func WithLayerField(layer string) zap.Field {
	return zap.String(LoggerKeyLayer, layer)
}

func WithTenantIDField(tenantID uint64) zap.Field {
	return zap.Uint64(LoggerKeyTenantID, tenantID)
}

func WithUserIDField(userID uint64) zap.Field {
	return zap.Uint64(LoggerKeyUserID, userID)
}

func WithAppIDField(appID uint64) zap.Field {
	return zap.Uint64(LoggerKeyAppID, appID)
}

func WithAppInstanceIDField(appID uint64) zap.Field {
	return zap.Uint64(LoggerKeyAppInstanceID, appID)
}

func WithWorkspaceIDField(appID uint64) zap.Field {
	return zap.Uint64(LoggerKeyWorkbenchID, appID)
}

func WithWorkbenchIDField(appID uint64) zap.Field {
	return zap.Uint64(LoggerKeyWorkspaceID, appID)
}

func WithCountField(count int) zap.Field {
	return zap.Int(LoggerKeyCount, count)
}

func WithElapsedMsField(elapsed float64) zap.Field {
	return zap.Float64(LoggerKeyElapsedMs, elapsed)
}

func WithErrorField(err error) zap.Field {
	return zap.Error(err)
}

func WithMethodField(method string) zap.Field {
	return zap.String(LoggerKeyMethod, method)
}

func WithTotalItemsField(totalItems uint64) zap.Field {
	return zap.Uint64(LoggerKeyTotalItems, totalItems)
}

func WithStartTimeField(startTime time.Time) zap.Field {
	return zap.Time(LoggerKeyStartTime, startTime)
}

func WithEndTimeField(endTime time.Time) zap.Field {
	return zap.Time(LoggerKeyEndTime, endTime)
}
