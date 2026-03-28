package pagination

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		rawQuery string
		want     Params
	}{
		{
			name:     "uses default page and page_size when query is missing",
			rawQuery: "",
			want: Params{
				Page:     DefaultPage,
				PageSize: DefaultPageSize,
				Offset:   0,
			},
		},
		{
			name:     "resets page when page is less than one",
			rawQuery: "page=0&page_size=10",
			want: Params{
				Page:     DefaultPage,
				PageSize: 10,
				Offset:   0,
			},
		},
		{
			name:     "resets page_size when page_size is less than one",
			rawQuery: "page=2&page_size=0",
			want: Params{
				Page:     2,
				PageSize: DefaultPageSize,
				Offset:   DefaultPageSize,
			},
		},
		{
			name:     "caps page_size when page_size exceeds max",
			rawQuery: "page=2&page_size=101",
			want: Params{
				Page:     2,
				PageSize: MaxPageSize,
				Offset:   MaxPageSize,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest("GET", "/items", nil)
			req.URL.RawQuery = tt.rawQuery
			c.Request = req

			got := Parse(c)

			if got.Page != tt.want.Page {
				t.Fatalf("Page = %d, want %d", got.Page, tt.want.Page)
			}
			if got.PageSize != tt.want.PageSize {
				t.Fatalf("PageSize = %d, want %d", got.PageSize, tt.want.PageSize)
			}
			if got.Offset != tt.want.Offset {
				t.Fatalf("Offset = %d, want %d", got.Offset, tt.want.Offset)
			}
		})
	}
}

func TestNewParams(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		want     Params
	}{
		{
			name:     "uses default page and page_size",
			page:     0,
			pageSize: 0,
			want: Params{
				Page:     DefaultPage,
				PageSize: DefaultPageSize,
				Offset:   0,
			},
		},
		{
			name:     "resets page when page is less than one",
			page:     0,
			pageSize: 20,
			want: Params{
				Page:     DefaultPage,
				PageSize: 20,
				Offset:   0,
			},
		},
		{
			name:     "resets page_size when page_size is less than one",
			page:     2,
			pageSize: 0,
			want: Params{
				Page:     2,
				PageSize: DefaultPageSize,
				Offset:   DefaultPageSize,
			},
		},
		{
			name:     "caps page_size when page_size exceeds max",
			page:     2,
			pageSize: 101,
			want: Params{
				Page:     2,
				PageSize: MaxPageSize,
				Offset:   MaxPageSize,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewParams(tt.page, tt.pageSize)

			if got.Page != tt.want.Page {
				t.Fatalf("Page = %d, want %d", got.Page, tt.want.Page)
			}
			if got.PageSize != tt.want.PageSize {
				t.Fatalf("PageSize = %d, want %d", got.PageSize, tt.want.PageSize)
			}
			if got.Offset != tt.want.Offset {
				t.Fatalf("Offset = %d, want %d", got.Offset, tt.want.Offset)
			}
		})
	}
}
