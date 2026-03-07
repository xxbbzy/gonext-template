package main

import (
	"fmt"
	"os"

	"github.com/xxbbzy/gonext-template/backend/internal/config"
	"github.com/xxbbzy/gonext-template/backend/internal/model"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger, err := config.NewLogger(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	db, err := config.NewDatabase(cfg, logger)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}

	if err := db.AutoMigrate(&model.User{}, &model.Item{}); err != nil {
		logger.Fatal("failed to bootstrap schema", zap.Error(err))
	}

	logger.Info("Development schema bootstrapped")
}
