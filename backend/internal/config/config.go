package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"

	uploadvalidation "github.com/xxbbzy/gonext-template/backend/internal/upload"
)

var (
	allowedAppEnvs = map[string]struct{}{
		"development": {},
		"test":        {},
		"staging":     {},
		"production":  {},
	}

	allowedDBDrivers = map[string]struct{}{
		"sqlite":   {},
		"postgres": {},
	}

	allowedStorageDrivers = map[string]struct{}{
		"local": {},
		"s3":    {},
	}

	disallowedProductionJWTSecrets = map[string]struct{}{
		"change-me-in-production":    {},
		"changeme":                   {},
		"your-jwt-secret":            {},
		"replace-me":                 {},
		"replace-with-strong-secret": {},
		"placeholder-secret":         {},
	}
)

// Config holds all application configuration.
type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	JWT       JWTConfig
	CORS      CORSConfig
	RateLimit RateLimitConfig
	Upload    UploadConfig
	Storage   StorageConfig
	Log       LogConfig
}

type AppConfig struct {
	Name    string `mapstructure:"APP_NAME"`
	Env     string `mapstructure:"APP_ENV"`
	GinMode string `mapstructure:"GIN_MODE"`
	Port    string `mapstructure:"APP_PORT"`
	BaseURL string `mapstructure:"APP_BASE_URL"`
}

type DatabaseConfig struct {
	Driver string `mapstructure:"DB_DRIVER"`
	DSN    string `mapstructure:"DB_DSN"`
}

type JWTConfig struct {
	Secret        string `mapstructure:"JWT_SECRET"`
	AccessExpiry  string `mapstructure:"JWT_ACCESS_EXPIRY"`
	RefreshExpiry string `mapstructure:"JWT_REFRESH_EXPIRY"`
}

type CORSConfig struct {
	AllowedOrigins string `mapstructure:"CORS_ALLOWED_ORIGINS"`
}

type RateLimitConfig struct {
	Requests int    `mapstructure:"RATE_LIMIT_REQUESTS"`
	Duration string `mapstructure:"RATE_LIMIT_DURATION"`
}

type UploadConfig struct {
	MaxSize       int64  `mapstructure:"UPLOAD_MAX_SIZE"`
	Dir           string `mapstructure:"UPLOAD_DIR"`
	AllowedTypes  string `mapstructure:"UPLOAD_ALLOWED_TYPES"`
	PublicBaseURL string `mapstructure:"UPLOAD_PUBLIC_BASE_URL"`
}

type StorageConfig struct {
	Driver string          `mapstructure:"STORAGE_DRIVER"`
	S3     S3StorageConfig `mapstructure:",squash"`
}

type S3StorageConfig struct {
	Bucket          string `mapstructure:"S3_BUCKET"`
	Region          string `mapstructure:"S3_REGION"`
	Endpoint        string `mapstructure:"S3_ENDPOINT"`
	AccessKeyID     string `mapstructure:"S3_ACCESS_KEY_ID"`
	SecretAccessKey string `mapstructure:"S3_SECRET_ACCESS_KEY"`
	Prefix          string `mapstructure:"S3_PREFIX"`
	UseSSL          bool   `mapstructure:"S3_USE_SSL"`
	ForcePathStyle  bool   `mapstructure:"S3_FORCE_PATH_STYLE"`
}

type LogConfig struct {
	Level  string `mapstructure:"LOG_LEVEL"`
	Format string `mapstructure:"LOG_FORMAT"`
}

