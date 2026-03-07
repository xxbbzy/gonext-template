package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/xxbbzy/gonext-template/backend/internal/dto"
	"github.com/xxbbzy/gonext-template/backend/internal/service"
	"github.com/xxbbzy/gonext-template/backend/pkg/errcode"
	"github.com/xxbbzy/gonext-template/backend/pkg/response"
)

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register handles user registration.
// @Summary Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body dto.RegisterRequest true "Registration data"
// @Success 201 {object} response.Response{data=dto.AuthResponse}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.authService.Register(&req)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
			return
		}
		response.InternalServerError(c, "registration failed")
		return
	}

	response.Created(c, result)
}

// Login handles user login.
// @Summary Login with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body dto.LoginRequest true "Login data"
// @Success 200 {object} response.Response{data=dto.AuthResponse}
// @Failure 401 {object} response.Response
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.authService.Login(&req)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
			return
		}
		response.InternalServerError(c, "login failed")
		return
	}

	response.Success(c, result)
}

// Refresh handles token refresh.
// @Summary Refresh access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body dto.RefreshRequest true "Refresh token"
// @Success 200 {object} response.Response{data=dto.AuthResponse}
// @Failure 401 {object} response.Response
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
			return
		}
		response.InternalServerError(c, "token refresh failed")
		return
	}

	response.Success(c, result)
}

// GetProfile returns the current user's profile.
// @Summary Get current user profile
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Failure 401 {object} response.Response
// @Router /api/v1/auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, ok := userID.(uint)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	profile, err := h.authService.GetProfile(uid)
	if err != nil {
		if appErr, ok := err.(*errcode.AppError); ok {
			response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
			return
		}
		response.InternalServerError(c, "failed to load profile")
		return
	}

	response.Success(c, profile)
}

// RegisterRoutes registers auth routes.
func (h *AuthHandler) RegisterRoutes(
	r *gin.RouterGroup,
	publicMiddlewares []gin.HandlerFunc,
	authMiddleware gin.HandlerFunc,
	protectedMiddlewares ...gin.HandlerFunc,
) {
	auth := r.Group("/auth")
	{
		registerHandlers := append(append([]gin.HandlerFunc{}, publicMiddlewares...), h.Register)
		loginHandlers := append(append([]gin.HandlerFunc{}, publicMiddlewares...), h.Login)
		refreshHandlers := append(append([]gin.HandlerFunc{}, publicMiddlewares...), h.Refresh)
		auth.POST("/register", registerHandlers...)
		auth.POST("/login", loginHandlers...)
		auth.POST("/refresh", refreshHandlers...)
		profileHandlers := append([]gin.HandlerFunc{authMiddleware}, protectedMiddlewares...)
		profileHandlers = append(profileHandlers, h.GetProfile)
		auth.GET("/profile", profileHandlers...)
	}
}

// RegisterHealthRoutes registers health check endpoints.
func RegisterHealthRoutes(r *gin.Engine, healthCheck func() bool) {
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive"})
	})

	r.GET("/readyz", func(c *gin.Context) {
		if healthCheck() {
			c.JSON(http.StatusOK, gin.H{"status": "ready"})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
		}
	})
}
