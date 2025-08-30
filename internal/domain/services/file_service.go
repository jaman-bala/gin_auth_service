package services

import (
	"context"
	"fmt"
	"gold_portal/config"
	"gold_portal/internal/infrastructure/storage"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type FileService interface {
	UploadFile(ctx context.Context, file *multipart.FileHeader, modelName string) (string, error)
	GetFileURL(ctx context.Context, objectName string, expiry time.Duration) (string, error)
}

type fileService struct {
	minioStorage *storage.MinioStorage
	config       *config.Config
}

func NewFileService(cfg *config.Config) (FileService, error) {
	var minioStorage *storage.MinioStorage
	var err error

	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		endpoint := fmt.Sprintf("%s:%s", cfg.Minio.MinioHost, cfg.Minio.MinioPort)

		minioStorage, err = storage.NewMinioStorage(endpoint, cfg.Minio.MinioAccessKey, cfg.Minio.MinioSecretKey, cfg.Minio.MinioSSL)
		if err == nil {
			break
		}

		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create minio storage after %d retries: %w", maxRetries, err)
	}

	exists, err := minioStorage.BucketExists(cfg.Minio.MinioBucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = minioStorage.CreateBucket(cfg.Minio.MinioBucket)
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &fileService{
		minioStorage: minioStorage,
		config:       cfg,
	}, nil
}

func (s *fileService) UploadFile(ctx context.Context, file *multipart.FileHeader, modelName string) (string, error) {
	if file == nil {
		return "", fmt.Errorf("file is required")
	}

	if modelName == "" {
		return "", fmt.Errorf("model name is required")
	}

	fileExt := filepath.Ext(file.Filename)
	fileName := fmt.Sprintf("%s_%s%s", uuid.New().String(), time.Now().Format("20060102150405"), fileExt)

	objectName := fmt.Sprintf("%s/%s", modelName, fileName)

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err = s.minioStorage.UploadFile(s.config.Minio.MinioBucket, objectName, src, file.Size, contentType, s.config.Minio.MinioSSL)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to minio: %w", err)
	}

	return fmt.Sprintf("/%s/%s", modelName, fileName), nil
}

func (s *fileService) GetFileURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	if objectName == "" {
		return "", fmt.Errorf("object name is required")
	}

	url, err := s.minioStorage.GetFileURL(s.config.Minio.MinioBucket, objectName, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to get file URL: %w", err)
	}

	return url, nil
}
