package requestlog

import (
	"context"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSummarizeQuery(t *testing.T) {
	summary := SummarizeQuery(url.Values{
		"page":    []string{"2"},
		"status":  []string{"active"},
		"keyword": []string{"very-secret"},
	})

	if got, want := len(summary.Keys), 3; got != want {
		t.Fatalf("keys len = %d, want %d", got, want)
	}
	if got := intFromAny(t, summary.Safe["page"]); got != 2 {
		t.Fatalf("safe page = %v, want %d", got, 2)
	}
	if got := summary.Safe["status"]; got != "active" {
		t.Fatalf("safe status = %v, want %q", got, "active")
	}
	if got := summary.Safe["keyword_present"]; got != true {
		t.Fatalf("safe keyword_present = %v, want true", got)
	}
	if got := summary.Safe["keyword_len"]; got != len("very-secret") {
		t.Fatalf("safe keyword_len = %v, want %d", got, len("very-secret"))
	}
	if _, exists := summary.Safe["keyword"]; exists {
		t.Fatal("safe summary leaked raw keyword value")
	}
}

func TestSummarizeQueryAllowlistedFallbacks(t *testing.T) {
	t.Run("invalid allowlisted values are summarized", func(t *testing.T) {
		summary := SummarizeQuery(url.Values{
			"page":      []string{"abc"},
			"page_size": []string{"-"},
			"status":    []string{"sensitive-status"},
		})

		if _, exists := summary.Safe["page"]; exists {
			t.Fatal("page raw value should not be logged when invalid")
		}
		if got := summary.Safe["page_present"]; got != true {
			t.Fatalf("page_present = %v, want true", got)
		}
		if got := summary.Safe["page_len"]; got != len("abc") {
			t.Fatalf("page_len = %v, want %d", got, len("abc"))
		}

		if got := summary.Safe["status_present"]; got != true {
			t.Fatalf("status_present = %v, want true", got)
		}
		if got := summary.Safe["status_len"]; got != len("sensitive-status") {
			t.Fatalf("status_len = %v, want %d", got, len("sensitive-status"))
		}

		if got := summary.Safe["page_size_present"]; got != true {
			t.Fatalf("page_size_present = %v, want true", got)
		}
		if got := summary.Safe["page_size_len"]; got != len("-") {
			t.Fatalf("page_size_len = %v, want %d", got, len("-"))
		}
	})

	t.Run("repeated allowlisted values log count only", func(t *testing.T) {
		summary := SummarizeQuery(url.Values{
			"page":      []string{"1", "2"},
			"page_size": []string{"20", "30"},
		})

		if _, exists := summary.Safe["page"]; exists {
			t.Fatal("page raw value should not be logged when repeated")
		}
		if got := summary.Safe["page_len"]; got != 2 {
			t.Fatalf("page_len = %v, want %d", got, 2)
		}
		if _, exists := summary.Safe["page_present"]; exists {
			t.Fatalf("page_present should be omitted for repeated values, got %v", summary.Safe["page_present"])
		}

		if _, exists := summary.Safe["page_size"]; exists {
			t.Fatal("page_size raw value should not be logged when repeated")
		}
		if got := summary.Safe["page_size_len"]; got != 2 {
			t.Fatalf("page_size_len = %v, want %d", got, 2)
		}
	})
}

func TestSetAndGetErrorCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	SetErrorCode(c, 404)
	got, ok := GetErrorCode(c)
	if !ok || got != 404 {
		t.Fatalf("GetErrorCode() = (%d, %v), want (404, true)", got, ok)
	}

	SetErrorCode(c, -1)
	got, ok = GetErrorCode(c)
	if ok || got != 0 {
		t.Fatalf("GetErrorCode() with negative = (%d, %v), want (0, false)", got, ok)
	}
}

func TestSetErrorCodeFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	SetErrorCodeFromContext(c, 429)
	got, ok := GetErrorCode(c)
	if !ok || got != 429 {
		t.Fatalf("GetErrorCode() = (%d, %v), want (429, true)", got, ok)
	}
}

func TestSetErrorCodeFromContext_NilOrNonGinContext(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SetErrorCodeFromContext panicked: %v", r)
		}
	}()

	var nilCtx context.Context
	SetErrorCodeFromContext(nilCtx, 400)
	SetErrorCodeFromContext(context.Background(), 400)
}

func TestGetErrorCode_AbsentOrInvalidValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	if got, ok := GetErrorCode(c); ok || got != 0 {
		t.Fatalf("GetErrorCode() with missing key = (%d, %v), want (0, false)", got, ok)
	}

	c.Set(ErrorCodeContextKey, "bad")
	if got, ok := GetErrorCode(c); ok || got != 0 {
		t.Fatalf("GetErrorCode() with string = (%d, %v), want (0, false)", got, ok)
	}

	c.Set(ErrorCodeContextKey, int64(-1))
	if got, ok := GetErrorCode(c); ok || got != 0 {
		t.Fatalf("GetErrorCode() with negative int64 = (%d, %v), want (0, false)", got, ok)
	}

	c.Set(ErrorCodeContextKey, uint64(math.MaxInt)+1)
	if got, ok := GetErrorCode(c); ok || got != 0 {
		t.Fatalf("GetErrorCode() with overflow uint64 = (%d, %v), want (0, false)", got, ok)
	}
}

func TestRouteAndPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/items/:id", func(c *gin.Context) {
		route, path := RouteAndPath(c)
		if route != "/items/:id" {
			t.Fatalf("route = %q, want %q", route, "/items/:id")
		}
		if path != "/items/123" {
			t.Fatalf("path = %q, want %q", path, "/items/123")
		}
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/items/123", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusNoContent)
	}
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
