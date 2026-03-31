package main

import (
	"testing"

	"github.com/xxbbzy/gonext-template/backend/internal/config"
	"github.com/xxbbzy/gonext-template/backend/internal/repository"
)

func TestNewUploadStorageRepositoryCreatesLocalRepository(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			BaseURL: "http://localhost:8080",
		},
		Upload: config.UploadConfig{
			Dir: "./uploads",
		},
		Storage: config.StorageConfig{
			Driver: "local",
		},
	}

	storage, err := newUploadStorageRepository(cfg)
	if err != nil {
		t.Fatalf("newUploadStorageRepository() error = %v", err)
	}
	if _, ok := storage.(*repository.LocalFileStorageRepository); !ok {
		t.Fatalf("storage type = %T, want *repository.LocalFileStorageRepository", storage)
	}
}

func TestNewUploadStorageRepositoryCreatesS3Repository(t *testing.T) {
	cfg := &config.Config{
		Upload: config.UploadConfig{
			PublicBaseURL: "",
		},
		Storage: config.StorageConfig{
			Driver: "s3",
			S3: config.S3StorageConfig{
				Bucket:          "media-bucket",
				Region:          "us-east-1",
				Endpoint:        "http://minio.local:9000",
				AccessKeyID:     "access",
				SecretAccessKey: "secret",
				Prefix:          "uploads",
				UseSSL:          false,
				ForcePathStyle:  true,
			},
		},
	}

	storage, err := newUploadStorageRepository(cfg)
	if err != nil {
		t.Fatalf("newUploadStorageRepository() error = %v", err)
	}
	if _, ok := storage.(*repository.S3FileStorageRepository); !ok {
		t.Fatalf("storage type = %T, want *repository.S3FileStorageRepository", storage)
	}
}

func TestNewUploadStorageRepositoryRejectsUnknownDriver(t *testing.T) {
	cfg := &config.Config{
		Storage: config.StorageConfig{
			Driver: "unknown",
		},
	}

	_, err := newUploadStorageRepository(cfg)
	if err == nil {
		t.Fatal("newUploadStorageRepository() error = nil, want error")
	}
}
