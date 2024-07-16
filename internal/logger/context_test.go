package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
)

func init() {
	//nolint:errcheck
	InitLoggers(config.Config{})
}

func TestContextLogger_NoFieldsOnRootLoggers(t *testing.T) {
	parentLogger := TechLog
	require.Len(t, parentLogger.fieldKeysUsed, 0)
}

func TestContextLogger_SetFieldsOnChildLoggers(t *testing.T) {
	parentLogger := TechLog
	log := With(parentLogger, zap.String("some_key", "value"))
	require.Len(t, log.fieldKeysUsed, 1)
	_, ok := log.fieldKeysUsed["some_key"]
	require.True(t, ok)
}

func TestContextLogger_GrandChildInheritsParentFields(t *testing.T) {
	parentLogger := TechLog
	childLog := With(parentLogger, zap.String("some_key", "value"))
	grandchildLog := With(childLog, zap.String("other_key", "other_value"))
	require.Len(t, grandchildLog.fieldKeysUsed, 2)
	_, ok := grandchildLog.fieldKeysUsed["some_key"]
	require.True(t, ok)
	_, ok = grandchildLog.fieldKeysUsed["other_key"]
	require.True(t, ok)
	// parent left unchanged
	require.Len(t, childLog.fieldKeysUsed, 1)
	_, ok = childLog.fieldKeysUsed["some_key"]
	require.True(t, ok)
}

func TestContextLogger_cleanDuplicateFields_takesLoggerLevelFieldsIntoAccount(t *testing.T) {
	parentLogger := TechLog
	childLog := With(parentLogger, zap.String("some_key", "value"))

	inFields := []zap.Field{
		zap.String("some_key", "value"),
		zap.Uint64("int_key", uint64(943)),
	}
	outFields := []zap.Field{
		inFields[1],
	}
	actualOut := childLog.cleanDuplicateFields(inFields)
	require.Equal(t, outFields, actualOut)
	// childLog internal left unchanged
	require.Len(t, childLog.fieldKeysUsed, 1)
}
