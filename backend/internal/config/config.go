package config

import (
	"errors"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	JWT       JWTConfig
	CORS      CORSConfig
	RateLimit RateLimitConfig
	Upload    UploadConfig
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
	MaxSize      int64  `mapstructure:"UPLOAD_MAX_SIZE"`
	Dir          string `mapstructure:"UPLOAD_DIR"`
	AllowedTypes string `mapstructure:"UPLOAD_ALLOWED_TYPES"`
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
	v.SetDefault("STORAGE_DRIVER", "local")
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
			MaxSize:      v.GetInt64("UPLOAD_MAX_SIZE"),
			Dir:          v.GetString("UPLOAD_DIR"),
			AllowedTypes: v.GetString("UPLOAD_ALLOWED_TYPES"),
		},
		Log: LogConfig{
			Level:  v.GetString("LOG_LEVEL"),
			Format: v.GetString("LOG_FORMAT"),
		},
	}

	return cfg, nil
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
	return strings.Split(c.Upload.AllowedTypes, ",")
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
