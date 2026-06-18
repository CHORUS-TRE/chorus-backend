package converter

import (
	"testing"
	"time"

	tenant_model "github.com/CHORUS-TRE/chorus-backend/pkg/tenant/model"
	"github.com/stretchr/testify/require"
)

func TestTenantFromBusiness(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	tenant := &tenant_model.Tenant{
		ID:           42,
		Name:         "acme",
		CreationDate: now,
		UpdateDate:   now.Add(time.Hour),
	}

	result, err := TenantFromBusiness(tenant)
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Equal(t, uint64(42), result.Id)
	require.Equal(t, "acme", result.Name)
	require.Equal(t, now.Unix(), result.CreatedAt.AsTime().Unix())
	require.Equal(t, now.Add(time.Hour).Unix(), result.UpdatedAt.AsTime().Unix())
}

func TestTenantFromBusiness_ZeroTimestamps(t *testing.T) {
	tenant := &tenant_model.Tenant{
		ID:   1,
		Name: "default",
	}

	result, err := TenantFromBusiness(tenant)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, uint64(1), result.Id)
	require.Nil(t, result.CreatedAt)
	require.Nil(t, result.UpdatedAt)
}
