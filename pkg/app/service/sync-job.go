package service

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/harbor"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app/model"

	"go.uber.org/zap"
)

const (
	labelAppName        = "ch.chorus-tre.app.name"
	labelAppIcon        = "ch.chorus-tre.app.icon"
	labelAppCategory    = "ch.chorus-tre.app.category"
	labelAppStability   = "ch.chorus-tre.app.stability"
	labelMaxCPU         = "ch.chorus-tre.resources.max-cpu"
	labelMinCPU         = "ch.chorus-tre.resources.min-cpu"
	labelMaxMemory      = "ch.chorus-tre.resources.max-memory"
	labelMinMemory      = "ch.chorus-tre.resources.min-memory"
	labelMaxEphemeral   = "ch.chorus-tre.resources.max-ephemeral-storage"
	labelMinEphemeral   = "ch.chorus-tre.resources.min-ephemeral-storage"
	labelShmSize        = "ch.chorus-tre.resources.shared-memory-size"
	labelOCITitle       = "org.opencontainers.image.title"
	labelOCIDescription = "org.opencontainers.image.description"

	// categoryChorus tags internal apps that should not be synced.
	categoryChorus = "chorus"

	// labelKioskConfigPrefix marks an image bundling several apps. Labels look like
	// "ch.chorus-tre.app.kiosk-config-url.<sub-app>.<field>"; each <sub-app> becomes
	// its own app sharing the image.
	labelKioskConfigPrefix = "ch.chorus-tre.app.kiosk-config-url."

	kioskFieldPrettyName   = "prettyname"
	kioskFieldDescription  = "description"
	kioskFieldCategory     = "category"
	kioskFieldURL          = "url"
	kioskFieldIcon         = "icon"
	kioskFieldMaxCPU       = "resources.max-cpu"
	kioskFieldMinCPU       = "resources.min-cpu"
	kioskFieldMaxMemory    = "resources.max-memory"
	kioskFieldMinMemory    = "resources.min-memory"
	kioskFieldMaxEphemeral = "resources.max-ephemeral-storage"
	kioskFieldMinEphemeral = "resources.min-ephemeral-storage"
	kioskFieldSharedMemory = "resources.shared-memory-size"
)

type AppSyncJob struct {
	appStore     AppStore
	appService   Apper
	harborClient harbor.HarborClient
	registry     string
	log          *logger.ContextLogger
}

func NewAppSyncJob(appStore AppStore, appService Apper, harborClient harbor.HarborClient, registry string, log *logger.ContextLogger) *AppSyncJob {
	return &AppSyncJob{
		appStore:     appStore,
		appService:   appService,
		harborClient: harborClient,
		registry:     registry,
		log:          log,
	}
}

func (j *AppSyncJob) Do(ctx context.Context, options map[string]interface{}) (string, error) {
	tenantID, err := uint64Option(options, "tenant_id", 1)
	if err != nil {
		return "", err
	}
	userID, err := uint64Option(options, "user_id", 1)
	if err != nil {
		return "", err
	}

	// List existing apps in store
	existingApps, _, err := j.appStore.ListApps(ctx, tenantID, nil)
	if err != nil {
		return "", fmt.Errorf("listing existing apps: %w", err)
	}

	existing := make(map[string]struct{}, len(existingApps))
	for _, a := range existingApps {
		existing[appIdentity(a)] = struct{}{}
	}
	logger.TechLog.Debug(ctx, fmt.Sprintf("app-sync: listed %d existing apps", len(existingApps)))

	// List apps and versions in Harbor
	harborApps, err := j.harborClient.ListApps()
	if err != nil {
		return "", fmt.Errorf("listing apps from harbor: %w", err)
	}
	logger.TechLog.Debug(ctx, fmt.Sprintf("app-sync: listed %d apps from harbor", len(harborApps)))

	toCreate := j.appsToCreate(harborApps, existing, tenantID, userID)
	if len(toCreate) == 0 {
		return "all apps already exist", nil
	}

	created, err := j.appService.BulkCreateApps(ctx, toCreate)
	if err != nil {
		return "", fmt.Errorf("bulk creating apps: %w", err)
	}

	j.log.Info(ctx, "synced apps from harbor", zap.Int("created", len(created)))

	return fmt.Sprintf("created %d new apps", len(created)), nil
}

func (j *AppSyncJob) appsToCreate(harborApps []harbor.App, existing map[string]struct{}, tenantID, userID uint64) []*model.App {
	// Sort Harbor apps by tag, lowest version first.
	sort.SliceStable(harborApps, func(i, k int) bool {
		return harborApps[i].Tag < harborApps[k].Tag
	})

	var toCreate []*model.App
	for _, ha := range harborApps {
		for _, app := range j.harborAppToModels(ha, tenantID, userID) {
			if app.Category == categoryChorus {
				continue
			}

			id := appIdentity(app)
			if _, exists := existing[id]; exists {
				continue
			}
			existing[id] = struct{}{}
			toCreate = append(toCreate, app)
		}
	}
	return toCreate
}

