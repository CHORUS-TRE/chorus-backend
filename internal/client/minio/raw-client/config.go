package miniorawclient

import (
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
)

type MinioClientConfig struct {
	Name                   string
	Endpoint               string
	AccessKeyID            string
	SecretAccessKey        string
	UseSSL                 bool
	BucketName             string
	MultipartMinPartSize   uint64
	MultipartMaxPartSize   uint64
	MultipartMaxTotalParts uint64
}

func getMinioClientConfig(cfg config.Config, clientName string) (MinioClientConfig, error) {
	return MinioClientConfig{
		Name:                   clientName,
		Endpoint:               cfg.Clients.MinioClients[clientName].Endpoint,
		AccessKeyID:            cfg.Clients.MinioClients[clientName].AccessKeyID,
		SecretAccessKey:        cfg.Clients.MinioClients[clientName].SecretAccessKey.PlainText(),
		UseSSL:                 cfg.Clients.MinioClients[clientName].UseSSL,
		BucketName:             cfg.Clients.MinioClients[clientName].BucketName,
		MultipartMinPartSize:   cfg.Clients.MinioClients[clientName].MultipartMinPartSize,
		MultipartMaxPartSize:   cfg.Clients.MinioClients[clientName].MultipartMaxPartSize,
		MultipartMaxTotalParts: cfg.Clients.MinioClients[clientName].MultipartMaxTotalParts,
	}, nil
}
