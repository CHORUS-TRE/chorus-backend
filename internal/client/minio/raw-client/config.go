package miniorawclient

import (
	"fmt"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
)

type MinioClientConfig struct {
	Name            string
	Prefix          string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
}

func getMinioClientConfig(cfg config.Config, clientName string) (MinioClientConfig, error) {
	prefix := cfg.Services.WorkspaceFileService.MinioStores[clientName].Prefix
	if prefix == "" {
		return MinioClientConfig{}, fmt.Errorf("minio client %s must have a prefix configured", clientName)
	}

	return MinioClientConfig{
		Name:            clientName,
		Prefix:          formatClientPrefix(prefix),
		Endpoint:        cfg.Services.WorkspaceFileService.MinioStores[clientName].Endpoint,
		AccessKeyID:     cfg.Services.WorkspaceFileService.MinioStores[clientName].AccessKeyID,
		SecretAccessKey: cfg.Services.WorkspaceFileService.MinioStores[clientName].SecretAccessKey,
		UseSSL:          cfg.Services.WorkspaceFileService.MinioStores[clientName].UseSSL,
		BucketName:      cfg.Services.WorkspaceFileService.MinioStores[clientName].BucketName,
	}, nil
}

// Ensures leading and trailing '/' on client prefix
func formatClientPrefix(prefix string) string {
	normalizedPrefix := "/" + strings.TrimPrefix(prefix, "/")
	normalizedPrefix = strings.TrimSuffix(normalizedPrefix, "/") + "/"

	return normalizedPrefix
}
