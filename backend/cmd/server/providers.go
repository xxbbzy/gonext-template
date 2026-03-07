package main

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/xxbbzy/gonext-template/backend/internal/config"
	"github.com/xxbbzy/gonext-template/backend/internal/handler"
	"github.com/xxbbzy/gonext-template/backend/internal/middleware"
	"github.com/xxbbzy/gonext-template/backend/internal/repository"
	"github.com/xxbbzy/gonext-template/backend/internal/service"
	pkgjwt "github.com/xxbbzy/gonext-template/backend/pkg/jwt"
)

type Application struct {
	Config            *config.Config
	Logger            *zap.Logger
	DB                *gorm.DB
	JWTManager        *pkgjwt.Manager
	AuthHandler       *handler.AuthHandler
	ItemHandler       *handler.ItemHandler
	UploadHandler     *handler.UploadHandler
	PublicRateLimiter *middleware.RateLimiter
	UserRateLimiter   *middleware.RateLimiter
}

func newJWTManager(cfg *config.Config) (*pkgjwt.Manager, error) {
	return pkgjwt.NewManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessExpiry,
		cfg.JWT.RefreshExpiry,
	)
}

func newUploadStorage(cfg *config.Config) handler.Storage {
	return handler.NewLocalStorage(cfg.Upload.Dir, cfg.App.BaseURL)
}

func newPublicRateLimiter(cfg *config.Config) *middleware.RateLimiter {
	return middleware.NewRateLimiter(cfg.RateLimit.Requests, parseRateLimitDuration(cfg))
}

func newUserRateLimiter(cfg *config.Config) *middleware.RateLimiter {
	return middleware.NewRateLimiter(cfg.RateLimit.Requests, parseRateLimitDuration(cfg))
}

func parseRateLimitDuration(cfg *config.Config) time.Duration {
	duration, err := time.ParseDuration(cfg.RateLimit.Duration)
	if err != nil || duration <= 0 {
		return time.Minute
	}
	return duration
}

func newAuthService(userRepo *repository.UserRepository, jwtManager *pkgjwt.Manager) *service.AuthService {
	return service.NewAuthService(userRepo, jwtManager)
}

func newItemService(itemRepo *repository.ItemRepository) *service.ItemService {
	return service.NewItemService(itemRepo)
}

func newApplication(
	cfg *config.Config,
	logger *zap.Logger,
	db *gorm.DB,
	jwtManager *pkgjwt.Manager,
	authHandler *handler.AuthHandler,
	itemHandler *handler.ItemHandler,
	uploadHandler *handler.UploadHandler,
	publicRateLimiter *middleware.RateLimiter,
	userRateLimiter *middleware.RateLimiter,
) *Application {
	return &Application{
		Config:            cfg,
		Logger:            logger,
		DB:                db,
		JWTManager:        jwtManager,
		AuthHandler:       authHandler,
		ItemHandler:       itemHandler,
		UploadHandler:     uploadHandler,
		PublicRateLimiter: publicRateLimiter,
		UserRateLimiter:   userRateLimiter,
	}
}
