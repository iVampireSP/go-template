package s3

import (
	"go-template/internal/infra/conf"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3 struct {
	Client *minio.Client
	Bucket string
}

func NewS3(config *conf.Config) *S3 {
	minioClient, err := minio.New(config.S3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.S3.AccessKey, config.S3.SecretKey, ""),
		Secure: config.S3.UseSSL,
	})

	if err != nil {
		panic(err)
	}

	return &S3{
		Client: minioClient,
		Bucket: config.S3.Bucket,
	}
}
