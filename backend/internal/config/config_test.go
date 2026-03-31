package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadPrefersEnvironmentOverYAML(t *testing.T) {
	t.Setenv("APP_PORT", "7070")

	tempDir := setupTempWD(t)
	configYAML := []byte("APP_PORT: \"9090\"\nAPP_BASE_URL: \"http://yaml.example\"\n")
	if err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), configYAML, 0644); err != nil {
		t.Fatalf("write config.yaml: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.App.Port != "7070" {
		t.Fatalf("expected APP_PORT from env, got %q", cfg.App.Port)
	}
	if cfg.App.BaseURL != "http://yaml.example" {
		t.Fatalf("expected APP_BASE_URL from yaml, got %q", cfg.App.BaseURL)
	}
}

func TestLoadFailsFastOnInvalidConfig(t *testing.T) {
	t.Setenv("APP_ENV", "invalid-env")

	setupTempWD(t)

	_, err := Load()
	if err == nil {
		t.Fatalf("expected Load() to fail for invalid APP_ENV")
	}
	if !strings.Contains(err.Error(), "APP_ENV") {
		t.Fatalf("expected APP_ENV in error, got %q", err.Error())
	}
}

func TestConfigValidateAcceptsAllAllowedAppEnvs(t *testing.T) {
	allowedEnvs := []string{"development", "test", "staging", "production"}
	for _, env := range allowedEnvs {
		t.Run(env, func(t *testing.T) {
			cfg := newValidConfig()
			cfg.App.Env = env
			cfg.JWT.Secret = "a-strong-secret"

			if err := cfg.Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}
}

func TestConfigValidateNormalizesEnvAndDriver(t *testing.T) {
	cfg := newValidConfig()
	cfg.App.Env = " Production "
	cfg.Database.Driver = " SQLITE "
	cfg.Storage.Driver = " S3 "
	cfg.Storage.S3.Bucket = " test-bucket "
	cfg.Storage.S3.Region = " us-east-1 "
	cfg.Storage.S3.AccessKeyID = " test-access "
	cfg.Storage.S3.SecretAccessKey = " test-secret "
	cfg.JWT.Secret = "strong-secret"

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if cfg.App.Env != "production" {
		t.Fatalf("expected normalized APP_ENV to be %q, got %q", "production", cfg.App.Env)
	}
	if cfg.Database.Driver != "sqlite" {
		t.Fatalf("expected normalized DB_DRIVER to be %q, got %q", "sqlite", cfg.Database.Driver)
	}
	if cfg.Storage.Driver != "s3" {
		t.Fatalf("expected normalized STORAGE_DRIVER to be %q, got %q", "s3", cfg.Storage.Driver)
	}
	if cfg.Storage.S3.Bucket != "test-bucket" {
		t.Fatalf("expected normalized S3_BUCKET to be %q, got %q", "test-bucket", cfg.Storage.S3.Bucket)
	}
}

func TestConfigValidateRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name       string
		mutate     func(*Config)
		wantSubstr string
	}{
		{
			name: "invalid app env",
			mutate: func(cfg *Config) {
				cfg.App.Env = "qa"
			},
			wantSubstr: "APP_ENV",
		},
		{
			name: "invalid db driver",
			mutate: func(cfg *Config) {
				cfg.Database.Driver = "mysql"
			},
			wantSubstr: "DB_DRIVER",
		},
		{
			name: "empty jwt secret",
			mutate: func(cfg *Config) {
				cfg.JWT.Secret = "   "
			},
			wantSubstr: "JWT_SECRET must be non-empty",
		},
		{
			name: "production jwt placeholder rejected",
			mutate: func(cfg *Config) {
				cfg.App.Env = "production"
				cfg.JWT.Secret = "change-me-in-production"
			},
			wantSubstr: "JWT_SECRET must not use default or placeholder values in production",
		},
		{
			name: "rate limit requests must be positive",
			mutate: func(cfg *Config) {
				cfg.RateLimit.Requests = 0
			},
			wantSubstr: "RATE_LIMIT_REQUESTS",
		},
		{
			name: "rate limit duration must parse",
			mutate: func(cfg *Config) {
				cfg.RateLimit.Duration = "abc"
			},
			wantSubstr: "RATE_LIMIT_DURATION",
		},
		{
			name: "rate limit duration must be positive",
			mutate: func(cfg *Config) {
				cfg.RateLimit.Duration = "0s"
			},
			wantSubstr: "RATE_LIMIT_DURATION",
		},
		{
			name: "upload max size must be positive",
			mutate: func(cfg *Config) {
				cfg.Upload.MaxSize = -1
			},
			wantSubstr: "UPLOAD_MAX_SIZE",
		},
		{
			name: "upload dir must be non-empty",
			mutate: func(cfg *Config) {
				cfg.Upload.Dir = "  "
			},
			wantSubstr: "UPLOAD_DIR",
		},
		{
			name: "upload allowed types must be non-empty",
			mutate: func(cfg *Config) {
				cfg.Upload.AllowedTypes = " "
			},
			wantSubstr: "UPLOAD_ALLOWED_TYPES",
		},
		{
			name: "upload allowed types entries must start with dot",
			mutate: func(cfg *Config) {
				cfg.Upload.AllowedTypes = "jpg,.png"
			},
			wantSubstr: "UPLOAD_ALLOWED_TYPES",
		},
		{
			name: "upload allowed types entries must be simple extensions",
			mutate: func(cfg *Config) {
				cfg.Upload.AllowedTypes = ".jp-g,.png"
			},
			wantSubstr: "UPLOAD_ALLOWED_TYPES",
		},
		{
			name: "upload allowed types must have MIME compatibility support",
			mutate: func(cfg *Config) {
				cfg.Upload.AllowedTypes = ".png,.heic"
			},
			wantSubstr: "UPLOAD_ALLOWED_TYPES",
		},
		{
			name: "upload public base url must be parseable",
			mutate: func(cfg *Config) {
				cfg.Upload.PublicBaseURL = "://bad-url"
			},
			wantSubstr: "UPLOAD_PUBLIC_BASE_URL",
		},
		{
			name: "storage driver must be supported",
			mutate: func(cfg *Config) {
				cfg.Storage.Driver = "oss"
			},
			wantSubstr: "STORAGE_DRIVER",
		},
		{
			name: "s3 bucket required when storage driver is s3",
			mutate: func(cfg *Config) {
				cfg.Storage.Driver = "s3"
				cfg.Storage.S3.Bucket = ""
			},
			wantSubstr: "S3_BUCKET",
		},
		{
			name: "s3 region required when storage driver is s3",
			mutate: func(cfg *Config) {
				cfg.Storage.Driver = "s3"
				cfg.Storage.S3.Region = ""
			},
			wantSubstr: "S3_REGION",
		},
		{
			name: "s3 access key required when storage driver is s3",
			mutate: func(cfg *Config) {
				cfg.Storage.Driver = "s3"
				cfg.Storage.S3.AccessKeyID = ""
			},
			wantSubstr: "S3_ACCESS_KEY_ID",
		},
		{
			name: "s3 secret key required when storage driver is s3",
			mutate: func(cfg *Config) {
				cfg.Storage.Driver = "s3"
				cfg.Storage.S3.SecretAccessKey = ""
			},
			wantSubstr: "S3_SECRET_ACCESS_KEY",
		},
		{
			name: "s3 endpoint must be parseable when configured",
			mutate: func(cfg *Config) {
				cfg.Storage.Driver = "s3"
				cfg.Storage.S3.Endpoint = "://bad-endpoint"
			},
			wantSubstr: "S3_ENDPOINT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := newValidConfig()
			tt.mutate(cfg)

			err := cfg.Validate()
			if err == nil {
				t.Fatalf("expected Validate() to fail")
			}
			if !strings.Contains(err.Error(), tt.wantSubstr) {
				t.Fatalf("expected error to contain %q, got %q", tt.wantSubstr, err.Error())
			}
		})
	}
}

