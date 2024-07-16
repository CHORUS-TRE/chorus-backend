package converter

import (
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/require"
)

func TestToProtoTimestamp(t *testing.T) {
	ts, err := ToProtoTimestamp(time.Time{})
	require.Nil(t, err)
	require.Nil(t, ts)
	require.Nil(t, ts)

	ts, err = ToProtoTimestamp(time.Now())
	require.NotNil(t, ts)
	require.Nil(t, err)
}

func TestFromProtoTimestamp(t *testing.T) {
	ts, err := FromProtoTimestamp(nil)
	require.True(t, ts.IsZero())
	require.Nil(t, err)

	ts, err = FromProtoTimestamp(&timestamp.Timestamp{})
	require.False(t, ts.IsZero())
	require.Nil(t, err)
}
