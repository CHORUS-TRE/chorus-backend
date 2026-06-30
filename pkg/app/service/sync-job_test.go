package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/harbor"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app/model"
)

const testRegistry = "registry.example.com"

func TestHarborAppToModels_SingleApp(t *testing.T) {
	ha := harbor.App{
		Repository: "vscode",
		Tag:        "1.106.1-11",
		Labels: map[string]string{
			"ch.chorus-tre.app.category":                    "Development",
			"ch.chorus-tre.app.icon":                        "data:image/png;base64,xxx",
			"ch.chorus-tre.app.name":                        "vscode",
			"ch.chorus-tre.app.stability":                   "ready",
			"ch.chorus-tre.image.name":                      "vscode",
			"ch.chorus-tre.image.tag":                       "1.106.1-11",
			"ch.chorus-tre.resources.max-cpu":               "1500m",
			"ch.chorus-tre.resources.max-ephemeral-storage": "10Gi",
			"ch.chorus-tre.resources.max-memory":            "3072Mi",
			"ch.chorus-tre.resources.min-cpu":               "",
			"ch.chorus-tre.resources.shared-memory-size":    "",
			"org.opencontainers.image.description":          "Lightweight source code editor",
			"org.opencontainers.image.title":                "Visual Studio Code",
		},
	}

	j := &AppSyncJob{registry: testRegistry}
	apps := j.harborAppToModels(ha, 7, 9)

	require.Len(t, apps, 1)
	app := apps[0]

	assert.Equal(t, uint64(7), app.TenantID)
	assert.Equal(t, uint64(9), app.UserID)
	// OCI title wins over the app name label.
	assert.Equal(t, "Visual Studio Code", app.Name)
	assert.Equal(t, "Development", app.Category)
	assert.Equal(t, "Lightweight source code editor", app.Description)
	assert.Equal(t, model.AppActive, app.Status)
	assert.Equal(t, testRegistry, app.DockerImageRegistry)
	assert.Equal(t, "vscode", app.DockerImageName)
	assert.Equal(t, "1.106.1-11", app.DockerImageTag)
	assert.Equal(t, "1500m", app.MaxCPU)
	assert.Equal(t, "3072Mi", app.MaxMemory)
	assert.Equal(t, "10Gi", app.MaxEphemeralStorage)
	assert.Equal(t, model.AppStabilityStatusReady, app.StabilityStatus)
	assert.Equal(t, "data:image/png;base64,xxx", app.IconURL)
	// A single app has no per-app browser config URL.
	assert.Empty(t, app.BrowserConfigURL)
}

func TestHarborAppToModels_KioskMultiApp(t *testing.T) {
	ha := harbor.App{
		Repository: "kiosk",
		Tag:        "142.0.7444.175-1",
		Labels: map[string]string{
			"ch.chorus-tre.app.name":               "kiosk",
			"ch.chorus-tre.app.stability":          "ready",
			"org.opencontainers.image.description": "Kiosk mode browser",
			"org.opencontainers.image.title":       "Kiosk",

			"ch.chorus-tre.app.kiosk-config-url.gitlab.prettyname":                   "GitLab",
			"ch.chorus-tre.app.kiosk-config-url.gitlab.category":                     "Development",
			"ch.chorus-tre.app.kiosk-config-url.gitlab.url":                          "https://gitlab.local.chorus-tre.ch",
			"ch.chorus-tre.app.kiosk-config-url.gitlab.resources.max-cpu":            "750m",
			"ch.chorus-tre.app.kiosk-config-url.gitlab.resources.max-memory":         "768Mi",
			"ch.chorus-tre.app.kiosk-config-url.gitlab.resources.shared-memory-size": "1Gi",

			"ch.chorus-tre.app.kiosk-config-url.didata.prettyname":           "HORUS Restitution",
			"ch.chorus-tre.app.kiosk-config-url.didata.category":             "CHUV Services",
			"ch.chorus-tre.app.kiosk-config-url.didata.url":                  "https://didata.local.chorus-tre.ch",
			"ch.chorus-tre.app.kiosk-config-url.didata.resources.max-cpu":    "750m",
			"ch.chorus-tre.app.kiosk-config-url.didata.resources.max-memory": "768Mi",
		},
	}

	j := &AppSyncJob{registry: testRegistry}
	apps := j.harborAppToModels(ha, 1, 1)

	require.Len(t, apps, 2)

	// Sub-apps are sorted by key: "didata" before "gitlab".
	didata, gitlab := apps[0], apps[1]

	assert.Equal(t, "HORUS Restitution", didata.Name)
	assert.Equal(t, "CHUV Services", didata.Category)
	assert.Equal(t, "https://didata.local.chorus-tre.ch", didata.BrowserConfigURL)

	assert.Equal(t, "GitLab", gitlab.Name)
	assert.Equal(t, "Development", gitlab.Category)
	assert.Equal(t, "https://gitlab.local.chorus-tre.ch", gitlab.BrowserConfigURL)
	assert.Equal(t, "750m", gitlab.MaxCPU)
	assert.Equal(t, "768Mi", gitlab.MaxMemory)
	assert.Equal(t, "1Gi", gitlab.ShmSize)

	// Every sub-app shares the same image and image-level metadata.
	for _, app := range apps {
		assert.Equal(t, "kiosk", app.DockerImageName)
		assert.Equal(t, "142.0.7444.175-1", app.DockerImageTag)
		assert.Equal(t, testRegistry, app.DockerImageRegistry)
		assert.Equal(t, "Kiosk mode browser", app.Description)
		assert.Equal(t, model.AppStabilityStatusReady, app.StabilityStatus)
		assert.Equal(t, model.AppActive, app.Status)
	}

	// Sharing one image, the sub-apps must still be distinct after dedup.
	assert.NotEqual(t, appIdentity(didata), appIdentity(gitlab))
}