// Load reads configuration from .env file and environment variables.
func Load() (*Config, error) {
	v := viper.New()
	v.AutomaticEnv()

	// Set defaults
	v.SetDefault("APP_NAME", "gonext-template")
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("GIN_MODE", "release")
	v.SetDefault("APP_PORT", "8080")
	v.SetDefault("APP_BASE_URL", "http://localhost:8080")
	v.SetDefault("DB_DRIVER", "sqlite")
	v.SetDefault("DB_DSN", "./data/app.db")
	v.SetDefault("JWT_SECRET", "change-me-in-production")
	v.SetDefault("JWT_ACCESS_EXPIRY", "15m")
	v.SetDefault("JWT_REFRESH_EXPIRY", "168h")
	v.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:3000")
	v.SetDefault("RATE_LIMIT_REQUESTS", 100)
	v.SetDefault("RATE_LIMIT_DURATION", "1m")
	v.SetDefault("UPLOAD_MAX_SIZE", 10485760)
	v.SetDefault("UPLOAD_DIR", "./uploads")
	v.SetDefault("UPLOAD_ALLOWED_TYPES", ".jpg,.jpeg,.png,.gif,.pdf,.doc,.docx")
	v.SetDefault("UPLOAD_PUBLIC_BASE_URL", "")
	v.SetDefault("STORAGE_DRIVER", "local")
	v.SetDefault("S3_BUCKET", "")
	v.SetDefault("S3_REGION", "")
	v.SetDefault("S3_ENDPOINT", "")
	v.SetDefault("S3_ACCESS_KEY_ID", "")
	v.SetDefault("S3_SECRET_ACCESS_KEY", "")
	v.SetDefault("S3_PREFIX", "")
	v.SetDefault("S3_USE_SSL", true)
	v.SetDefault("S3_FORCE_PATH_STYLE", false)
	v.SetDefault("LOG_LEVEL", "debug")
	v.SetDefault("LOG_FORMAT", "json")

	loadOptionalConfig(v, "config.yaml")
	loadOptionalConfig(v, "config.yml")
	loadOptionalConfig(v, ".env")

	cfg := &Config{
		App: AppConfig{
			Name:    v.GetString("APP_NAME"),
			Env:     v.GetString("APP_ENV"),
			GinMode: v.GetString("GIN_MODE"),
			Port:    v.GetString("APP_PORT"),
			BaseURL: v.GetString("APP_BASE_URL"),
		},
		Database: DatabaseConfig{
			Driver: v.GetString("DB_DRIVER"),
			DSN:    v.GetString("DB_DSN"),
		},
		JWT: JWTConfig{
			Secret:        v.GetString("JWT_SECRET"),
			AccessExpiry:  v.GetString("JWT_ACCESS_EXPIRY"),
			RefreshExpiry: v.GetString("JWT_REFRESH_EXPIRY"),
		},
		CORS: CORSConfig{
			AllowedOrigins: v.GetString("CORS_ALLOWED_ORIGINS"),
		},
		RateLimit: RateLimitConfig{
			Requests: v.GetInt("RATE_LIMIT_REQUESTS"),
			Duration: v.GetString("RATE_LIMIT_DURATION"),
		},
		Upload: UploadConfig{
			MaxSize:       v.GetInt64("UPLOAD_MAX_SIZE"),
			Dir:           v.GetString("UPLOAD_DIR"),
			AllowedTypes:  v.GetString("UPLOAD_ALLOWED_TYPES"),
			PublicBaseURL: v.GetString("UPLOAD_PUBLIC_BASE_URL"),
		},
		Storage: StorageConfig{
			Driver: v.GetString("STORAGE_DRIVER"),
			S3: S3StorageConfig{
				Bucket:          v.GetString("S3_BUCKET"),
				Region:          v.GetString("S3_REGION"),
				Endpoint:        v.GetString("S3_ENDPOINT"),
				AccessKeyID:     v.GetString("S3_ACCESS_KEY_ID"),
				SecretAccessKey: v.GetString("S3_SECRET_ACCESS_KEY"),
				Prefix:          v.GetString("S3_PREFIX"),
				UseSSL:          v.GetBool("S3_USE_SSL"),
				ForcePathStyle:  v.GetBool("S3_FORCE_PATH_STYLE"),
			},
		},
		Log: LogConfig{
			Level:  v.GetString("LOG_LEVEL"),
			Format: v.GetString("LOG_FORMAT"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate enforces startup-critical configuration constraints.
func (c *Config) Validate() error {
	var validationErrors []string

	env := strings.ToLower(strings.TrimSpace(c.App.Env))
	if _, ok := allowedAppEnvs[env]; !ok {
		validationErrors = append(validationErrors, "APP_ENV must be one of: development, test, staging, production")
	}

	driver := strings.ToLower(strings.TrimSpace(c.Database.Driver))
	if _, ok := allowedDBDrivers[driver]; !ok {
		validationErrors = append(validationErrors, "DB_DRIVER must be one of: sqlite, postgres")
	}

	secret := strings.TrimSpace(c.JWT.Secret)
	if secret == "" {
		validationErrors = append(validationErrors, "JWT_SECRET must be non-empty")
	}
	if env == "production" && secret != "" {
		if _, blocked := disallowedProductionJWTSecrets[strings.ToLower(secret)]; blocked {
			validationErrors = append(validationErrors, "JWT_SECRET must not use default or placeholder values in production")
		}
	}

	if c.RateLimit.Requests <= 0 {
		validationErrors = append(validationErrors, "RATE_LIMIT_REQUESTS must be greater than 0")
	}

	rateLimitDuration := strings.TrimSpace(c.RateLimit.Duration)
	duration, err := time.ParseDuration(rateLimitDuration)
	if err != nil || duration <= 0 {
		validationErrors = append(validationErrors, "RATE_LIMIT_DURATION must be a parseable positive duration (for example: 1m, 30s)")
	}

	if c.Upload.MaxSize <= 0 {
		validationErrors = append(validationErrors, "UPLOAD_MAX_SIZE must be greater than 0")
	}

	storageDriver := strings.ToLower(strings.TrimSpace(c.Storage.Driver))
	if _, ok := allowedStorageDrivers[storageDriver]; !ok {
		validationErrors = append(validationErrors, "STORAGE_DRIVER must be one of: local, s3")
	}

	if storageDriver == "local" && strings.TrimSpace(c.Upload.Dir) == "" {
		validationErrors = append(validationErrors, "UPLOAD_DIR must be non-empty when STORAGE_DRIVER=local")
	}

	if _, err := parseAllowedUploadTypes(c.Upload.AllowedTypes); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("UPLOAD_ALLOWED_TYPES %s", err.Error()))
	} else if err := uploadvalidation.ValidateSupportedExtensions(c.GetAllowedFileTypes()); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("UPLOAD_ALLOWED_TYPES %s", err.Error()))
	}

	normalizedUploadBaseURL := ""
	if strings.TrimSpace(c.Upload.PublicBaseURL) != "" {
		parsedUploadBaseURL, err := parseUploadPublicBaseURL(c.Upload.PublicBaseURL)
		if err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("UPLOAD_PUBLIC_BASE_URL %s", err.Error()))
		} else {
			normalizedUploadBaseURL = parsedUploadBaseURL
		}
	}

	normalizedS3Endpoint := ""
	if storageDriver == "s3" {
		if strings.TrimSpace(c.Storage.S3.Bucket) == "" {
			validationErrors = append(validationErrors, "S3_BUCKET must be non-empty when STORAGE_DRIVER=s3")
		}
		if strings.TrimSpace(c.Storage.S3.Region) == "" {
			validationErrors = append(validationErrors, "S3_REGION must be non-empty when STORAGE_DRIVER=s3")
		}
		if strings.TrimSpace(c.Storage.S3.AccessKeyID) == "" {
			validationErrors = append(validationErrors, "S3_ACCESS_KEY_ID must be non-empty when STORAGE_DRIVER=s3")
		}
		if strings.TrimSpace(c.Storage.S3.SecretAccessKey) == "" {
			validationErrors = append(validationErrors, "S3_SECRET_ACCESS_KEY must be non-empty when STORAGE_DRIVER=s3")
		}

		if strings.TrimSpace(c.Storage.S3.Endpoint) != "" {
			endpoint, err := parseS3Endpoint(c.Storage.S3.Endpoint, c.Storage.S3.UseSSL)
			if err != nil {
				validationErrors = append(validationErrors, fmt.Sprintf("S3_ENDPOINT %s", err.Error()))
			} else {
				normalizedS3Endpoint = endpoint
			}
		}
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("config validation failed: %s", strings.Join(validationErrors, "; "))
	}

	c.App.Env = env
	c.App.BaseURL = strings.TrimRight(strings.TrimSpace(c.App.BaseURL), "/")
	c.Database.Driver = driver
	c.Upload.PublicBaseURL = normalizedUploadBaseURL
	c.Upload.Dir = strings.TrimSpace(c.Upload.Dir)
	c.Storage.Driver = storageDriver
	c.Storage.S3.Bucket = strings.TrimSpace(c.Storage.S3.Bucket)
	c.Storage.S3.Region = strings.TrimSpace(c.Storage.S3.Region)
	c.Storage.S3.Endpoint = normalizedS3Endpoint
	c.Storage.S3.AccessKeyID = strings.TrimSpace(c.Storage.S3.AccessKeyID)
	c.Storage.S3.SecretAccessKey = strings.TrimSpace(c.Storage.S3.SecretAccessKey)
	c.Storage.S3.Prefix = strings.Trim(c.Storage.S3.Prefix, "/ ")

	return nil
}

