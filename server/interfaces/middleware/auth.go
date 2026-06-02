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
		token, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			if c.GetHeader("Authorization") == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未提供认证 token"})
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token 格式错误，应为 Bearer <token>"})
			return
		}

		email, err := auth.ParseJWT(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token 无效或已过期"})
			return
		}

		// 将用户邮箱注入上下文
		c.Set("user_email", email)
		c.Next()
	}
}

// OptionalJWTAuth 在公开接口上尽力解析 JWT。
// 没有 token 或 token 无效时继续公开访问；token 有效时注入 user_email。
func OptionalJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c.GetHeader("Authorization"))
		if ok {
			if email, err := auth.ParseJWT(token); err == nil {
				c.Set("user_email", email)
			}
		}
		c.Next()
	}
}

func bearerToken(authHeader string) (string, bool) {
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", false
	}
	return parts[1], true
}
