package requestlog

import (
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
			"page":   []string{"abc"},
			"status": []string{"sensitive-status"},
			"sort":   []string{"id desc"},
			"order":  []string{"ascending"},
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

		if got := summary.Safe["sort_present"]; got != true {
			t.Fatalf("sort_present = %v, want true", got)
		}
		if got := summary.Safe["sort_len"]; got != len("id desc") {
			t.Fatalf("sort_len = %v, want %d", got, len("id desc"))
		}

		if got := summary.Safe["order_present"]; got != true {
			t.Fatalf("order_present = %v, want true", got)
		}
		if got := summary.Safe["order_len"]; got != len("ascending") {
			t.Fatalf("order_len = %v, want %d", got, len("ascending"))
		}
	})

	t.Run("repeated allowlisted values log count only", func(t *testing.T) {
		summary := SummarizeQuery(url.Values{
			"page":  []string{"1", "2"},
			"order": []string{"asc", "desc"},
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

		if _, exists := summary.Safe["order"]; exists {
			t.Fatal("order raw value should not be logged when repeated")
		}
		if got := summary.Safe["order_len"]; got != 2 {
			t.Fatalf("order_len = %v, want %d", got, 2)
		}
	})
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
