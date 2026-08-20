//go:build unit

package k8s

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

func TestMain(m *testing.M) {
	// The converters log a warning on invalid web-service ports; give the package
	// a no-op logger so the tests don't nil-deref the global TechLog.
	logger.TechLog = logger.NewNop()
	logger.BizLog = logger.NewNop()
	logger.SecLog = logger.NewNop()
	os.Exit(m.Run())
}

func TestAppInstanceToK8sWorkbenchApp_WebService(t *testing.T) {
	c := &client{}

	// AppRegistry is set so the image branch never dereferences the (empty) config.
	base := func(port string) AppInstance {
		return AppInstance{
			ID:                   1,
			AppName:              "jupyterlab",
			AppRegistry:          "registry.example.com",
			AppImage:             "jupyterlab",
			AppTag:               "4.6.1-1",
			WebServicePort:       port,
			WebServicePath:       "/lab",
			WebServiceTokenParam: "token",
		}
	}

	t.Run("valid port sets WebService", func(t *testing.T) {
		w := c.appInstanceToK8sWorkbenchApp(base("8888"))
		require.NotNil(t, w.WebService)
		assert.Equal(t, 8888, w.WebService.Port)
		assert.Equal(t, "/lab", w.WebService.Path)
		assert.Equal(t, "token", w.WebService.TokenParam)
	})

	// Empty means "not a web-service app"; every other invalid value is ignored
	// (WebService left nil), never silently coerced to port 0.
	for name, port := range map[string]string{
		"empty":       "",
		"non-numeric": "abc",
		"zero":        "0",
		"negative":    "-1",
		"above range": "70000",
	} {
		t.Run("no WebService for "+name, func(t *testing.T) {
			w := c.appInstanceToK8sWorkbenchApp(base(port))
			assert.Nil(t, w.WebService)
		})
	}
}

func TestK8sWorkbenchAppToAppInstance_WebService(t *testing.T) {
	c := &client{}

	t.Run("nil WebService leaves fields empty, not \"0\"", func(t *testing.T) {
		app, err := c.k8sWorkbenchAppToAppInstance(WorkbenchApp{Name: "jupyterlab-1"})
		require.NoError(t, err)
		assert.Empty(t, app.WebServicePort)
		assert.Empty(t, app.WebServicePath)
		assert.Empty(t, app.WebServiceTokenParam)
	})

	t.Run("set WebService round-trips port back to a string", func(t *testing.T) {
		app, err := c.k8sWorkbenchAppToAppInstance(WorkbenchApp{
			Name:       "jupyterlab-1",
			WebService: &WebServiceConfig{Port: 8888, Path: "/lab", TokenParam: "token"},
		})
		require.NoError(t, err)
		assert.Equal(t, "8888", app.WebServicePort)
		assert.Equal(t, "/lab", app.WebServicePath)
		assert.Equal(t, "token", app.WebServiceTokenParam)
	})
}
