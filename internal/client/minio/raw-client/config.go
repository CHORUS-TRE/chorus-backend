package miniorawclient

import (
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
)

type MinioClientConfig struct {
	Name            string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
}

func getMinioClientConfig(cfg config.Config, clientName string) (MinioClientConfig, error) {
	return MinioClientConfig{
		Name:            clientName,
		Endpoint:        cfg.Clients.MinioClients[clientName].Endpoint,
		AccessKeyID:     cfg.Clients.MinioClients[clientName].AccessKeyID,
		SecretAccessKey: cfg.Clients.MinioClients[clientName].SecretAccessKey.PlainText(),
		UseSSL:          cfg.Clients.MinioClients[clientName].UseSSL,
		BucketName:      cfg.Clients.MinioClients[clientName].BucketName,
	}, nil
}
