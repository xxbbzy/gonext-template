package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestRequestIDUsesIncomingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestID())
	router.GET("/request-id", func(c *gin.Context) {
		c.Header("X-Seen-Request-ID", GetRequestID(c))
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/request-id", nil)
	req.Header.Set(RequestIDHeader, "upstream-123")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusNoContent)
	}
	if got := resp.Header().Get(RequestIDHeader); got != "upstream-123" {
		t.Fatalf("response %s = %q, want %q", RequestIDHeader, got, "upstream-123")
	}
	if got := resp.Header().Get("X-Seen-Request-ID"); got != "upstream-123" {
		t.Fatalf("handler saw request ID %q, want %q", got, "upstream-123")
	}
}

func TestRequestIDGeneratesForMissingOrBlankHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name   string
		header string
	}{
		{name: "missing header"},
		{name: "blank header", header: "   "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.Use(RequestID())
			router.GET("/request-id", func(c *gin.Context) {
				c.Header("X-Seen-Request-ID", GetRequestID(c))
				c.Status(http.StatusNoContent)
			})

			req := httptest.NewRequest(http.MethodGet, "/request-id", nil)
			if tc.header != "" {
				req.Header.Set(RequestIDHeader, tc.header)
			}
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			if resp.Code != http.StatusNoContent {
				t.Fatalf("status = %d, want %d", resp.Code, http.StatusNoContent)
			}

			got := resp.Header().Get(RequestIDHeader)
			if got == "" {
				t.Fatal("response request ID is empty, want generated value")
			}
			if _, err := uuid.Parse(got); err != nil {
				t.Fatalf("response request ID = %q, want valid UUID: %v", got, err)
			}
			if seen := resp.Header().Get("X-Seen-Request-ID"); seen != got {
				t.Fatalf("handler saw request ID %q, want %q", seen, got)
			}
		})
	}
}

func TestRequestLoggerIncludesRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core, recorded := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	router := gin.New()
	router.Use(RequestID(), RequestLogger(logger))
	router.GET("/request-id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/request-id", nil)
	req.Header.Set(RequestIDHeader, "req-log-123")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusNoContent)
	}

	entries := recorded.FilterMessage("HTTP Request").All()
	if len(entries) != 1 {
		t.Fatalf("log entries = %d, want 1", len(entries))
	}
	if got := entries[0].ContextMap()["request_id"]; got != "req-log-123" {
		t.Fatalf("logged request_id = %v, want %q", got, "req-log-123")
	}
}

func TestRecoveryIncludesRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core, recorded := observer.New(zap.ErrorLevel)
	logger := zap.New(core)

	router := gin.New()
	router.Use(RequestID(), Recovery(logger))
	router.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusInternalServerError)
	}

	requestID := resp.Header().Get(RequestIDHeader)
	if requestID == "" {
		t.Fatal("response request ID is empty, want generated value")
	}
	if _, err := uuid.Parse(requestID); err != nil {
		t.Fatalf("response request ID = %q, want valid UUID: %v", requestID, err)
	}

	entries := recorded.FilterMessage("Panic recovered").All()
	if len(entries) != 1 {
		t.Fatalf("log entries = %d, want 1", len(entries))
	}
	if got := entries[0].ContextMap()["request_id"]; got != requestID {
		t.Fatalf("logged request_id = %v, want %q", got, requestID)
	}
}
