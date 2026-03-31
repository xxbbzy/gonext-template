package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/xxbbzy/gonext-template/backend/internal/middleware"
	"github.com/xxbbzy/gonext-template/backend/pkg/errcode"
	"github.com/xxbbzy/gonext-template/backend/pkg/requestlog"
)

func TestOpenAPIErrorHelpersSetRequestLogErrorCode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name        string
		build       func(*gin.Context)
		wantLogCode int
	}{
		{
			name: "business error code",
			build: func(c *gin.Context) {
				_ = loginUserErrorResponse(c, http.StatusUnauthorized, errcode.ErrInvalidCreds, "invalid credentials")
			},
			wantLogCode: errcode.ErrInvalidCreds,
		},
		{
			name: "validation error code",
			build: func(c *gin.Context) {
				_ = registerUserErrorResponse(c, http.StatusBadRequest, errcode.ErrBadRequest, "bad request")
			},
			wantLogCode: errcode.ErrBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)

			tc.build(c)

			got, exists := requestlog.GetErrorCode(c)
			if !exists {
				t.Fatal("request-log error_code metadata missing")
			}
			if got != tc.wantLogCode {
				t.Fatalf("error_code = %d, want %d", got, tc.wantLogCode)
			}
		})
	}
}

func TestErrorResponseBodyUsesUnifiedPayloadFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Set(middleware.RequestIDKey, "req-openapi-123")
	c.Writer.Header().Set(middleware.RequestIDHeader, "req-openapi-123")

	body := errorResponseBody(c, errcode.ErrBadRequest, "bad request")

	if body.Code != errcode.ErrBadRequest {
		t.Fatalf("body.code = %d, want %d", body.Code, errcode.ErrBadRequest)
	}
	if body.Message != "bad request" {
		t.Fatalf("body.message = %q, want %q", body.Message, "bad request")
	}
	if body.RequestId != "req-openapi-123" {
		t.Fatalf("body.request_id = %q, want %q", body.RequestId, "req-openapi-123")
	}
	if body.Details != nil {
		t.Fatalf("body.details should be nil, got %#v", body.Details)
	}
}
