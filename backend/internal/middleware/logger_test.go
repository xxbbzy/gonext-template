package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/xxbbzy/gonext-template/backend/pkg/errcode"
	pkgjwt "github.com/xxbbzy/gonext-template/backend/pkg/jwt"
	"github.com/xxbbzy/gonext-template/backend/pkg/response"
)

func TestRequestLoggerIncludesRoutePathAndUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core, recorded := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	router := gin.New()
	router.Use(RequestID(), RequestLogger(logger))
	router.GET("/items/:id", func(c *gin.Context) {
		c.Set("user_id", uint(42))
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/items/123?page=2&page_size=20&status=active&keyword=super-secret", nil)
	req.Header.Set(RequestIDHeader, "req-logger-1")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusNoContent)
	}

	fields := singleRequestLog(t, recorded)
	if got := fields["request_id"]; got != "req-logger-1" {
		t.Fatalf("request_id = %v, want %q", got, "req-logger-1")
	}
	if got := fields["route"]; got != "/items/:id" {
		t.Fatalf("route = %v, want %q", got, "/items/:id")
	}
	if got := fields["path"]; got != "/items/123" {
		t.Fatalf("path = %v, want %q", got, "/items/123")
	}
	if got := intFromAny(t, fields["user_id"]); got != 42 {
		t.Fatalf("user_id = %d, want 42", got)
	}

	queryKeys := stringSliceFromAny(t, fields["query_keys"])
	if !contains(queryKeys, "keyword") || !contains(queryKeys, "page") {
		t.Fatalf("query_keys = %v, want keyword/page present", queryKeys)
	}

	querySafe, ok := fields["query_safe"].(map[string]interface{})
	if !ok {
		t.Fatalf("query_safe type = %T, want map[string]interface{}", fields["query_safe"])
	}
	if got := intFromAny(t, querySafe["page"]); got != 2 {
		t.Fatalf("query_safe.page = %v, want %d", got, 2)
	}
	if got := querySafe["status"]; got != "active" {
		t.Fatalf("query_safe.status = %v, want %q", got, "active")
	}
	if got := querySafe["keyword_present"]; got != true {
		t.Fatalf("query_safe.keyword_present = %v, want true", got)
	}
	if got := intFromAny(t, querySafe["keyword_len"]); got != len("super-secret") {
		t.Fatalf("query_safe.keyword_len = %v, want %d", got, len("super-secret"))
	}
	if _, exists := querySafe["keyword"]; exists {
		t.Fatal("query_safe leaked raw keyword value")
	}
}

func TestRequestLoggerOmitsUserIDForAnonymousRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core, recorded := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	router := gin.New()
	router.Use(RequestID(), RequestLogger(logger))
	router.GET("/public", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusNoContent)
	}

	fields := singleRequestLog(t, recorded)
	if _, exists := fields["user_id"]; exists {
		t.Fatalf("user_id should be omitted for anonymous request, got %v", fields["user_id"])
	}
}

