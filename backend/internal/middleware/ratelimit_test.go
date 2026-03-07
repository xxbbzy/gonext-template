package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimiterUsesUserKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := NewRateLimiter(1, time.Hour)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint(42))
		c.Next()
	})
	router.Use(limiter.MiddlewareWithKey(UserKey))
	router.GET("/limited", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	firstReq := httptest.NewRequest(http.MethodGet, "/limited", nil)
	firstResp := httptest.NewRecorder()
	router.ServeHTTP(firstResp, firstReq)
	if firstResp.Code != http.StatusOK {
		t.Fatalf("first request status = %d, want %d", firstResp.Code, http.StatusOK)
	}

	secondReq := httptest.NewRequest(http.MethodGet, "/limited", nil)
	secondResp := httptest.NewRecorder()
	router.ServeHTTP(secondResp, secondReq)
	if secondResp.Code != http.StatusTooManyRequests {
		t.Fatalf("second request status = %d, want %d", secondResp.Code, http.StatusTooManyRequests)
	}
}
