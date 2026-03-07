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
	"github.com/xxbbzy/gonext-template/backend/internal/handler"
	"github.com/xxbbzy/gonext-template/backend/internal/middleware"
	"github.com/xxbbzy/gonext-template/backend/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	app, err := InitializeApplication()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize application: %v\n", err)
		os.Exit(1)
	}
	defer app.Logger.Sync()

	app.Logger.Info("Starting application",
		zap.String("name", app.Config.App.Name),
		zap.String("env", app.Config.App.Env),
	)

	// Auto-migrate in development
	if app.Config.IsDevelopment() {
		app.Logger.Info("Running auto-migration (development mode)")
		if err := app.DB.AutoMigrate(&model.User{}, &model.Item{}); err != nil {
			app.Logger.Fatal("Failed to auto-migrate", zap.Error(err))
		}
	}

	// Setup Gin
	if app.Config.App.GinMode != "" {
		gin.SetMode(app.Config.App.GinMode)
	} else if !app.Config.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	// Global middleware
	r.Use(middleware.Recovery(app.Logger))
	r.Use(middleware.RequestLogger(app.Logger))
	r.Use(middleware.ErrorHandler())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     app.Config.GetAllowedOrigins(),
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Static file serving for uploads
	r.Static("/uploads", app.Config.Upload.Dir)
	r.Static("/swagger", "./docs")
	r.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, "/swagger/index.html")
	})

	// Health check routes
	handler.RegisterHealthRoutes(r, func() bool {
		return checkDBHealth(app.DB)
	})

	// API routes
	publicRateLimit := app.PublicRateLimiter.Middleware()
	authMiddleware := middleware.Auth(app.JWTManager)
	userRateLimit := app.UserRateLimiter.MiddlewareWithKey(middleware.UserKey)
	v1 := r.Group("/api/v1")
	{
		app.AuthHandler.RegisterRoutes(v1, []gin.HandlerFunc{publicRateLimit}, authMiddleware, userRateLimit)
		app.ItemHandler.RegisterRoutes(v1, authMiddleware, userRateLimit)
		app.UploadHandler.RegisterRoutes(v1, authMiddleware, userRateLimit)
	}

	// Start server
	srv := &http.Server{
		Addr:    ":" + app.Config.App.Port,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		app.Logger.Info("Server starting", zap.String("port", app.Config.App.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.Logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	app.Logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		app.Logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	// Close database
	sqlDB, err := app.DB.DB()
	if err == nil {
		sqlDB.Close()
	}

	app.Logger.Info("Server exited")
}

func checkDBHealth(db *gorm.DB) bool {
	sqlDB, err := db.DB()
	if err != nil {
		return false
	}
	return sqlDB.Ping() == nil
}