func TestRequestLoggerIncludesErrorCodeForManualAndRateLimitErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("manual response helper", func(t *testing.T) {
		core, recorded := observer.New(zap.InfoLevel)
		logger := zap.New(core)

		router := gin.New()
		router.Use(RequestID(), RequestLogger(logger))
		router.GET("/missing", func(c *gin.Context) {
			response.NotFound(c, "resource not found")
		})

		req := httptest.NewRequest(http.MethodGet, "/missing", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", resp.Code, http.StatusNotFound)
		}

		fields := singleRequestLog(t, recorded)
		if got := intFromAny(t, fields["error_code"]); got != errcode.ErrNotFound {
			t.Fatalf("error_code = %d, want %d", got, errcode.ErrNotFound)
		}
	})

	t.Run("rate limiter response helper", func(t *testing.T) {
		core, recorded := observer.New(zap.InfoLevel)
		logger := zap.New(core)

		limiter := NewRateLimiter(1, time.Hour)
		router := gin.New()
		router.Use(RequestID(), RequestLogger(logger), limiter.Middleware())
		router.GET("/limited", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		for i := 0; i < 2; i++ {
			req := httptest.NewRequest(http.MethodGet, "/limited", nil)
			req.RemoteAddr = "198.51.100.99:1234"
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			wantStatus := http.StatusOK
			if i == 1 {
				wantStatus = http.StatusTooManyRequests
			}
			if got := resp.Code; got != wantStatus {
				t.Fatalf("request %d status = %d, want %d", i+1, got, wantStatus)
			}
		}

		entries := recorded.FilterMessage("HTTP Request").All()
		if len(entries) != 2 {
			t.Fatalf("log entries = %d, want 2", len(entries))
		}

		last := entries[len(entries)-1].ContextMap()
		if got := intFromAny(t, last["status"]); got != http.StatusTooManyRequests {
			t.Fatalf("status = %d, want %d", got, http.StatusTooManyRequests)
		}
		if got := intFromAny(t, last["error_code"]); got != errcode.ErrTooManyReqs {
			t.Fatalf("error_code = %d, want %d", got, errcode.ErrTooManyReqs)
		}
	})
}

func TestAuthFailureErrorCodesAreLogged(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name      string
		token     func(t *testing.T) string
		wantCode  int
		wantError int
	}{
		{
			name: "expired token",
			token: func(t *testing.T) string {
				manager, err := pkgjwt.NewManager("test-secret", "-1s", "24h")
				if err != nil {
					t.Fatalf("new jwt manager: %v", err)
				}
				token, err := manager.GenerateAccessToken(1, "user")
				if err != nil {
					t.Fatalf("generate token: %v", err)
				}
				return token
			},
			wantCode:  http.StatusUnauthorized,
			wantError: errcode.ErrTokenExpired,
		},
		{
			name: "invalid token",
			token: func(_ *testing.T) string {
				return "not-a-valid-token"
			},
			wantCode:  http.StatusUnauthorized,
			wantError: errcode.ErrTokenInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			core, recorded := observer.New(zap.InfoLevel)
			logger := zap.New(core)

			manager, err := pkgjwt.NewManager("test-secret", "15m", "24h")
			if err != nil {
				t.Fatalf("new jwt manager: %v", err)
			}

			router := gin.New()
			router.Use(RequestID(), RequestLogger(logger))
			router.GET("/protected", Auth(manager), func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", "Bearer "+tc.token(t))
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			if resp.Code != tc.wantCode {
				t.Fatalf("status = %d, want %d", resp.Code, tc.wantCode)
			}

			var body response.Response
			if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
				t.Fatalf("unmarshal response: %v", err)
			}
			if body.Code != tc.wantError {
				t.Fatalf("body.code = %d, want %d", body.Code, tc.wantError)
			}

			fields := singleRequestLog(t, recorded)
			if got := intFromAny(t, fields["error_code"]); got != tc.wantError {
				t.Fatalf("error_code = %d, want %d", got, tc.wantError)
			}
		})
	}
}

func singleRequestLog(t *testing.T, recorded *observer.ObservedLogs) map[string]interface{} {
	t.Helper()

	entries := recorded.FilterMessage("HTTP Request").All()
	if len(entries) != 1 {
		t.Fatalf("log entries = %d, want 1", len(entries))
	}
	return entries[0].ContextMap()
}

func intFromAny(t *testing.T, value interface{}) int {
	t.Helper()

	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case uint:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
	case float64:
		return int(v)
	default:
		t.Fatalf("unexpected numeric type %T (%v)", value, value)
		return 0
	}
}

func contains(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

func stringSliceFromAny(t *testing.T, value interface{}) []string {
	t.Helper()

	switch v := value.(type) {
	case []string:
		return v
	case []interface{}:
		out := make([]string, len(v))
		for i, item := range v {
			s, ok := item.(string)
			if !ok {
				t.Fatalf("query_keys[%d] type = %T, want string", i, item)
			}
			out[i] = s
		}
		return out
	default:
		t.Fatalf("unexpected query_keys type %T", value)
		return nil
	}
}
