package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	pkgjwt "github.com/xxbbzy/gonext-template/backend/pkg/jwt"
	"github.com/xxbbzy/gonext-template/backend/pkg/response"
)

// Auth returns a middleware that validates JWT tokens.
func Auth(jwtManager *pkgjwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "unauthorized")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		claims, err := jwtManager.ParseToken(parts[1])
		if err != nil {
			if err.Error() == "token expired" {
				response.Unauthorized(c, "token expired")
			} else {
				response.Unauthorized(c, "invalid token")
			}
			c.Abort()
			return
		}

		// Inject user info into context
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

// RequireRole returns a middleware that checks if the user has the required role.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			response.Unauthorized(c, "unauthorized")
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			response.Unauthorized(c, "unauthorized")
			c.Abort()
			return
		}

		for _, r := range roles {
			if r == role {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "forbidden")
		c.Abort()
	}
}
