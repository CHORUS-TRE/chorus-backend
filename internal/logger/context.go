package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/CHORUS-TRE/chorus-backend/internal/utils/correlation"
)

// ContextLogger is a logger that automatically logs predefined key value taken from the context.
// The logged fields are defined in the function 'appendContextFields'.
// ContextLogger embed two loggers (Logger and loggerCallerSkip). It is needed to print the caller (filename
// and line number) accurately. Because it is wrapped, loggerCallerSkip must be instantiatied with the option
// AddCallerSkip(1) while the other logger must not.
type ContextLogger struct {
	Logger           *zap.Logger
	loggerCallerSkip *zap.Logger
	fieldKeysUsed    map[string]struct{}
}

func NewNop() *ContextLogger {
	return &ContextLogger{
		Logger:           zap.NewNop(),
		loggerCallerSkip: zap.NewNop(),
	}
}

type UserIDContextKey struct{}
type TenantIDContextKey struct{}
type RolesContextKey struct{}
type JWTRenewableAmountKey struct{}
type LoggedInWithSSOKey struct{}

func appendContextFields(ctx context.Context, fields []zapcore.Field) []zapcore.Field {

	if correlationID, ok := ctx.Value(correlation.CorrelationIDContextKey{}).(string); ok {
		fields = append(fields, zap.String(LoggerKeyCorrelationID, correlationID))
	}
	if userId, ok := ctx.Value(UserIDContextKey{}).(uint64); ok && userId != 0 {
		fields = append(fields, zap.Uint64(LoggerKeyUserID, userId))
	}
	if tenantId, ok := ctx.Value(TenantIDContextKey{}).(uint64); ok && tenantId != 0 {
		fields = append(fields, zap.Uint64(LoggerKeyTenantID, tenantId))
	}
	if roles, ok := ctx.Value(RolesContextKey{}).([]string); ok {
		fields = append(fields, zap.Strings(LoggerKeyRoles, roles))
	}

	return fields
}

func (l *ContextLogger) cleanDuplicateFields(fields []zapcore.Field) []zapcore.Field {

	m := map[string]struct{}{}
	for fieldKey := range l.fieldKeysUsed {
		m[fieldKey] = struct{}{}
	}

	cleanedFields := []zapcore.Field{}

	for _, f := range fields {
		if _, ok := m[f.Key]; !ok {
			m[f.Key] = struct{}{}
			cleanedFields = append(cleanedFields, f)
		}
	}

	return cleanedFields
}

func (l *ContextLogger) DPanic(ctx context.Context, msg string, fields ...zapcore.Field) {
	ctxFields := appendContextFields(ctx, fields)
	ctxFields = l.cleanDuplicateFields(ctxFields)
	l.loggerCallerSkip.DPanic(msg, ctxFields...)
}

func (l *ContextLogger) Debug(ctx context.Context, msg string, fields ...zapcore.Field) {
	ctxFields := appendContextFields(ctx, fields)
	ctxFields = l.cleanDuplicateFields(ctxFields)
	l.loggerCallerSkip.Debug(msg, ctxFields...)
}

func (l *ContextLogger) Error(ctx context.Context, msg string, fields ...zapcore.Field) {
	ctxFields := appendContextFields(ctx, fields)
	ctxFields = l.cleanDuplicateFields(ctxFields)
	l.loggerCallerSkip.Error(msg, ctxFields...)
}

func (l *ContextLogger) Fatal(ctx context.Context, msg string, fields ...zapcore.Field) {
	ctxFields := appendContextFields(ctx, fields)
	ctxFields = l.cleanDuplicateFields(ctxFields)
	l.loggerCallerSkip.Fatal(msg, ctxFields...)
}

func (l *ContextLogger) Info(ctx context.Context, msg string, fields ...zapcore.Field) {
	ctxFields := appendContextFields(ctx, fields)
	ctxFields = l.cleanDuplicateFields(ctxFields)
	l.loggerCallerSkip.Info(msg, ctxFields...)
}

func (l *ContextLogger) Panic(ctx context.Context, msg string, fields ...zapcore.Field) {
	ctxFields := appendContextFields(ctx, fields)
	ctxFields = l.cleanDuplicateFields(ctxFields)
	l.loggerCallerSkip.Panic(msg, ctxFields...)
}

func (l *ContextLogger) Warn(ctx context.Context, msg string, fields ...zapcore.Field) {
	ctxFields := appendContextFields(ctx, fields)
	ctxFields = l.cleanDuplicateFields(ctxFields)
	l.loggerCallerSkip.Warn(msg, ctxFields...)
}

// With creates a child logger and adds structured context to it. Fields added
// to the child don't affect the parent, and vice versa.
func With(logger *ContextLogger, fields ...zap.Field) *ContextLogger {
	var fieldsToAdd []zap.Field
	fieldKeysToAdd := map[string]struct{}{}
	for fieldKey := range logger.fieldKeysUsed {
		fieldKeysToAdd[fieldKey] = struct{}{}
	}

	for _, field := range fields {
		if _, ok := logger.fieldKeysUsed[field.Key]; !ok {
			fieldsToAdd = append(fieldsToAdd, field)
			fieldKeysToAdd[field.Key] = struct{}{}
		}
	}

	return &ContextLogger{
		Logger:           logger.Logger.With(fieldsToAdd...),
		loggerCallerSkip: logger.loggerCallerSkip.With(fieldsToAdd...),
		fieldKeysUsed:    fieldKeysToAdd,
	}
}

func NewContextLogger(logger *zap.Logger) *ContextLogger {
	return &ContextLogger{
		Logger:           logger,
		loggerCallerSkip: logger.WithOptions(zap.AddCallerSkip(1)),
		fieldKeysUsed:    map[string]struct{}{},
	}
}
