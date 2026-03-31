//go:build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/xxbbzy/gonext-template/backend/internal/config"
	"github.com/xxbbzy/gonext-template/backend/internal/handler"
	"github.com/xxbbzy/gonext-template/backend/internal/repository"
)

//go:generate go run github.com/google/wire/cmd/wire

func InitializeApplication() (*Application, error) {
	wire.Build(
		config.Load,
		config.NewLogger,
		config.NewDatabase,
		newJWTManager,
		repository.NewUserRepository,
		repository.NewItemRepository,
		newAuthService,
		newItemService,
		newUploadStorageRepository,
		newUploadService,
		handler.NewAuthHandler,
		handler.NewItemHandler,
		handler.NewUploadHandler,
		newPublicRateLimiter,
		newUserRateLimiter,
		newHTTPMetrics,
		newPrometheusRegistry,
		newApplication,
	)
	return nil, nil
}