// IsDevelopment returns true if the app is running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

// GetAllowedOrigins returns CORS allowed origins as a slice.
func (c *Config) GetAllowedOrigins() []string {
	return strings.Split(c.CORS.AllowedOrigins, ",")
}

// GetAllowedFileTypes returns upload allowed file types as a slice.
func (c *Config) GetAllowedFileTypes() []string {
	allowedTypes, err := parseAllowedUploadTypes(c.Upload.AllowedTypes)
	if err == nil {
		return allowedTypes
	}

	parts := strings.Split(c.Upload.AllowedTypes, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.ToLower(strings.TrimSpace(part))
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// ResolvedUploadPublicBaseURL returns UPLOAD_PUBLIC_BASE_URL, or APP_BASE_URL when unset.
func (c *Config) ResolvedUploadPublicBaseURL() string {
	if strings.TrimSpace(c.Upload.PublicBaseURL) != "" {
		return c.Upload.PublicBaseURL
	}
	return c.App.BaseURL
}

// UsesLocalUploadStorage returns true when local file storage is selected.
func (c *Config) UsesLocalUploadStorage() bool {
	return c.Storage.Driver == "local"
}

func loadOptionalConfig(v *viper.Viper, path string) {
	v.SetConfigFile(path)
	if err := v.MergeInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) {
			return
		}
	}
}

