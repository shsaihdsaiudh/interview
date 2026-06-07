package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"interview-server/domain/user"
	"interview-server/infrastructure/auth"
)

// AdminOnly 管理员认证中间件。
// 先验证 JWT，再查询用户角色，非 admin 返回 403。
func AdminOnly(repo user.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未提供认证 token"})
			return
		}

		email, err := auth.ParseJWT(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token 无效或已过期"})
			return
		}

		u, err := repo.FindByEmail(email)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "无管理员权限"})
			return
		}

		if !u.IsAdmin() {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "无管理员权限"})
			return
		}

		c.Set("user_email", email)
		c.Next()
	}
}