func TestHarborAppToModels_SkipsNonApp(t *testing.T) {
	ha := harbor.App{
		Repository: "ubuntu",
		Tag:        "24.04",
		Labels: map[string]string{
			"org.opencontainers.image.title": "Ubuntu",
		},
	}

	j := &AppSyncJob{registry: testRegistry}
	assert.Nil(t, j.harborAppToModels(ha, 1, 1))
}

func appLabels(name, category, title string) map[string]string {
	return map[string]string{
		labelAppName:     name,
		labelAppCategory: category,
		labelOCITitle:    title,
	}
}

func TestAppsToCreate(t *testing.T) {
	j := &AppSyncJob{registry: testRegistry}

	// Deliberately out of order; Tag should drive creation order.
	harborApps := []harbor.App{
		{Repository: "newer", Tag: "2.0.0-1", Labels: appLabels("newer", "Development", "Newer")},
		{Repository: "internal", Tag: "1.5.0-1", Labels: appLabels("internal", categoryChorus, "Internal")},
		{Repository: "older", Tag: "1.0.0-1", Labels: appLabels("older", "Science", "Older")},
		{Repository: "dup", Tag: "3.0.0-1", Labels: appLabels("dup", "Science", "Existing")},
	}

	existing := map[string]struct{}{
		appIdentity(&model.App{DockerImageName: "dup", DockerImageTag: "3.0.0-1", Name: "Existing"}): {},
	}

	toCreate := j.appsToCreate(harborApps, existing, 1, 1)

	// "internal" (chorus) and "dup" (already existing) are skipped; the rest are
	// ordered by tag.
	require.Len(t, toCreate, 2)
	assert.Equal(t, "Older", toCreate[0].Name)
	assert.Equal(t, "Newer", toCreate[1].Name)
}

func TestUint64Option(t *testing.T) {
	t.Run("missing returns default", func(t *testing.T) {
		v, err := uint64Option(map[string]interface{}{}, "tenant_id", 1)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), v)
	})

	t.Run("supported types", func(t *testing.T) {
		for name, raw := range map[string]interface{}{
			"float64": float64(42),
			"int":     int(42),
			"int64":   int64(42),
			"uint64":  uint64(42),
			"string":  "42",
		} {
			v, err := uint64Option(map[string]interface{}{"tenant_id": raw}, "tenant_id", 1)
			require.NoError(t, err, name)
			assert.Equal(t, uint64(42), v, name)
		}
	})

	t.Run("invalid string errors", func(t *testing.T) {
		_, err := uint64Option(map[string]interface{}{"tenant_id": "abc"}, "tenant_id", 1)
		require.Error(t, err)
	})
}
