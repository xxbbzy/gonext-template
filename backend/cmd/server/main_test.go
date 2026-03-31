package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	genapi "github.com/xxbbzy/gonext-template/backend/internal/api"
	"github.com/xxbbzy/gonext-template/backend/internal/middleware"
	"github.com/xxbbzy/gonext-template/backend/pkg/errcode"
	"github.com/xxbbzy/gonext-template/backend/pkg/response"
)

func TestGeneratedWrapperBindingFailuresUseStandardErrorEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core, recorded := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	router := gin.New()
	router.Use(middleware.RequestID(), middleware.RequestLogger(logger))

	genapi.RegisterHandlersWithOptions(router, noopGeneratedServer{}, genapi.GinServerOptions{
		ErrorHandler: generatedRequestErrorHandler,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/items?page=not-a-number", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusBadRequest)
	}

	var body response.ErrorResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Code != errcode.ErrBadRequest {
		t.Fatalf("body.code = %d, want %d", body.Code, errcode.ErrBadRequest)
	}
	if !strings.Contains(body.Message, "Invalid format for parameter page") {
		t.Fatalf("body.message = %q, want parse error details", body.Message)
	}
	if body.RequestID == "" {
		t.Fatal("body.request_id should not be empty")
	}
	if got := resp.Header().Get(middleware.RequestIDHeader); got != body.RequestID {
		t.Fatalf("header request id = %q, want %q", got, body.RequestID)
	}

	entries := recorded.FilterMessage("HTTP Request").All()
	if len(entries) != 1 {
		t.Fatalf("log entries = %d, want 1", len(entries))
	}

	fields := entries[0].ContextMap()
	if got := asInt(t, fields["error_code"]); got != errcode.ErrBadRequest {
		t.Fatalf("error_code = %d, want %d", got, errcode.ErrBadRequest)
	}
}

func TestStrictHandlerJSONBindingFailuresUseStandardErrorEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core, recorded := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	router := gin.New()
	router.Use(
		middleware.RequestID(),
		middleware.RequestLogger(logger),
		middleware.ErrorHandler(),
	)

	strictHandler := genapi.NewStrictHandler(noopStrictServer{}, nil)
	genapi.RegisterHandlersWithOptions(router, strictHandler, genapi.GinServerOptions{
		ErrorHandler: generatedRequestErrorHandler,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusBadRequest)
	}

	var body response.ErrorResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Code != errcode.ErrBadRequest {
		t.Fatalf("body.code = %d, want %d", body.Code, errcode.ErrBadRequest)
	}
	if !strings.Contains(body.Message, "unexpected EOF") {
		t.Fatalf("body.message = %q, want JSON parse error details", body.Message)
	}
	if body.RequestID == "" {
		t.Fatal("body.request_id should not be empty")
	}
	if got := resp.Header().Get(middleware.RequestIDHeader); got != body.RequestID {
		t.Fatalf("header request id = %q, want %q", got, body.RequestID)
	}

	entries := recorded.FilterMessage("HTTP Request").All()
	if len(entries) != 1 {
		t.Fatalf("log entries = %d, want 1", len(entries))
	}
	fields := entries[0].ContextMap()
	if got := asInt(t, fields["error_code"]); got != errcode.ErrBadRequest {
		t.Fatalf("error_code = %d, want %d", got, errcode.ErrBadRequest)
	}
}

func TestRequestErrorCodeMapping(t *testing.T) {
	testCases := []struct {
		status int
		want   int
	}{
		{status: http.StatusBadRequest, want: errcode.ErrBadRequest},
		{status: http.StatusUnauthorized, want: errcode.ErrUnauthorized},
		{status: http.StatusForbidden, want: errcode.ErrForbidden},
		{status: http.StatusNotFound, want: errcode.ErrNotFound},
		{status: http.StatusConflict, want: errcode.ErrConflict},
		{status: http.StatusTooManyRequests, want: errcode.ErrTooManyReqs},
		{status: http.StatusRequestEntityTooLarge, want: errcode.ErrFileTooLarge},
		{status: http.StatusInternalServerError, want: errcode.ErrInternal},
		{status: 499, want: 499},
	}

	for _, tc := range testCases {
		if got := requestErrorCode(tc.status); got != tc.want {
			t.Fatalf("requestErrorCode(%d) = %d, want %d", tc.status, got, tc.want)
		}
	}
}

type noopGeneratedServer struct{}

func (noopGeneratedServer) LoginUser(c *gin.Context, _ genapi.LoginUserParams) {
	c.Status(http.StatusNoContent)
}

func (noopGeneratedServer) GetProfile(c *gin.Context, _ genapi.GetProfileParams) {
	c.Status(http.StatusNoContent)
}

func (noopGeneratedServer) RefreshToken(c *gin.Context, _ genapi.RefreshTokenParams) {
	c.Status(http.StatusNoContent)
}

func (noopGeneratedServer) RegisterUser(c *gin.Context, _ genapi.RegisterUserParams) {
	c.Status(http.StatusNoContent)
}

func (noopGeneratedServer) ListItems(c *gin.Context, _ genapi.ListItemsParams) {
	c.Status(http.StatusNoContent)
}

func (noopGeneratedServer) CreateItem(c *gin.Context, _ genapi.CreateItemParams) {
	c.Status(http.StatusNoContent)
}

func (noopGeneratedServer) DeleteItem(c *gin.Context, _ int, _ genapi.DeleteItemParams) {
	c.Status(http.StatusNoContent)
}

func (noopGeneratedServer) GetItem(c *gin.Context, _ int, _ genapi.GetItemParams) {
	c.Status(http.StatusNoContent)
}

func (noopGeneratedServer) UpdateItem(c *gin.Context, _ int, _ genapi.UpdateItemParams) {
	c.Status(http.StatusNoContent)
}

type noopStrictServer struct{}

func (noopStrictServer) LoginUser(_ context.Context, _ genapi.LoginUserRequestObject) (genapi.LoginUserResponseObject, error) {
	return nil, nil
}

func (noopStrictServer) GetProfile(_ context.Context, _ genapi.GetProfileRequestObject) (genapi.GetProfileResponseObject, error) {
	return nil, nil
}

func (noopStrictServer) RefreshToken(_ context.Context, _ genapi.RefreshTokenRequestObject) (genapi.RefreshTokenResponseObject, error) {
	return nil, nil
}

func (noopStrictServer) RegisterUser(_ context.Context, _ genapi.RegisterUserRequestObject) (genapi.RegisterUserResponseObject, error) {
	return nil, nil
}

func (noopStrictServer) ListItems(_ context.Context, _ genapi.ListItemsRequestObject) (genapi.ListItemsResponseObject, error) {
	return nil, nil
}

func (noopStrictServer) CreateItem(_ context.Context, _ genapi.CreateItemRequestObject) (genapi.CreateItemResponseObject, error) {
	return nil, nil
}

func (noopStrictServer) DeleteItem(_ context.Context, _ genapi.DeleteItemRequestObject) (genapi.DeleteItemResponseObject, error) {
	return nil, nil
}

func (noopStrictServer) GetItem(_ context.Context, _ genapi.GetItemRequestObject) (genapi.GetItemResponseObject, error) {
	return nil, nil
}

func (noopStrictServer) UpdateItem(_ context.Context, _ genapi.UpdateItemRequestObject) (genapi.UpdateItemResponseObject, error) {
	return nil, nil
}

func asInt(t *testing.T, value interface{}) int {
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