func TestGetAllowedFileTypesNormalizesWhitespaceAndCase(t *testing.T) {
	cfg := newValidConfig()
	cfg.Upload.AllowedTypes = " .JPG , .Pdf "

	got := cfg.GetAllowedFileTypes()
	want := []string{".jpg", ".pdf"}
	if len(got) != len(want) {
		t.Fatalf("len(GetAllowedFileTypes()) = %d, want %d (values: %v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("GetAllowedFileTypes()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestConfigValidateNormalizesUploadPublicBaseURL(t *testing.T) {
	cfg := newValidConfig()
	cfg.Upload.PublicBaseURL = " https://assets.example.com/ "

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if cfg.Upload.PublicBaseURL != "https://assets.example.com" {
		t.Fatalf(
			"Upload.PublicBaseURL = %q, want %q",
			cfg.Upload.PublicBaseURL,
			"https://assets.example.com",
		)
	}
}

func TestConfigValidateFallsBackToAppBaseURLForUploadPublicBaseURL(t *testing.T) {
	cfg := newValidConfig()
	cfg.App.BaseURL = " https://api.example.com/ "
	cfg.Upload.PublicBaseURL = " "

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if cfg.Upload.PublicBaseURL != "" {
		t.Fatalf(
			"Upload.PublicBaseURL = %q, want empty string",
			cfg.Upload.PublicBaseURL,
		)
	}
	if cfg.ResolvedUploadPublicBaseURL() != "https://api.example.com" {
		t.Fatalf(
			"ResolvedUploadPublicBaseURL() = %q, want %q",
			cfg.ResolvedUploadPublicBaseURL(),
			"https://api.example.com",
		)
	}
}

func TestConfigValidateNormalizesS3EndpointWithoutSchemeUsingUseSSL(t *testing.T) {
	cfg := newValidConfig()
	cfg.Storage.Driver = "s3"
	cfg.Storage.S3.UseSSL = false
	cfg.Storage.S3.Endpoint = "minio.local:9000/"

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if cfg.Storage.S3.Endpoint != "http://minio.local:9000" {
		t.Fatalf("Storage.S3.Endpoint = %q, want %q", cfg.Storage.S3.Endpoint, "http://minio.local:9000")
	}
}

func TestMetricsEnabledDefaultsToFalse(t *testing.T) {
	cfg := newValidConfig()

	if cfg.MetricsEnabled() {
		t.Fatal("MetricsEnabled() = true, want false by default")
	}
}

func TestMetricsEnabledReflectsObservabilityConfig(t *testing.T) {
	cfg := newValidConfig()
	cfg.Observability.MetricsEnabled = true

	if !cfg.MetricsEnabled() {
		t.Fatal("MetricsEnabled() = false, want true when enabled")
	}
}

func TestMetricsEnabledNilSafe(t *testing.T) {
	var cfg *Config

	if cfg.MetricsEnabled() {
		t.Fatal("MetricsEnabled() = true, want false for nil config")
	}
}

func newValidConfig() *Config {
	return &Config{
		App: AppConfig{
			Env:     "development",
			BaseURL: "http://localhost:8080",
		},
		Database: DatabaseConfig{
			Driver: "sqlite",
		},
		JWT: JWTConfig{
			Secret: "dev-secret-123",
		},
		RateLimit: RateLimitConfig{
			Requests: 100,
			Duration: "1m",
		},
		Upload: UploadConfig{
			MaxSize:       1024,
			Dir:           "./uploads",
			AllowedTypes:  ".jpg,.png",
			PublicBaseURL: "http://localhost:8080",
		},
		Storage: StorageConfig{
			Driver: "local",
			S3: S3StorageConfig{
				Bucket:          "bucket",
				Region:          "us-east-1",
				AccessKeyID:     "access",
				SecretAccessKey: "secret",
				UseSSL:          true,
			},
		},
		Observability: ObservabilityConfig{},
	}
}

func setupTempWD(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Chdir(previousDir)
	})

	return tempDir
}