func parseAllowedUploadTypes(raw string) ([]string, error) {
	trimmedRaw := strings.TrimSpace(raw)
	if trimmedRaw == "" {
		return nil, fmt.Errorf("must be a non-empty comma-separated list like .jpg,.png")
	}

	parts := strings.Split(trimmedRaw, ",")
	types := make([]string, 0, len(parts))
	for i, part := range parts {
		trimmed := strings.ToLower(strings.TrimSpace(part))
		if trimmed == "" {
			return nil, fmt.Errorf("contains an empty extension entry at position %d", i+1)
		}
		if !strings.HasPrefix(trimmed, ".") || len(trimmed) == 1 {
			return nil, fmt.Errorf("entry %q must start with '.' and include extension characters", trimmed)
		}
		for _, ch := range trimmed[1:] {
			if (ch < 'a' || ch > 'z') && (ch < '0' || ch > '9') {
				return nil, fmt.Errorf("entry %q must use format .ext with lowercase letters and numbers", trimmed)
			}
		}
		types = append(types, trimmed)
	}

	return types, nil
}

func parseUploadPublicBaseURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("must be non-empty")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("must be a valid URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("must use http or https")
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("must include host")
	}

	return strings.TrimRight(parsed.String(), "/"), nil
}

func parseS3Endpoint(raw string, useSSL bool) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("must be non-empty")
	}

	if !strings.Contains(trimmed, "://") {
		scheme := "http"
		if useSSL {
			scheme = "https"
		}
		trimmed = scheme + "://" + trimmed
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("must be a valid URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("must use http or https")
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("must include host")
	}

	return strings.TrimRight(parsed.String(), "/"), nil
}
