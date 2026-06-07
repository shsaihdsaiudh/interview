package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"interview-server/application"
	domainUser "interview-server/domain/user"
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
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://interview:interview123@localhost:5432/interview_platform?sslmode=disable"
		log.Println("⚠️  [安全警告] DATABASE_URL 环境变量未设置，使用默认开发数据库连接。生产环境务必通过环境变量配置！")
	}
	pool := persistence.NewPool(ctx, dsn)
	defer pool.Close()

	// ── 1.1 自动建表/迁移（幂等，多次启动不会报错）──
	if err := persistence.RunMigrations(ctx, pool); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// ── 2. 基础设施层：PostgresRepo 同时满足两个领域接口 ──
	repo := persistence.NewPostgresRepo(pool)

	// ── 3. 邮件发送器（从环境变量读取凭据，未配置则 nil）──
	mailSender := mail.NewSMTPSenderFromEnv()

	// ── 4. 应用层：注入领域接口（依赖反转 — 应用层不依赖具体实现）──
	userSvc := application.NewUserService(repo, repo, mailSender)
	apptSvc := application.NewAppointmentService(repo, repo)
	recruitSvc := application.NewRecruitmentService(repo, repo)
	adminSvc := application.NewAdminService(repo, repo, repo)

	// ── 5. 接口层：HTTP 处理器 ──
	userH := handler.NewUserHandler(userSvc)
	apptH := handler.NewAppointmentHandler(apptSvc)
	recruitH := handler.NewRecruitmentHandler(recruitSvc)
	adminH := handler.NewAdminHandler(adminSvc)

	// ── 5.1 自动提升管理员（从 ADMIN_EMAIL 环境变量读取）──
	if adminEmail := os.Getenv("ADMIN_EMAIL"); adminEmail != "" {
		if u, err := repo.FindByEmail(adminEmail); err == nil {
			if !u.IsAdmin() {
				u.Role = domainUser.RoleAdmin
				if err := repo.Update(u); err == nil {
					log.Printf("👑 已将 %s 提升为管理员", adminEmail)
				}
			}
		} else {
			log.Printf("⚠️  ADMIN_EMAIL=%s 对应的用户尚未注册，将在注册后自动提升为管理员", adminEmail)
		}
	}

	// ── 6. 路由注册 ──
	r := gin.Default()
	interfaces.RegisterRoutes(r, userH, apptH, recruitH, adminH, repo)

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
