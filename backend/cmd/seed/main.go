package main

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/xxbbzy/gonext-template/backend/internal/config"
	"github.com/xxbbzy/gonext-template/backend/internal/model"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger := config.NewLoggerOrPanic(cfg)
	defer func() { _ = logger.Sync() }()

	db, err := config.NewDatabase(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	if err := db.AutoMigrate(&model.User{}, &model.Item{}); err != nil {
		logger.Fatal("Failed to auto-migrate", zap.Error(err))
	}

	adminPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	admin := model.User{
		Username:     "admin",
		Email:        "admin@example.com",
		PasswordHash: string(adminPassword),
		Role:         "admin",
	}
	if err := db.FirstOrCreate(&admin, model.User{Email: "admin@example.com"}).Error; err != nil {
		logger.Error("Failed to seed admin user", zap.Error(err))
	} else {
		logger.Info("Seeded admin user", zap.String("email", admin.Email))
	}

	userPassword, _ := bcrypt.GenerateFromPassword([]byte("user123"), bcrypt.DefaultCost)
	user := model.User{
		Username:     "testuser",
		Email:        "user@example.com",
		PasswordHash: string(userPassword),
		Role:         "user",
	}
	if err := db.FirstOrCreate(&user, model.User{Email: "user@example.com"}).Error; err != nil {
		logger.Error("Failed to seed regular user", zap.Error(err))
	} else {
		logger.Info("Seeded regular user", zap.String("email", user.Email))
	}

	items := []model.Item{
		{Title: "示例任务 1", Description: "这是第一个示例任务", Status: "active", UserID: admin.ID},
		{Title: "示例任务 2", Description: "这是第二个示例任务", Status: "active", UserID: admin.ID},
		{Title: "示例任务 3", Description: "这是第三个示例任务", Status: "inactive", UserID: user.ID},
		{Title: "Sample Task 4", Description: "This is a sample task in English", Status: "active", UserID: user.ID},
		{Title: "Sample Task 5", Description: "Another sample task for testing", Status: "active", UserID: admin.ID},
	}

	for _, item := range items {
		if err := db.FirstOrCreate(&item, model.Item{Title: item.Title}).Error; err != nil {
			logger.Error("Failed to seed item", zap.String("title", item.Title), zap.Error(err))
		}
	}

	logger.Info("Seed data created successfully")
}
