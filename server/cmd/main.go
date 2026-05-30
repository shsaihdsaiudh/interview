package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"interview-server/internal/handler"
	"interview-server/internal/middleware"
	"interview-server/internal/repository"
	"interview-server/internal/service"
)

func main() {
	r := gin.Default()

	// CORS 中间件 — 允许前端跨域访问
	r.Use(middleware.CORS())

	// 健康检查
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 初始化分层依赖
	userRepo := repository.NewUserRepo()
	authSvc := service.NewAuthService(userRepo)
	authH := handler.NewAuthHandler(authSvc)

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", handler.Ping)

		// 认证路由（无需登录）
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authH.Register)
			auth.GET("/verify-email", authH.VerifyEmail)
			auth.POST("/login", authH.Login)

			// 需要登录的路由
			authRequired := auth.Group("")
			authRequired.Use(middleware.JWTAuth())
			{
				authRequired.GET("/me", authH.Me)
			}
		}
	}

	log.Println("🚀 Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
