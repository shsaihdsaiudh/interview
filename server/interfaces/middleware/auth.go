package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"interview-server/infrastructure/auth"
)

// JWTAuth JWT 认证中间件，保护需要登录的路由。
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未提供认证 token"})
			return
		}

		// 格式: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token 格式错误，应为 Bearer <token>"})
			return
		}

		email, err := auth.ParseJWT(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token 无效或已过期"})
			return
		}

		// 将用户邮箱注入上下文
		c.Set("user_email", email)
		c.Next()
	}
}
