package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"interview-server/application"
	"interview-server/infrastructure/mail"
	"interview-server/infrastructure/persistence"
	"interview-server/interfaces"
	"interview-server/interfaces/handler"
)

func main() {
	// ── 0. 加载 .env 文件（兼容从项目根目录或 server/ 目录启动）──
	_ = godotenv.Load(".env")     // 从项目根目录启动：go run ./server/cmd/main.go
	_ = godotenv.Load("../.env")  // 从 server/ 目录启动：go run ./cmd/main.go

	// ── 1. 基础设施层：PostgreSQL 连接池 ──
	ctx := context.Background()
	dsn := "postgres://interview:interview123@localhost:5432/interview_platform?sslmode=disable"
	pool := persistence.NewPool(ctx, dsn)
	defer pool.Close()

	// ── 2. 基础设施层：PostgresRepo 同时满足两个领域接口 ──
	repo := persistence.NewPostgresRepo(pool)

	// ── 3. 邮件发送器（从环境变量读取凭据，未配置则 nil）──
	mailSender := mail.NewSMTPSenderFromEnv()

	// ── 4. 应用层：注入领域接口（依赖反转 — 应用层不依赖具体实现）──
	userSvc := application.NewUserService(repo, repo, mailSender)
	apptSvc := application.NewAppointmentService(repo, repo)

	// ── 5. 接口层：HTTP 处理器 ──
	userH := handler.NewUserHandler(userSvc)
	apptH := handler.NewAppointmentHandler(apptSvc)

	// ── 6. 路由注册 ──
	r := gin.Default()
	interfaces.RegisterRoutes(r, userH, apptH)

	// ── 7. 启动服务 ──
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
