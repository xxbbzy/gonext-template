package jwt

import (
	"errors"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims.
type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwtv5.RegisteredClaims
}

// Manager handles JWT token operations.
type Manager struct {
	secret        []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewManager creates a new JWT manager.
func NewManager(secret, accessExpiry, refreshExpiry string) (*Manager, error) {
	aDuration, err := time.ParseDuration(accessExpiry)
	if err != nil {
		return nil, errors.New("invalid access token expiry duration")
	}

	rDuration, err := time.ParseDuration(refreshExpiry)
	if err != nil {
		return nil, errors.New("invalid refresh token expiry duration")
	}

	return &Manager{
		secret:        []byte(secret),
		accessExpiry:  aDuration,
		refreshExpiry: rDuration,
	}, nil
}

// GenerateAccessToken generates a short-lived access token.
func (m *Manager) GenerateAccessToken(userID uint, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(m.accessExpiry)),
			IssuedAt:  jwtv5.NewNumericDate(time.Now()),
		},
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// GenerateRefreshToken generates a long-lived refresh token.
func (m *Manager) GenerateRefreshToken(userID uint, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(m.refreshExpiry)),
			IssuedAt:  jwtv5.NewNumericDate(time.Now()),
		},
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ParseToken parses and validates a JWT token.
func (m *Manager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwtv5.ParseWithClaims(tokenString, &Claims{}, func(token *jwtv5.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtv5.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwtv5.ErrTokenExpired) {
			return nil, errors.New("token expired")
		}
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
