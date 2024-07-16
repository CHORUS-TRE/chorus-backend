package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoleType_String(t *testing.T) {

	assert.Equal(t, "admin", RoleAdmin.String())
	assert.Equal(t, "authenticated", RoleAuthenticated.String())
	assert.Equal(t, "chorus", RoleChorus.String())
}
