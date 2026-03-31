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

func TestNewHTTPMetricsDisabledByDefault(t *testing.T) {
	cfg := &config.Config{}

	if got := newHTTPMetrics(cfg); got != nil {
		t.Fatal("newHTTPMetrics() returned collector when metrics are disabled")
	}
}

func TestNewHTTPMetricsEnabled(t *testing.T) {
	cfg := &config.Config{
		Observability: config.ObservabilityConfig{
			MetricsEnabled: true,
		},
	}

	if got := newHTTPMetrics(cfg); got == nil {
		t.Fatal("newHTTPMetrics() = nil, want collector when metrics are enabled")
	}
}

func TestNewPrometheusRegistryDisabledWhenMetricsOff(t *testing.T) {
	cfg := &config.Config{}

	registry, err := newPrometheusRegistry(cfg, nil)
	if err != nil {
		t.Fatalf("newPrometheusRegistry() error = %v", err)
	}
	if registry != nil {
		t.Fatal("newPrometheusRegistry() returned registry when metrics are disabled")
	}
}

func TestNewPrometheusRegistryEnabledWhenMetricsOn(t *testing.T) {
	cfg := &config.Config{
		Observability: config.ObservabilityConfig{
			MetricsEnabled: true,
		},
	}

	httpMetrics := newHTTPMetrics(cfg)
	registry, err := newPrometheusRegistry(cfg, httpMetrics)
	if err != nil {
		t.Fatalf("newPrometheusRegistry() error = %v", err)
	}
	if registry == nil {
		t.Fatal("newPrometheusRegistry() = nil, want registry when metrics are enabled")
	}
}
