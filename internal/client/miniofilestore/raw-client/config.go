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

func GetMinioClientConfigFromFileStore(fileStoreName string, minioConfig config.FileStoreMinioConfig) MinioClientConfig {
	return MinioClientConfig{
		Name:                   fileStoreName,
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
