package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/harbor"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app/model"

	"go.uber.org/zap"
)

const (
	labelAppName        = "ch.chorus-tre.app.name"
	labelAppIcon        = "ch.chorus-tre.app.icon"
	labelImageName      = "ch.chorus-tre.image.name"
	labelImageTag       = "ch.chorus-tre.image.tag"
	labelMaxCPU         = "ch.chorus-tre.resources.max-cpu"
	labelMinCPU         = "ch.chorus-tre.resources.min-cpu"
	labelMaxMemory      = "ch.chorus-tre.resources.max-memory"
	labelMinMemory      = "ch.chorus-tre.resources.min-memory"
	labelMaxEphemeral   = "ch.chorus-tre.resources.max-ephemeral-storage"
	labelMinEphemeral   = "ch.chorus-tre.resources.min-ephemeral-storage"
	labelShmSize        = "ch.chorus-tre.resources.shared-memory-size"
	labelOCITitle       = "org.opencontainers.image.title"
	labelOCIDescription = "org.opencontainers.image.description"
)

type AppSyncJob struct {
	store        AppStore
	harborClient harbor.HarborClient
	registry     string
	log          *logger.ContextLogger
}

func NewAppSyncJob(store AppStore, harborClient harbor.HarborClient, registry string, log *logger.ContextLogger) *AppSyncJob {
	return &AppSyncJob{
		store:        store,
		harborClient: harborClient,
		registry:     registry,
		log:          log,
	}
}

func (j *AppSyncJob) Do(ctx context.Context, options map[string]interface{}) (string, error) {
	tenantID := uint64(1)
	if v, ok := options["tenant_id"]; ok {
		switch tid := v.(type) {
		case float64:
			tenantID = uint64(tid)
		case int:
			tenantID = uint64(tid)
		case string:
			parsed, err := strconv.ParseUint(tid, 10, 64)
			if err != nil {
				return "", fmt.Errorf("invalid tenant_id option: %w", err)
			}
			tenantID = parsed
		}
	}

	userID := uint64(1)
	if v, ok := options["user_id"]; ok {
		switch uid := v.(type) {
		case float64:
			userID = uint64(uid)
		case int:
			userID = uint64(uid)
		case string:
			parsed, err := strconv.ParseUint(uid, 10, 64)
			if err != nil {
				return "", fmt.Errorf("invalid user_id option: %w", err)
			}
			userID = parsed
		}
	}

	existingApps, _, err := j.store.ListApps(ctx, tenantID, nil)
	if err != nil {
		return "", fmt.Errorf("listing existing apps: %w", err)
	}

	existingSet := make(map[string]struct{}, len(existingApps))
	for _, a := range existingApps {
		key := a.DockerImageName + ":" + a.DockerImageTag
		existingSet[key] = struct{}{}
	}

	harborApps, err := j.harborClient.ListApps(existingSet)
	if err != nil {
		return "", fmt.Errorf("listing apps from harbor: %w", err)
	}

	if len(harborApps) == 0 {
		return "no apps found in harbor", nil
	}

	var toCreate []*model.App
	for _, ha := range harborApps {
		app := j.harborAppToModel(ha, tenantID, userID)
		key := app.DockerImageName + ":" + app.DockerImageTag
		if _, exists := existingSet[key]; exists {
			continue
		}
		existingSet[key] = struct{}{}
		toCreate = append(toCreate, app)
	}

	if len(toCreate) == 0 {
		return "all apps already exist", nil
	}

	created, err := j.store.BulkCreateApps(ctx, tenantID, toCreate)
	if err != nil {
		return "", fmt.Errorf("bulk creating apps: %w", err)
	}

	j.log.Info(ctx, "synced apps from harbor",
		zap.Int("created", len(created)),
		zap.Int("harbor_total", len(harborApps)))

	return fmt.Sprintf("created %d new apps", len(created)), nil
}

func (j *AppSyncJob) harborAppToModel(ha harbor.App, tenantID, userID uint64) *model.App {
	labels := ha.Labels

	imageName := labels[labelImageName]
	if imageName == "" {
		imageName = ha.Repository
	}

	imageTag := labels[labelImageTag]
	if imageTag == "" {
		imageTag = ha.Tag
	}

	app := &model.App{
		TenantID:            tenantID,
		UserID:              userID,
		Name:                labelOrDefault(labels, labelAppName, ha.Repository),
		PrettyName:          labels[labelOCITitle],
		Description:         labels[labelOCIDescription],
		Status:              model.AppActive,
		DockerImageRegistry: j.registry,
		DockerImageName:     imageName,
		DockerImageTag:      imageTag,
		ShmSize:             labels[labelShmSize],
		MaxCPU:              labels[labelMaxCPU],
		MinCPU:              labels[labelMinCPU],
		MaxMemory:           labels[labelMaxMemory],
		MinMemory:           labels[labelMinMemory],
		MaxEphemeralStorage: labels[labelMaxEphemeral],
		MinEphemeralStorage: labels[labelMinEphemeral],
		IconURL:             labels[labelAppIcon],
	}

	return app
}

func labelOrDefault(labels map[string]string, key, fallback string) string {
	if v, ok := labels[key]; ok && v != "" {
		return v
	}
	return fallback
}
