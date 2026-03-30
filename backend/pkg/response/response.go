package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/xxbbzy/gonext-template/backend/pkg/requestlog"
)

// Response is the standard API response structure.
type Response struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
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

// Error returns an error response.
func Error(c *gin.Context, httpStatus int, code int, message string) {
	requestlog.SetErrorCode(c, code)
	c.JSON(httpStatus, Response{
		Code:    code,
		Data:    nil,
		Message: message,
	})
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
