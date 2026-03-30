package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

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
