// Package interfaces 注册所有 HTTP 路由。
package interfaces

import (
	"time"

	"github.com/gin-gonic/gin"

	domainUser "interview-server/domain/user"
	"interview-server/interfaces/handler"
	"interview-server/interfaces/middleware"
)

// RegisterRoutes 注册所有 API 路由。
// 接受 handler 实例作为参数 — 依赖由 main.go 注入。
func RegisterRoutes(
	r *gin.Engine,
	userH *handler.UserHandler,
	apptH *handler.AppointmentHandler,
	recruitH *handler.RecruitmentHandler,
	adminH *handler.AdminHandler,
	userRepo domainUser.UserRepository,
) {
	// ── 速率限制器（基于 IP 的滑动窗口）──
	limiter := middleware.NewRateLimiter()

	// ── 静态文件服务：映射 /uploads/ 到 server/uploads/ 目录 ──
	r.Static("/uploads", "./server/uploads")

	// CORS 中间件
	r.Use(middleware.CORS())

	// 健康检查
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", handler.Ping)

		// ── 认证路由（无需登录，已应用速率限制防暴力破解）──
		auth := v1.Group("/auth")
		{
			// 发送验证码：每 IP 每分钟最多 2 次（防邮件轰炸）
			auth.POST("/send-code", limiter.Limit(2, time.Minute), userH.SendCode)
			// 注册：每 IP 每分钟最多 10 次（防暴力破解验证码）
			auth.POST("/register", limiter.Limit(10, time.Minute), userH.Register)
			// 登录：每 IP 每分钟最多 20 次（防暴力破解密码）
			auth.POST("/login", limiter.Limit(20, time.Minute), userH.Login)
			// 忘记密码：每 IP 每分钟最多 2 次（防邮件轰炸）
			auth.POST("/forgot-password", limiter.Limit(2, time.Minute), userH.ForgotPassword)
			// 重置密码：每 IP 每分钟最多 5 次（防暴力破解）
			auth.POST("/reset-password", limiter.Limit(5, time.Minute), userH.ResetPassword)

			authRequired := auth.Group("")
			authRequired.Use(middleware.JWTAuth())
			{
				authRequired.GET("/me", userH.Me)
				authRequired.PUT("/change-password", userH.ChangePassword)
				authRequired.DELETE("/account", userH.DeleteAccount)
			}
		}

		// ── 用户路由（部分公开）──
		v1.GET("/users", userH.ListUsers)
		v1.GET("/users/:id", middleware.OptionalJWTAuth(), userH.GetUser)

		// ── 招募卡片路由（部分公开）──
		v1.GET("/recruitment-card", recruitH.GetCardByUserID)
		v1.GET("/recruitment-cards", recruitH.ListCards)

		// ── 需要登录的路由 ──
		protected := v1.Group("")
		protected.Use(middleware.JWTAuth())
		{
			// 个人资料
			protected.GET("/profile", userH.GetProfile)
			protected.PUT("/profile", userH.UpdateProfile)
			protected.POST("/profile/avatar", userH.UploadAvatar)

			// 空闲时间管理
			protected.GET("/availability", apptH.GetMyAvailability)
			protected.POST("/availability", apptH.AddAvailability)
			protected.DELETE("/availability/:id", apptH.DeleteAvailability)

			// 预约管理
			protected.POST("/appointments", apptH.CreateAppointment)
			protected.GET("/appointments", apptH.GetMyAppointments)
			protected.PUT("/appointments/:id/accept", apptH.AcceptAppointment)
			protected.PUT("/appointments/:id/reject", apptH.RejectAppointment)

			// 招募卡片管理
			protected.PUT("/recruitment-card", recruitH.CreateOrUpdateCard)

			// ── 管理员路由 ──
			admin := v1.Group("/admin")
			admin.Use(middleware.AdminOnly(userRepo))
			{
				admin.GET("/stats", adminH.GetStats)

				admin.GET("/users", adminH.ListUsers)
				admin.PUT("/users/:email/role", adminH.UpdateUserRole)
				admin.PUT("/users/:email/ban", adminH.BanUser)
				admin.PUT("/users/:email/unban", adminH.UnbanUser)

				admin.GET("/cards", adminH.ListCards)
				admin.DELETE("/cards/:id", adminH.DeleteCard)

				admin.GET("/appointments", adminH.ListAppointments)
				admin.DELETE("/appointments/:id", adminH.DeleteAppointment)
			}
		}
	}
}
