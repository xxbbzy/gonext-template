package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xxbbzy/gonext-template/backend/pkg/requestlog"
)

// RequestLogger returns a middleware that logs each request.
func RequestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()
		route, path := requestlog.RouteAndPath(c)
		querySummary := requestlog.SummarizeContextQuery(c)
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("route", route),
			zap.String("path", path),
			zap.String("request_id", GetRequestID(c)),
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.String("client_ip", c.ClientIP()),
			zap.Int("body_size", c.Writer.Size()),
			zap.Strings("query_keys", querySummary.Keys),
			zap.Any("query_safe", querySummary.Safe),
		}

		if userID, exists := c.Get("user_id"); exists {
			fields = append(fields, zap.Any("user_id", userID))
		}

		if errorCode, exists := requestlog.GetErrorCode(c); exists {
			fields = append(fields, zap.Int("error_code", errorCode))
		}

		logger.Info("HTTP Request", fields...)
	}
}
