package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/xxbbzy/gonext-template/backend/internal/config"
	"github.com/xxbbzy/gonext-template/backend/internal/handler"
	"github.com/xxbbzy/gonext-template/backend/internal/middleware"
	"github.com/xxbbzy/gonext-template/backend/internal/observability"
	"github.com/xxbbzy/gonext-template/backend/internal/repository"
	"github.com/xxbbzy/gonext-template/backend/internal/service"
	pkgjwt "github.com/xxbbzy/gonext-template/backend/pkg/jwt"
)

type Application struct {
	Config            *config.Config
	Logger            *zap.Logger
	DB                *gorm.DB
	JWTManager        *pkgjwt.Manager
	AuthService       *service.AuthService
	ItemService       *service.ItemService
	UploadService     *service.UploadService
	AuthHandler       *handler.AuthHandler
	ItemHandler       *handler.ItemHandler
	UploadHandler     *handler.UploadHandler
	PublicRateLimiter *middleware.RateLimiter
	UserRateLimiter   *middleware.RateLimiter
	HTTPMetrics       *middleware.HTTPMetrics
	MetricsRegistry   *prometheus.Registry
}

func newJWTManager(cfg *config.Config) (*pkgjwt.Manager, error) {
	return pkgjwt.NewManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessExpiry,
		cfg.JWT.RefreshExpiry,
	)
}

func newUploadStorageRepository(cfg *config.Config, logger *zap.Logger) (repository.FileStorageRepository, error) {
	switch cfg.Storage.Driver {
	case "local":
		return repository.NewLocalFileStorageRepository(cfg.Upload.Dir, cfg.ResolvedUploadPublicBaseURL())
	case "s3":
		return repository.NewS3FileStorageRepository(context.Background(), repository.S3FileStorageConfig{
			Bucket:          cfg.Storage.S3.Bucket,
			Region:          cfg.Storage.S3.Region,
			Endpoint:        cfg.Storage.S3.Endpoint,
			AccessKeyID:     cfg.Storage.S3.AccessKeyID,
			SecretAccessKey: cfg.Storage.S3.SecretAccessKey,
			Prefix:          cfg.Storage.S3.Prefix,
			UseSSL:          cfg.Storage.S3.UseSSL,
			ForcePathStyle:  cfg.Storage.S3.ForcePathStyle,
			PublicBaseURL:   cfg.ResolvedUploadPublicBaseURL(),
			Logger:          logger,
		})
	default:
		return nil, fmt.Errorf("unsupported storage driver %q", cfg.Storage.Driver)
	}
}

func newPublicRateLimiter(cfg *config.Config) *middleware.RateLimiter {
	return middleware.NewRateLimiter(cfg.RateLimit.Requests, mustParseRateLimitDuration(cfg))
}

func newUserRateLimiter(cfg *config.Config) *middleware.RateLimiter {
	return middleware.NewRateLimiter(cfg.RateLimit.Requests, mustParseRateLimitDuration(cfg))
}

func newHTTPMetrics(cfg *config.Config) *middleware.HTTPMetrics {
	if cfg == nil || !cfg.MetricsEnabled() {
		return nil
	}
	return middleware.NewHTTPMetrics(middleware.HTTPMetricsOptions{})
}

func newPrometheusRegistry(cfg *config.Config, httpMetrics *middleware.HTTPMetrics) (*prometheus.Registry, error) {
	if cfg == nil || !cfg.MetricsEnabled() {
		return nil, nil
	}
	if httpMetrics == nil {
		return nil, errors.New("http metrics collector is required when metrics are enabled")
	}
	return observability.NewPrometheusRegistry(observability.RegistryOptions{
		IncludeRuntimeCollectors: true,
		ApplicationCollectors:    httpMetrics.Collectors(),
	})
}

func mustParseRateLimitDuration(cfg *config.Config) time.Duration {
	duration, err := time.ParseDuration(cfg.RateLimit.Duration)
	if err != nil || duration <= 0 {
		panic(fmt.Sprintf("invalid RATE_LIMIT_DURATION after validation: %q", cfg.RateLimit.Duration))
	}
	return duration
}

func newAuthService(userRepo *repository.UserRepository, jwtManager *pkgjwt.Manager) *service.AuthService {
	return service.NewAuthService(userRepo, jwtManager)
}

func newItemService(itemRepo *repository.ItemRepository) *service.ItemService {
	return service.NewItemService(itemRepo)
}

func newUploadService(fileStorage repository.FileStorageRepository, logger *zap.Logger) *service.UploadService {
	return service.NewUploadService(fileStorage, logger)
}

func newApplication(
	cfg *config.Config,
	logger *zap.Logger,
	db *gorm.DB,
	jwtManager *pkgjwt.Manager,
	authService *service.AuthService,
	itemService *service.ItemService,
	uploadService *service.UploadService,
	authHandler *handler.AuthHandler,
	itemHandler *handler.ItemHandler,
	uploadHandler *handler.UploadHandler,
	publicRateLimiter *middleware.RateLimiter,
	userRateLimiter *middleware.RateLimiter,
	httpMetrics *middleware.HTTPMetrics,
	metricsRegistry *prometheus.Registry,
) *Application {
	return &Application{
		Config:            cfg,
		Logger:            logger,
		DB:                db,
		JWTManager:        jwtManager,
		AuthService:       authService,
		ItemService:       itemService,
		UploadService:     uploadService,
		AuthHandler:       authHandler,
		ItemHandler:       itemHandler,
		UploadHandler:     uploadHandler,
		PublicRateLimiter: publicRateLimiter,
		UserRateLimiter:   userRateLimiter,
		HTTPMetrics:       httpMetrics,
		MetricsRegistry:   metricsRegistry,
	}
}
