package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xxbbzy/gonext-template/backend/pkg/errcode"
	"github.com/xxbbzy/gonext-template/backend/pkg/response"
)

// Recovery returns a middleware that recovers from panics.
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("stack", string(debug.Stack())),
				)
				response.Error(c, http.StatusInternalServerError,
					errcode.ErrInternal, "internal server error")
				c.Abort()
			}
		}()
		c.Next()
	}
}

// ErrorHandler returns a middleware that handles AppError returned in context.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors set
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			if appErr, ok := err.(*errcode.AppError); ok {
				response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
				return
			}
			response.Error(c, http.StatusInternalServerError,
				errcode.ErrInternal, "internal server error")
		}
	}
}
