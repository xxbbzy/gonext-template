package response

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/xxbbzy/gonext-template/backend/pkg/requestlog"
)

const requestIDContextKey = "request_id"

// Response is the standard API response structure.
type Response struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// ErrorResponse is the canonical API error response structure.
type ErrorResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
	Details   interface{} `json:"details,omitempty"`
}

// PagedData wraps paginated results.
type PagedData struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// Success returns a successful response.
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Data:    data,
		Message: "success",
	})
}

// Created returns a 201 response.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:    0,
		Data:    data,
		Message: "success",
	})
}

// PagedSuccess returns a paginated success response.
func PagedSuccess(c *gin.Context, items interface{}, total int64, page, pageSize int) {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, Response{
		Code: 0,
		Data: PagedData{
			Items:      items,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
		Message: "success",
	})
}

// BuildErrorPayload builds the canonical error payload and records request-log metadata.
func BuildErrorPayload(ctx context.Context, code int, message string, details interface{}) ErrorResponse {
	requestlog.SetErrorCodeFromContext(ctx, code)
	return ErrorResponse{
		Code:      code,
		Message:   message,
		RequestID: requestIDFromContext(ctx),
		Details:   details,
	}
}

// Error returns an error response.
func Error(c *gin.Context, httpStatus int, code int, message string) {
	ErrorWithDetails(c, httpStatus, code, message, nil)
}

// ErrorWithDetails returns a canonical error response with optional details metadata.
func ErrorWithDetails(c *gin.Context, httpStatus int, code int, message string, details interface{}) {
	c.JSON(httpStatus, BuildErrorPayload(c, code, message, details))
}

func requestIDFromContext(ctx context.Context) string {
	ginCtx, ok := ctx.(*gin.Context)
	if !ok || ginCtx == nil {
		return ""
	}

	if requestID := ginCtx.Writer.Header().Get("X-Request-ID"); requestID != "" {
		return requestID
	}

	requestID, _ := ginCtx.Get(requestIDContextKey)
	value, ok := requestID.(string)
	if !ok {
		return ""
	}
	return value
}

// BadRequest returns a 400 error.
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, 400, message)
}

// Unauthorized returns a 401 error.
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, 401, message)
}

// Forbidden returns a 403 error.
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, 403, message)
}

// NotFound returns a 404 error.
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, 404, message)
}

// InternalServerError returns a 500 error.
func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, 500, message)
}
