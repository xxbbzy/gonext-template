package jwt

import (
	"strings"
	"testing"
	"time"
)

func newTestManager(t *testing.T, secret, accessExpiry, refreshExpiry string) *Manager {
	t.Helper()

	m, err := NewManager(secret, accessExpiry, refreshExpiry)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	return m
}

func mustGenerateAccessToken(t *testing.T, m *Manager, userID uint, role string) string {
	t.Helper()

	token, err := m.GenerateAccessToken(userID, role)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}
	return token
}

func mustGenerateRefreshToken(t *testing.T, m *Manager, userID uint, role string) string {
	t.Helper()

	token, err := m.GenerateRefreshToken(userID, role)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}
	return token
}

func TestNewManager_InvalidDuration(t *testing.T) {
	_, err := NewManager("secret", "invalid", "24h")
	if err == nil {
		t.Fatal("NewManager() expected error for invalid access token duration")
	}

	_, err = NewManager("secret", "15m", "invalid")
	if err == nil {
		t.Fatal("NewManager() expected error for invalid refresh token duration")
	}
}

func TestManager_GenerateTokens(t *testing.T) {
	m := newTestManager(t, "test-secret", "15m", "24h")

	testCases := []struct {
		name string
		gen  func(*testing.T, *Manager, uint, string) string
	}{
		{
			name: "generate access token",
			gen:  mustGenerateAccessToken,
		},
		{
			name: "generate refresh token",
			gen:  mustGenerateRefreshToken,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token := tc.gen(t, m, 42, "admin")
			if token == "" {
				t.Fatal("generated token should not be empty")
			}

			claims, err := m.ParseToken(token)
			if err != nil {
				t.Fatalf("ParseToken() error = %v", err)
			}
			if claims.UserID != 42 {
				t.Fatalf("claims.UserID = %d, want 42", claims.UserID)
			}
			if claims.Role != "admin" {
				t.Fatalf("claims.Role = %q, want %q", claims.Role, "admin")
			}
		})
	}
}

func TestManager_ParseToken_Valid(t *testing.T) {
	m := newTestManager(t, "test-secret", "15m", "24h")
	token := mustGenerateAccessToken(t, m, 7, "user")

	claims, err := m.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if claims.UserID != 7 {
		t.Fatalf("claims.UserID = %d, want 7", claims.UserID)
	}
	if claims.Role != "user" {
		t.Fatalf("claims.Role = %q, want %q", claims.Role, "user")
	}
}

func TestManager_ParseToken_Expired(t *testing.T) {
	m := newTestManager(t, "test-secret", "2s", "24h")
	token := mustGenerateAccessToken(t, m, 9, "user")

	if _, err := m.ParseToken(token); err != nil {
		t.Fatalf("ParseToken() before expiry error = %v", err)
	}

	time.Sleep(2500 * time.Millisecond)

	_, err := m.ParseToken(token)
	if err == nil {
		t.Fatal("ParseToken() expected error for expired token")
	}
	if !strings.Contains(err.Error(), "token expired") {
		t.Fatalf("ParseToken() error = %q, want contains %q", err.Error(), "token expired")
	}
}

func TestManager_ParseToken_Invalid(t *testing.T) {
	t.Run("malformed token", func(t *testing.T) {
		m := newTestManager(t, "test-secret", "15m", "24h")

		_, err := m.ParseToken("not-a-jwt")
		if err == nil {
			t.Fatal("ParseToken() expected error for malformed token")
		}
		if !strings.Contains(err.Error(), "invalid token") {
			t.Fatalf("ParseToken() error = %q, want contains %q", err.Error(), "invalid token")
		}
	})

	t.Run("invalid signature token", func(t *testing.T) {
		issuer := newTestManager(t, "secret-a", "15m", "24h")
		verifier := newTestManager(t, "secret-b", "15m", "24h")
		token := mustGenerateAccessToken(t, issuer, 11, "admin")

		_, err := verifier.ParseToken(token)
		if err == nil {
			t.Fatal("ParseToken() expected error for invalid signature token")
		}
		if !strings.Contains(err.Error(), "invalid token") {
			t.Fatalf("ParseToken() error = %q, want contains %q", err.Error(), "invalid token")
		}
	})
}
