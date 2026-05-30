package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Ping 健康检查接口。
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
