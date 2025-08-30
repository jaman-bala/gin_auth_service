package storage

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorage struct {
	client *minio.Client
}

func NewMinioStorage(endpoint, accessKey, secretKey string, useSSL bool) (*MinioStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return &MinioStorage{client: client}, nil
}

func (m *MinioStorage) UploadFile(bucket, objectName string, file multipart.File, fileSize int64, contentType string, useSSL bool) (string, error) {
	ctx := context.Background()

	// проверяем что bucket существует
	exists, err := m.client.BucketExists(ctx, bucket)
	if err != nil {
		return "", err
	}
	if !exists {
		err = m.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return "", err
		}
	}

	_, err = m.client.PutObject(ctx, bucket, objectName, file, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	// генерируем presigned URL на скачивание
	url, err := m.client.PresignedGetObject(ctx, bucket, objectName, time.Hour*24, nil)
	if err != nil {
		return "", err
	}

	// заменяем протокол на тот, который нужен
	if useSSL {
		url.Scheme = "https"
	} else {
		url.Scheme = "http"
	}

	return url.String(), nil
}

func (m *MinioStorage) DeleteFile(bucket, objectName string) error {
	ctx := context.Background()
	return m.client.RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{})
}

func (m *MinioStorage) GetFileURL(bucket, objectName string, expiry time.Duration) (string, error) {
	ctx := context.Background()
	if expiry == 0 {
		expiry = time.Hour * 24
	}

	url, err := m.client.PresignedGetObject(ctx, bucket, objectName, expiry, nil)
	if err != nil {
		return "", err
	}

	return url.String(), nil
}

func (m *MinioStorage) BucketExists(bucket string) (bool, error) {
	ctx := context.Background()
	return m.client.BucketExists(ctx, bucket)
}

func (m *MinioStorage) CreateBucket(bucket string) error {
	ctx := context.Background()
	return m.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
}
