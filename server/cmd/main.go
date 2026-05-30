package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"interview-server/internal/handler"
	"interview-server/internal/middleware"
)

func main() {
	r := gin.Default()

	// CORS 中间件 — 允许前端跨域访问
	r.Use(middleware.CORS())

	// 健康检查
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", handler.Ping)
		// 后续在这里注册更多路由
	}

	log.Println("🚀 Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
