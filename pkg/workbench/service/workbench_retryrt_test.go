package service

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/tests/unit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	unit.InitTestLogger()
}

// roundTripFunc adapts a function to http.RoundTripper for testing.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func cfgWithMaxRetry(n int) config.Config {
	var cfg config.Config
	cfg.Services.WorkbenchService.RoundTripper.MaxTransientRetry = n
	return cfg
}

func TestRetryRT_SkipsRetryForWebSocketUpgrade(t *testing.T) {
	callCount := 0
	rt := retryRT{
		rt: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			callCount++
			return nil, errors.New("connection reset by peer")
		}),
		cfg: cfgWithMaxRetry(3),
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")

	_, err := rt.RoundTrip(req)
	require.Error(t, err)
	assert.Equal(t, 1, callCount, "should not retry WebSocket upgrades")
}

func TestRetryRT_RetriesTransientErrors(t *testing.T) {
	callCount := 0
	rt := retryRT{
		rt: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			callCount++
			return nil, errors.New("connection reset by peer")
		}),
		cfg: cfgWithMaxRetry(3),
	}

	req := httptest.NewRequest("GET", "/", nil)

	_, err := rt.RoundTrip(req)
	require.Error(t, err)
	assert.Equal(t, 3, callCount, "should retry 3 times on transient error")
}

func TestRetryRT_DoesNotRetryNonTransientErrors(t *testing.T) {
	callCount := 0
	rt := retryRT{
		rt: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			callCount++
			return nil, errors.New("some permanent error")
		}),
		cfg: cfgWithMaxRetry(3),
	}

	req := httptest.NewRequest("GET", "/", nil)

	_, err := rt.RoundTrip(req)
	require.Error(t, err)
	assert.Equal(t, 1, callCount, "should not retry non-transient errors")
}

func TestRetryRT_ReturnsOnSuccess(t *testing.T) {
	callCount := 0
	rt := retryRT{
		rt: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount == 2 {
				return &http.Response{StatusCode: 200}, nil
			}
			return nil, errors.New("connection reset by peer")
		}),
		cfg: cfgWithMaxRetry(3),
	}

	req := httptest.NewRequest("GET", "/", nil)

	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, 2, callCount, "should stop retrying after success")
}
