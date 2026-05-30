package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateJWT(t *testing.T) {
	token, err := GenerateJWT("alice@school.edu")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}

	// 解析验证
	parsed, err := ParseJWT(token)
	if err != nil {
		t.Fatalf("failed to parse generated token: %v", err)
	}
	if parsed != "alice@school.edu" {
		t.Errorf("email = %q, want %q", parsed, "alice@school.edu")
	}
}

func TestParseJWT(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		token, _ := GenerateJWT("bob@school.edu")
		email, err := ParseJWT(token)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if email != "bob@school.edu" {
			t.Errorf("email = %q, want bob@school.edu", email)
		}
	})

	t.Run("invalid token string", func(t *testing.T) {
		_, err := ParseJWT("not.a.valid.token")
		if err == nil {
			t.Error("expected error for invalid token")
		}
	})

	t.Run("empty token", func(t *testing.T) {
		_, err := ParseJWT("")
		if err == nil {
			t.Error("expected error for empty token")
		}
	})

	t.Run("wrong signing method", func(t *testing.T) {
		// 用不同的签名方法生成 token
		claims := jwt.MapClaims{
			"email": "alice@school.edu",
			"exp":   time.Now().Add(24 * time.Hour).Unix(),
			"iat":   time.Now().Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
		// SigningMethodNone 不需要 key
		tokenStr, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

		_, err := ParseJWT(tokenStr)
		if err == nil {
			t.Error("expected error for unsupported signing method")
		}
	})

	t.Run("expired token", func(t *testing.T) {
		// 手动创建一个已过期的 token
		claims := jwt.MapClaims{
			"email": "alice@school.edu",
			"exp":   time.Now().Add(-1 * time.Hour).Unix(),
			"iat":   time.Now().Add(-2 * time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, err := token.SignedString(jwtSecret)
		if err != nil {
			t.Fatalf("failed to create expired token: %v", err)
		}

		_, err = ParseJWT(tokenStr)
		if err == nil {
			t.Error("expected error for expired token")
		}
	})

	t.Run("missing email field", func(t *testing.T) {
		claims := jwt.MapClaims{
			"exp": time.Now().Add(24 * time.Hour).Unix(),
			"iat": time.Now().Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, err := token.SignedString(jwtSecret)
		if err != nil {
			t.Fatalf("failed to create token: %v", err)
		}

		_, err = ParseJWT(tokenStr)
		if err == nil {
			t.Error("expected error for missing email")
		}
	})

	t.Run("wrong secret", func(t *testing.T) {
		// 用不同的 secret 签名
		claims := jwt.MapClaims{
			"email": "alice@school.edu",
			"exp":   time.Now().Add(24 * time.Hour).Unix(),
			"iat":   time.Now().Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, err := token.SignedString([]byte("wrong-secret-key"))
		if err != nil {
			t.Fatalf("failed to create token: %v", err)
		}

		_, err = ParseJWT(tokenStr)
		if err == nil {
			t.Error("expected error for wrong secret")
		}
	})
}
