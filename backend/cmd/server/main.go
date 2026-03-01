package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/xxbbzy/gonext-template/backend/internal/config"
	"github.com/xxbbzy/gonext-template/backend/internal/handler"
	"github.com/xxbbzy/gonext-template/backend/internal/middleware"
	"github.com/xxbbzy/gonext-template/backend/internal/model"
	"github.com/xxbbzy/gonext-template/backend/internal/repository"
	"github.com/xxbbzy/gonext-template/backend/internal/service"
	pkgjwt "github.com/xxbbzy/gonext-template/backend/pkg/jwt"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := config.NewLoggerOrPanic(cfg)
	defer logger.Sync()

	logger.Info("Starting application",
		zap.String("name", cfg.App.Name),
		zap.String("env", cfg.App.Env),
	)

	// Initialize database
	db, err := config.NewDatabase(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Auto-migrate in development
	if cfg.IsDevelopment() {
		logger.Info("Running auto-migration (development mode)")
		if err := db.AutoMigrate(&model.User{}, &model.Item{}); err != nil {
			logger.Fatal("Failed to auto-migrate", zap.Error(err))
		}
	}

	// Initialize JWT manager
	jwtManager, err := pkgjwt.NewManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessExpiry,
		cfg.JWT.RefreshExpiry,
	)
	if err != nil {
		logger.Fatal("Failed to initialize JWT manager", zap.Error(err))
	}

	// Initialize layers (manual DI — Wire can replace this)
	userRepo := repository.NewUserRepository(db)
	itemRepo := repository.NewItemRepository(db)

	authService := service.NewAuthService(userRepo, jwtManager)
	itemService := service.NewItemService(itemRepo)

	authHandler := handler.NewAuthHandler(authService)
	itemHandler := handler.NewItemHandler(itemService)

	// File upload
	storage := handler.NewLocalStorage(cfg.Upload.Dir, fmt.Sprintf("http://localhost:%s", cfg.App.Port))
	uploadHandler := handler.NewUploadHandler(storage, cfg)

	// Rate limiter
	rateLimitDuration, _ := time.ParseDuration(cfg.RateLimit.Duration)
	if rateLimitDuration == 0 {
		rateLimitDuration = time.Minute
	}
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit.Requests, rateLimitDuration)

	// Setup Gin
	if !cfg.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	// Global middleware
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.RequestLogger(logger))
	r.Use(middleware.ErrorHandler())
	r.Use(rateLimiter.Middleware())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.GetAllowedOrigins(),
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Static file serving for uploads
	r.Static("/uploads", cfg.Upload.Dir)

	// Health check routes
	handler.RegisterHealthRoutes(r, func() bool {
		return checkDBHealth(db)
	})

	// API routes
	authMiddleware := middleware.Auth(jwtManager)
	v1 := r.Group("/api/v1")
	{
		authHandler.RegisterRoutes(v1, authMiddleware)
		itemHandler.RegisterRoutes(v1, authMiddleware)
		uploadHandler.RegisterRoutes(v1, authMiddleware)
	}

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.App.Port,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		logger.Info("Server starting", zap.String("port", cfg.App.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	// Close database
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}

	logger.Info("Server exited")
}

func checkDBHealth(db *gorm.DB) bool {
	sqlDB, err := db.DB()
	if err != nil {
		return false
	}
	return sqlDB.Ping() == nil
}
