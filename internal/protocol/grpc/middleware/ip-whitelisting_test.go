package middleware

import (
	"context"
	"testing"

	"github.com/CHORUS-TRE/chorus-backend/tests/unit"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/protocol/grpc/middleware/mocks"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func init() {
	unit.InitTestLogger()
}

func TestNewIPWhitelistUnaryServerInterceptor(t *testing.T) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetDefault("tenants.88888.enabled", true)
	v.SetDefault("tenants.88888.ip_whitelist.enabled", true)
	v.SetDefault("tenants.88888.ip_whitelist.subnetworks", []string{"10.1.1.0/24", "10.2.0.0/16"})

	var cfg config.Config
	err := v.Unmarshal(&cfg, func(c *mapstructure.DecoderConfig) { c.TagName = "yaml" })
	require.NoError(t, err)

	whitelister, err := NewIPWhitelister(cfg)
	require.NoError(t, err)
	require.NotNil(t, whitelister)

	interceptor := NewIPWhitelistUnaryServerInterceptor(whitelister)

	ctx := context.WithValue(context.Background(), logger.TenantIDContextKey{}, uint64(88888))
	ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(headerXForwardedFor, "10.1.1.123, 127.0.0.1"))

	_, err = interceptor(ctx, nil, &grpc.UnaryServerInfo{}, testHandler)
	require.NoError(t, err)

	ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(headerXForwardedFor, "10.50.1.123"))
	_, err = interceptor(ctx, nil, &grpc.UnaryServerInfo{}, testHandler)
	require.Error(t, err)

	status, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.PermissionDenied, status.Code())
}

func TestNewIPWhitelistUnaryServerInterceptor_AuthPath(t *testing.T) {

	whitelister, err := NewIPWhitelister(config.Config{})
	require.NoError(t, err)
	require.NotNil(t, whitelister)

	interceptor := NewIPWhitelistUnaryServerInterceptor(whitelister)

	info := &grpc.UnaryServerInfo{
		FullMethod: "/chorus.AuthenticationService/Authenticate",
	}

	_, err = interceptor(context.Background(), nil, info, testHandler)
	require.NoError(t, err)
}

func TestNewIPWhitelistStreamServerInterceptor(t *testing.T) {

	stream := &mocks.ServerStream{}

	v := viper.New()
	v.SetConfigType("yaml")
	v.SetDefault("tenants.88888.enabled", true)
	v.SetDefault("tenants.88888.ip_whitelist.enabled", true)
	v.SetDefault("tenants.88888.ip_whitelist.subnetworks", []string{"10.1.1.0/24", "10.2.0.0/16"})

	var cfg config.Config
	err := v.Unmarshal(&cfg, func(c *mapstructure.DecoderConfig) { c.TagName = "yaml" })
	require.NoError(t, err)

	whitelister, err := NewIPWhitelister(cfg)
	require.NoError(t, err)
	require.NotNil(t, whitelister)

	interceptor := NewIPWhitelistStreamServerInterceptor(whitelister)

	ctx := context.WithValue(context.Background(), logger.TenantIDContextKey{}, uint64(88888))
	ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(headerXForwardedFor, "10.1.1.123"))

	stream.On("Context").Return(ctx).Once()

	err = interceptor(ctx, stream, &grpc.StreamServerInfo{}, testStreamHandler)
	require.NoError(t, err)

	ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(headerXForwardedFor, "10.50.1.123"))
	stream.On("Context").Return(ctx).Once()

	err = interceptor(ctx, stream, &grpc.StreamServerInfo{}, testStreamHandler)
	require.Error(t, err)

	status, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.PermissionDenied, status.Code())
}

func TestIPWhitelister_Verify(t *testing.T) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetDefault("daemon.http.header_client_ip", "True-Client-IP")
	v.SetDefault("tenants.88888.enabled", true)
	v.SetDefault("tenants.88888.ip_whitelist.enabled", true)
	v.SetDefault("tenants.88888.ip_whitelist.subnetworks", []string{"10.1.1.0/24", "10.2.0.0/16"})

	var cfg config.Config
	err := v.Unmarshal(&cfg, func(c *mapstructure.DecoderConfig) { c.TagName = "yaml" })
	require.NoError(t, err)

	whitelister, err := NewIPWhitelister(cfg)
	require.NoError(t, err)
	require.NotNil(t, whitelister)

	ctx := context.WithValue(context.Background(), logger.TenantIDContextKey{}, uint64(88888))

	ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("True-Client-IP", "10.1.1.123"))
	err = whitelister.Verify(ctx)
	require.NoError(t, err)

	ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("True-Client-IP", "10.3.0.1"))
	err = whitelister.Verify(ctx)
	require.EqualError(t, err, "client is not allowed")
}

