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
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"gorm.io/gorm"

	genapi "github.com/xxbbzy/gonext-template/backend/internal/api"
	"github.com/xxbbzy/gonext-template/backend/internal/handler"
	"github.com/xxbbzy/gonext-template/backend/internal/middleware"
	"github.com/xxbbzy/gonext-template/backend/internal/model"
	"github.com/xxbbzy/gonext-template/backend/pkg/errcode"
	"github.com/xxbbzy/gonext-template/backend/pkg/response"
)

func main() {
	app, err := InitializeApplication()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize application: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = app.Logger.Sync() }()

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
	r.Use(middleware.RequestID())
	if app.Config.MetricsEnabled() {
		r.Use(app.HTTPMetrics.Middleware())
	}
	r.Use(middleware.Recovery(app.Logger))
	r.Use(middleware.RequestLogger(app.Logger))
	r.Use(middleware.ErrorHandler())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     app.Config.GetAllowedOrigins(),
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", middleware.RequestIDHeader},
		ExposeHeaders:    []string{"Content-Length", middleware.RequestIDHeader},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Static file serving for uploads
	r.Static("/uploads", app.Config.Upload.Dir)
	r.Static("/swagger", "./docs")
	r.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, "/swagger/index.html")
	})
	if app.Config.MetricsEnabled() {
		r.GET(middleware.MetricsEndpointPath, gin.WrapH(promhttp.HandlerFor(app.MetricsRegistry, promhttp.HandlerOpts{})))
	}

	// Health check routes
	handler.RegisterHealthRoutes(r, func() bool {
		return checkDBHealth(app.DB)
	})

	// API routes — per-operation middleware
	publicRateLimit := app.PublicRateLimiter.Middleware()
	authMiddleware := middleware.Auth(app.JWTManager)
	userRateLimit := app.UserRateLimiter.MiddlewareWithKey(middleware.UserKey)

	// Create generated strict server
	apiServer := genapi.NewServer(app.AuthService, app.ItemService)
	strictHandler := genapi.NewStrictHandler(apiServer, nil)

	// Register generated routes with per-operation middleware.
	// We can't use RegisterHandlersWithOptions because its Middlewares field
	// is a flat list applied to ALL operations — not per-operation.
	// Instead, we register routes manually using the generated ServerInterface.
	genapi.RegisterHandlersWithOptions(r, strictHandler, genapi.GinServerOptions{
		Middlewares: []genapi.MiddlewareFunc{
			func(c *gin.Context) {
				// Resolve per-operation middleware based on the matched route.
				// The generated wrapper already sets BearerAuth scopes for authenticated ops.
				path := c.FullPath()
				switch {
				case isPublicAuthRoute(path):
					publicRateLimit(c)
				case isAuthenticatedRoute(path):
					authMiddleware(c)
					if c.IsAborted() {
						return
					}
					userRateLimit(c)
				}
			},
		},
		ErrorHandler: generatedRequestErrorHandler,
	})

	// Manual routes — upload (gin.Context-dependent, excluded from codegen)
	v1 := r.Group("/api/v1")
	v1.POST("/upload", authMiddleware, userRateLimit, app.UploadHandler.Upload)

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
		_ = sqlDB.Close()
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

// isPublicAuthRoute returns true for public auth routes that only need publicRateLimit.
func isPublicAuthRoute(path string) bool {
	switch path {
	case "/api/v1/auth/register", "/api/v1/auth/login", "/api/v1/auth/refresh":
		return true
	}
	return false
}

// isAuthenticatedRoute returns true for routes that need authMiddleware + userRateLimit.
func isAuthenticatedRoute(path string) bool {
	switch path {
	case "/api/v1/auth/profile",
		"/api/v1/items",
		"/api/v1/items/:id":
		return true
	}
	return false
}

func generatedRequestErrorHandler(c *gin.Context, err error, statusCode int) {
	message := http.StatusText(statusCode)
	if err != nil {
		message = err.Error()
	}
	response.Error(c, statusCode, requestErrorCode(statusCode), message)
}

func requestErrorCode(statusCode int) int {
	return errcode.FromHTTPStatus(statusCode)
}
