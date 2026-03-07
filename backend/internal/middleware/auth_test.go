package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	pkgjwt "github.com/xxbbzy/gonext-template/backend/pkg/jwt"
)

func TestAuthRejectsMissingAuthorizationHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtManager, err := pkgjwt.NewManager("test-secret", "15m", "24h")
	if err != nil {
		t.Fatalf("new jwt manager: %v", err)
	}

	router := gin.New()
	router.GET("/protected", Auth(jwtManager), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusUnauthorized)
	}
}

func TestAuthAcceptsValidBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtManager, err := pkgjwt.NewManager("test-secret", "15m", "24h")
	if err != nil {
		t.Fatalf("new jwt manager: %v", err)
	}

	token, err := jwtManager.GenerateAccessToken(42, "admin")
	if err != nil {
		t.Fatalf("generate access token: %v", err)
	}

	router := gin.New()
	router.GET("/protected", Auth(jwtManager), func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		role, _ := c.Get("user_role")
		if userID != uint(42) {
			t.Fatalf("user_id = %v, want 42", userID)
		}
		if role != "admin" {
			t.Fatalf("role = %v, want admin", role)
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusOK)
	}
}
