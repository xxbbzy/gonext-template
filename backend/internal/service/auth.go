package service

import (
	"errors"

	"github.com/xxbbzy/gonext-template/backend/internal/dto"
	"github.com/xxbbzy/gonext-template/backend/internal/model"
	"github.com/xxbbzy/gonext-template/backend/internal/repository"
	"github.com/xxbbzy/gonext-template/backend/pkg/errcode"
	pkgjwt "github.com/xxbbzy/gonext-template/backend/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService handles authentication business logic.
type AuthService struct {
	userRepo   *repository.UserRepository
	jwtManager *pkgjwt.Manager
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo *repository.UserRepository, jwtManager *pkgjwt.Manager) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// Register registers a new user.
func (s *AuthService) Register(req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	// Check if email already exists
	if _, err := s.userRepo.FindByEmail(req.Email); err == nil {
		return nil, errcode.ErrEmailAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errcode.ErrInternalServer
	}

	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         "user",
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errcode.ErrInternalServer
	}

	return s.generateTokenResponse(user)
}

// Login authenticates a user with email and password.
func (s *AuthService) Login(req *dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrInvalidCredentials
		}
		return nil, errcode.ErrInternalServer
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errcode.ErrInvalidCredentials
	}

	return s.generateTokenResponse(user)
}

// RefreshToken refreshes the access token using a refresh token.
func (s *AuthService) RefreshToken(refreshToken string) (*dto.AuthResponse, error) {
	claims, err := s.jwtManager.ParseToken(refreshToken)
	if err != nil {
		if err.Error() == "token expired" {
			return nil, errcode.ErrRefreshTokenExpired
		}
		return nil, errcode.ErrTokenInvalidMsg
	}

	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, errcode.ErrInternalServer
	}

	return s.generateTokenResponse(user)
}

func (s *AuthService) generateTokenResponse(user *model.User) (*dto.AuthResponse, error) {
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, errcode.ErrInternalServer
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		return nil, errcode.ErrInternalServer
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: dto.UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	}, nil
}
