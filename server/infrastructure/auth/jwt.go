// Package auth 提供 JWT token 的生成与解析。
// 属于基础设施层 — 领域层不需要知道 token 的实现细节。
package auth

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwtSecret JWT 签名密钥，通过环境变量 JWT_SECRET 注入。
// 生产环境必须设置强随机密钥（至少 32 字节）；未设置时使用开发默认值并打印警告。
var jwtSecret []byte

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "interview-platform-dev-secret-key--change-in-production!!"
		log.Println("⚠️  [安全警告] JWT_SECRET 环境变量未设置，使用不安全的默认密钥！生产环境务必设置 JWT_SECRET。")
	}
	jwtSecret = []byte(secret)
}

// GenerateJWT 生成 JWT token（24 小时有效）。
func GenerateJWT(email string) (string, error) {
	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseJWT 解析 JWT token，返回用户邮箱。
func ParseJWT(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("不支持的签名方法: %v", t.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", errors.New("无效的 token")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return "", errors.New("token 中缺少 email 字段")
	}

	return email, nil
}