func TestIPWhitelister_NotFromGateway_Verify(t *testing.T) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetDefault("daemon.http.header_client_ip", "True-Client-IP")
	v.SetDefault("tenants.88888.enabled", true)
	v.SetDefault("tenants.88888.ip_whitelist.enabled", true)
	v.SetDefault("tenants.88888.ip_whitelist.subnetworks", []string{"10.1.1.0/24", "10.2.0.0/16"})

	var cfg config.Config
	err := v.Unmarshal(&cfg, func(c *mapstructure.DecoderConfig) { c.TagName = "yaml" })
	require.NoError(t, err)

	whitelister, err := NewIPWhitelister(cfg)
	require.NoError(t, err)
	require.NotNil(t, whitelister)

	ctx := context.WithValue(context.Background(), logger.TenantIDContextKey{}, uint64(88888))

	ctx = metadata.NewIncomingContext(ctx, metadata.Pairs())
	err = whitelister.Verify(ctx)
	require.NoError(t, err)
}

func TestIPWhitelister_MissingHeader_Verify(t *testing.T) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetDefault("daemon.http.header_client_ip", "Another-Client-IP")
	v.SetDefault("tenants.88888.enabled", true)
	v.SetDefault("tenants.88888.ip_whitelist.enabled", true)

	var cfg config.Config
	err := v.Unmarshal(&cfg, func(c *mapstructure.DecoderConfig) { c.TagName = "yaml" })
	require.NoError(t, err)

	whitelister, err := NewIPWhitelister(cfg)
	require.NoError(t, err)
	require.NotNil(t, whitelister)

	ctx := context.WithValue(context.Background(), logger.TenantIDContextKey{}, uint64(88888))

	ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(headerXForwardedFor, "127.0.0.1"))
	err = whitelister.Verify(ctx)
	require.EqualError(t, err, "missing client ip header in metadata")
}

func TestIPWhitelister_NoTenant_Verify(t *testing.T) {
	whitelister, err := NewIPWhitelister(config.Config{})
	require.NoError(t, err)
	require.NotNil(t, whitelister)

	err = whitelister.Verify(context.Background())
	require.NoError(t, err)
}

func TestIPWhitelister_NoMetadata_Verify(t *testing.T) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetDefault("tenants.88888.enabled", true)
	v.SetDefault("tenants.88888.ip_whitelist.enabled", true)

	var cfg config.Config
	err := v.Unmarshal(&cfg, func(c *mapstructure.DecoderConfig) { c.TagName = "yaml" })
	require.NoError(t, err)

	whitelister, err := NewIPWhitelister(cfg)
	require.NoError(t, err)
	require.NotNil(t, whitelister)

	ctx := context.WithValue(context.Background(), logger.TenantIDContextKey{}, uint64(88888))

	err = whitelister.Verify(ctx)
	require.EqualError(t, err, "missing metadata in context")
}

func TestIPWhitelister_TenantDisabled_Verify(t *testing.T) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetDefault("tenants.88888.enabled", true)
	v.SetDefault("tenants.88888.ip_whitelist.enabled", true)
	v.SetDefault("tenants.88888.ip_whitelist.subnetworks", []string{"10.1.1.0/24"})
	v.SetDefault("tenants.88889.enabled", true)
	v.SetDefault("tenants.88889.ip_whitelist.enabled", false)
	v.SetDefault("tenants.88889.ip_whitelist.subnetworks", []string{"10.1.1.0/24"})

	var cfg config.Config
	err := v.Unmarshal(&cfg, func(c *mapstructure.DecoderConfig) { c.TagName = "yaml" })
	require.NoError(t, err)

	whitelister, err := NewIPWhitelister(cfg)
	require.NoError(t, err)
	require.NotNil(t, whitelister)

	ctx := context.WithValue(context.Background(), logger.TenantIDContextKey{}, uint64(88888))
	ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(headerXForwardedFor, "10.12.1.123"))
	err = whitelister.Verify(ctx)
	require.EqualError(t, err, "client is not allowed")

	ctx = context.WithValue(context.Background(), logger.TenantIDContextKey{}, uint64(88889))
	ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(headerXForwardedFor, "10.12.1.123"))
	err = whitelister.Verify(ctx)
	require.NoError(t, err)
}

// Utilities -------------------------------------------------------------------

func testHandler(context.Context, interface{}) (interface{}, error) {
	return nil, nil
}

func testStreamHandler(interface{}, grpc.ServerStream) error {
	return nil
}
