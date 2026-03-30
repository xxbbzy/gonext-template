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
	if got := summary.Safe["page"]; got != "2" {
		t.Fatalf("safe page = %v, want %q", got, "2")
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
