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

func GetMinioClientConfigFromBlockStore(blockStoreName string, minioConfig *config.BlockStoreMinioConfig) MinioClientConfig {
	return MinioClientConfig{
		Name:                   blockStoreName,
		Endpoint:               minioConfig.Endpoint,
		AccessKeyID:            minioConfig.AccessKeyID,
		SecretAccessKey:        minioConfig.SecretAccessKey.PlainText(),
		UseSSL:                 minioConfig.UseSSL,
		BucketName:             minioConfig.BucketName,
		MultipartMinPartSize:   minioConfig.MultipartMinPartSize,
		MultipartMaxPartSize:   minioConfig.MultipartMaxPartSize,
		MultipartMaxTotalParts: minioConfig.MultipartMaxTotalParts,
	}
}
