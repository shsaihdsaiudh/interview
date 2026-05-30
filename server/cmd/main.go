package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"interview-server/application"
	"interview-server/infrastructure/persistence"
	"interview-server/interfaces"
	"interview-server/interfaces/handler"
)

func main() {
	// ── 1. 基础设施层：PostgreSQL 连接池 ──
	ctx := context.Background()
	dsn := "postgres://interview:interview123@localhost:5432/interview_platform?sslmode=disable"
	pool := persistence.NewPool(ctx, dsn)
	defer pool.Close()

	// ── 2. 基础设施层：PostgresRepo 同时满足两个领域接口 ──
	repo := persistence.NewPostgresRepo(pool)

	// ── 3. 应用层：注入领域接口（依赖反转 — 应用层不依赖具体实现）──
	userSvc := application.NewUserService(repo, repo)
	apptSvc := application.NewAppointmentService(repo, repo)

	// ── 4. 接口层：HTTP 处理器 ──
	userH := handler.NewUserHandler(userSvc)
	apptH := handler.NewAppointmentHandler(apptSvc)

	// ── 5. 路由注册 ──
	r := gin.Default()
	interfaces.RegisterRoutes(r, userH, apptH)

	// ── 6. 启动服务 ──
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("🚀 Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
