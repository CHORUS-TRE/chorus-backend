package uuid

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNext(t *testing.T) {
	uuid := Next()
	require.NotEqual(t, "", uuid)
}
