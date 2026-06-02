// Package interfaces 注册所有 HTTP 路由。
package interfaces

import (
	"github.com/gin-gonic/gin"

	"interview-server/interfaces/handler"
	"interview-server/interfaces/middleware"
)

// RegisterRoutes 注册所有 API 路由。
// 接受 handler 实例作为参数 — 依赖由 main.go 注入。
func RegisterRoutes(
	r *gin.Engine,
	userH *handler.UserHandler,
	apptH *handler.AppointmentHandler,
) {
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

		// ── 认证路由（无需登录）──
		auth := v1.Group("/auth")
		{
			auth.POST("/send-code", userH.SendCode)
			auth.POST("/register", userH.Register)
			auth.POST("/login", userH.Login)
			auth.POST("/forgot-password", userH.ForgotPassword)
			auth.POST("/reset-password", userH.ResetPassword)

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
		}
	}
}