func (j *AppSyncJob) harborAppToModels(ha harbor.App, tenantID, userID uint64) []*model.App {
	labels := ha.Labels
	if _, ok := labels[labelAppName]; !ok {
		return nil
	}

	// base carries everything shared by every app built from this image
	base := model.App{
		TenantID:            tenantID,
		UserID:              userID,
		Description:         labels[labelOCIDescription],
		Status:              model.AppActive,
		DockerImageRegistry: j.registry,
		DockerImageName:     ha.Repository,
		DockerImageTag:      ha.Tag,
		IconURL:             labels[labelAppIcon],
		StabilityStatus:     model.AppStabilityStatus(labels[labelAppStability]),
	}

	if subApps := parseKioskSubApps(labels); len(subApps) > 0 {
		return kioskApps(base, subApps)
	}

	app := base
	app.Name = labelOrDefault(labels, labelOCITitle, labels[labelAppName])
	app.Category = labels[labelAppCategory]
	app.ShmSize = labels[labelShmSize]
	app.MaxCPU = labels[labelMaxCPU]
	app.MinCPU = labels[labelMinCPU]
	app.MaxMemory = labels[labelMaxMemory]
	app.MinMemory = labels[labelMinMemory]
	app.MaxEphemeralStorage = labels[labelMaxEphemeral]
	app.MinEphemeralStorage = labels[labelMinEphemeral]
	return []*model.App{&app}
}

// kioskApps expands the parsed kiosk sub-apps into one app each, sorted by
// sub-app key for deterministic output.
func kioskApps(base model.App, subApps map[string]map[string]string) []*model.App {
	keys := make([]string, 0, len(subApps))
	for k := range subApps {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	apps := make([]*model.App, 0, len(keys))
	for _, key := range keys {
		fields := subApps[key]

		app := base
		app.Name = labelOrDefault(fields, kioskFieldPrettyName, key)
		app.Description = labelOrDefault(fields, kioskFieldDescription, base.Description)
		app.Category = fields[kioskFieldCategory]
		app.IconURL = labelOrDefault(fields, kioskFieldIcon, base.IconURL)
		app.BrowserConfigURL = fields[kioskFieldURL]
		app.ShmSize = fields[kioskFieldSharedMemory]
		app.MaxCPU = fields[kioskFieldMaxCPU]
		app.MinCPU = fields[kioskFieldMinCPU]
		app.MaxMemory = fields[kioskFieldMaxMemory]
		app.MinMemory = fields[kioskFieldMinMemory]
		app.MaxEphemeralStorage = fields[kioskFieldMaxEphemeral]
		app.MinEphemeralStorage = fields[kioskFieldMinEphemeral]
		apps = append(apps, &app)
	}
	return apps
}

// parseKioskSubApps groups kiosk-config-url labels by sub-app name, returning
// sub-app name -> field -> value (e.g. "gitlab" -> "url" -> "https://...").
func parseKioskSubApps(labels map[string]string) map[string]map[string]string {
	var subApps map[string]map[string]string
	for k, v := range labels {
		rest, ok := strings.CutPrefix(k, labelKioskConfigPrefix)
		if !ok {
			continue
		}
		name, field, ok := strings.Cut(rest, ".")
		if !ok {
			continue
		}
		if subApps == nil {
			subApps = make(map[string]map[string]string)
		}
		if subApps[name] == nil {
			subApps[name] = make(map[string]string)
		}
		subApps[name][field] = v
	}
	return subApps
}

// appIdentity uniquely identifies an app
func appIdentity(a *model.App) string {
	return a.DockerImageName + "\x00" + a.DockerImageTag + "\x00" + a.Name
}

func uint64Option(options map[string]interface{}, key string, def uint64) (uint64, error) {
	v, ok := options[key]
	if !ok {
		return def, nil
	}
	switch n := v.(type) {
	case float64:
		return uint64(n), nil
	case int:
		return uint64(n), nil
	case int64:
		return uint64(n), nil
	case uint64:
		return n, nil
	case string:
		parsed, err := strconv.ParseUint(n, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid %s option: %w", key, err)
		}
		return parsed, nil
	default:
		return def, nil
	}
}

func labelOrDefault(labels map[string]string, key, fallback string) string {
	if v, ok := labels[key]; ok && v != "" {
		return v
	}
	return fallback
}
