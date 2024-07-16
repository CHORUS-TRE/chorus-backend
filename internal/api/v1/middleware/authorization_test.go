package middleware

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func buildAuthorization(roles ...string) authorization {
	return authorization{authorizedRoles: roles}
}

func TestIsAuthorized(t *testing.T) {
	require.False(t, buildAuthorization("").isAuthorized([]string{"client"}))
	require.False(t, buildAuthorization("admin").isAuthorized([]string{"client"}))
	require.False(t, buildAuthorization("admin").isAuthorized([]string{""}))
	require.True(t, buildAuthorization("admin").isAuthorized([]string{"admin"}))
	require.True(t, buildAuthorization("admin").isAuthorized([]string{"admin", "role1"}))
	require.True(t, buildAuthorization("admin", "role1").isAuthorized([]string{"admin"}))
	require.True(t, buildAuthorization("admin", "role1").isAuthorized([]string{"admin", "role2"}))
}

func TestHasRole(t *testing.T) {
	require.False(t, hasRole("", []string{"client"}))
	require.False(t, hasRole("admin", []string{"client"}))
	require.True(t, hasRole("admin", []string{"admin", "client"}))
}
