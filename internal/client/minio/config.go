package minio

import "github.com/CHORUS-TRE/chorus-backend/internal/config"

type MinioClientConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
}

func getMinioClientConfig(cfg config.Config) (MinioClientConfig, error) {
	return MinioClientConfig{
		Endpoint:        cfg.Clients.MinioClient.Endpoint,
		AccessKeyID:     cfg.Clients.MinioClient.AccessKeyID,
		SecretAccessKey: cfg.Clients.MinioClient.SecretAccessKey,
		UseSSL:          cfg.Clients.MinioClient.UseSSL,
		BucketName:      cfg.Clients.MinioClient.BucketName,
	}, nil
}
