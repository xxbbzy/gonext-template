package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	CORS     CORSConfig
	RateLimit RateLimitConfig
	Upload   UploadConfig
	Log      LogConfig
}

type AppConfig struct {
	Name string `mapstructure:"APP_NAME"`
	Env  string `mapstructure:"APP_ENV"`
	Port string `mapstructure:"APP_PORT"`
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
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("APP_NAME", "gonext-template")
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("DB_DRIVER", "sqlite")
	viper.SetDefault("DB_DSN", "./data/app.db")
	viper.SetDefault("JWT_SECRET", "change-me-in-production")
	viper.SetDefault("JWT_ACCESS_EXPIRY", "15m")
	viper.SetDefault("JWT_REFRESH_EXPIRY", "168h")
	viper.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:3000")
	viper.SetDefault("RATE_LIMIT_REQUESTS", 100)
	viper.SetDefault("RATE_LIMIT_DURATION", "1m")
	viper.SetDefault("UPLOAD_MAX_SIZE", 10485760)
	viper.SetDefault("UPLOAD_DIR", "./uploads")
	viper.SetDefault("UPLOAD_ALLOWED_TYPES", ".jpg,.jpeg,.png,.gif,.pdf,.doc,.docx")
	viper.SetDefault("STORAGE_DRIVER", "local")
	viper.SetDefault("LOG_LEVEL", "debug")
	viper.SetDefault("LOG_FORMAT", "json")

	// Read .env file (ignore error if not exists)
	_ = viper.ReadInConfig()

	cfg := &Config{
		App: AppConfig{
			Name: viper.GetString("APP_NAME"),
			Env:  viper.GetString("APP_ENV"),
			Port: viper.GetString("APP_PORT"),
		},
		Database: DatabaseConfig{
			Driver: viper.GetString("DB_DRIVER"),
			DSN:    viper.GetString("DB_DSN"),
		},
		JWT: JWTConfig{
			Secret:        viper.GetString("JWT_SECRET"),
			AccessExpiry:  viper.GetString("JWT_ACCESS_EXPIRY"),
			RefreshExpiry: viper.GetString("JWT_REFRESH_EXPIRY"),
		},
		CORS: CORSConfig{
			AllowedOrigins: viper.GetString("CORS_ALLOWED_ORIGINS"),
		},
		RateLimit: RateLimitConfig{
			Requests: viper.GetInt("RATE_LIMIT_REQUESTS"),
			Duration: viper.GetString("RATE_LIMIT_DURATION"),
		},
		Upload: UploadConfig{
			MaxSize:      viper.GetInt64("UPLOAD_MAX_SIZE"),
			Dir:          viper.GetString("UPLOAD_DIR"),
			AllowedTypes: viper.GetString("UPLOAD_ALLOWED_TYPES"),
		},
		Log: LogConfig{
			Level:  viper.GetString("LOG_LEVEL"),
			Format: viper.GetString("LOG_FORMAT"),
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
