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

	// ── 初始化共享存储 ──
	store := repository.NewStore()

	// ── 初始化仓库层 ──
	userRepo := repository.NewUserRepo(store)
	availRepo := repository.NewAvailabilityRepo(store)
	apptRepo := repository.NewAppointmentRepo(store)

	// ── 初始化服务层 ──
	authSvc := service.NewAuthService(userRepo)
	availSvc := service.NewAvailabilityService(availRepo)
	apptSvc := service.NewAppointmentService(apptRepo, availRepo, userRepo)
	userSvc := service.NewUserService(userRepo, availRepo, apptRepo)

	// ── 初始化处理器层 ──
	authH := handler.NewAuthHandler(authSvc)
	userH := handler.NewUserHandler(userSvc)
	availH := handler.NewAvailabilityHandler(availSvc)
	apptH := handler.NewAppointmentHandler(apptSvc)

	// ── API v1 路由组 ──
	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", handler.Ping)

		// 认证路由（无需登录）
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authH.Register)
			auth.GET("/verify-email", authH.VerifyEmail)
			auth.POST("/login", authH.Login)

			authRequired := auth.Group("")
			authRequired.Use(middleware.JWTAuth())
			{
				authRequired.GET("/me", authH.Me)
			}
		}

		// 用户路由（部分公开）
		v1.GET("/users", userH.ListUsers)
		v1.GET("/users/:id", userH.GetUser)

		// 需要登录的路由
		protected := v1.Group("")
		protected.Use(middleware.JWTAuth())
		{
			// 个人资料
			protected.GET("/profile", userH.GetProfile)
			protected.PUT("/profile", userH.UpdateProfile)

			// 空闲时间管理
			protected.GET("/availability", availH.GetMyAvailability)
			protected.POST("/availability", availH.AddAvailability)
			protected.DELETE("/availability/:id", availH.DeleteAvailability)

			// 预约管理
			protected.POST("/appointments", apptH.CreateAppointment)
			protected.GET("/appointments", apptH.GetMyAppointments)
			protected.PUT("/appointments/:id/accept", apptH.AcceptAppointment)
			protected.PUT("/appointments/:id/reject", apptH.RejectAppointment)
		}
	}

	log.Println("🚀 Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
