// Package mail 提供邮件发送能力。
// 当前通过学校 SMTP 服务器（smtp.std.uestc.edu.cn:465 SSL）发送。
// 未配置凭据时自动降级为仅打印日志。
package mail

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
)

// Sender 邮件发送接口 — 便于测试 mock。
type Sender interface {
	SendVerificationCode(to, code string) error
}

// SMTPSender 通过 SMTP SSL (端口 465) 发送邮件。
type SMTPSender struct {
	host          string
	port          string
	username      string
	password      string
	from          string
	tlsServerName string // TLS 证书校验用的域名，默认同 SMTP host；若证书与实际 host 不匹配时通过 SMTP_TLS_SERVERNAME 覆盖
}

// NewSMTPSender 创建 SMTP 发送器。
func NewSMTPSender(host, port, username, password string) *SMTPSender {
	return &SMTPSender{
		host:          host,
		port:          port,
		username:      username,
		password:      password,
		from:          username, // 发件人即自己的邮箱
		tlsServerName: host,     // 默认用 SMTP host 做 TLS 校验
	}
}

// NewSMTPSenderFromEnv 从环境变量创建。
// SMTP_USER 和 SMTP_PASS 都设置时才创建；缺一则返回 nil（降级为日志模式）。
// 返回 Sender 接口而非具体指针，避免 nil 指针包装在非 nil 接口中导致 nil dereference。
func NewSMTPSenderFromEnv() Sender {
	host := envOrDefault("SMTP_HOST", "smtp.std.uestc.edu.cn")
	port := envOrDefault("SMTP_PORT", "465")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	if user == "" || pass == "" {
		log.Println("📧 SMTP 凭据未设置（SMTP_USER / SMTP_PASS），验证码仅打印到日志")
		return nil
	}

	sender := NewSMTPSender(host, port, user, pass)

	// 允许覆盖 TLS 证书校验域名（当 SMTP 服务器证书与实际 host 不一致时使用）
	// 例如学校邮箱底层是 icoremail，证书匹配 *.icoremail.net，需设置 SMTP_TLS_SERVERNAME=icoremail.net
	if sn := os.Getenv("SMTP_TLS_SERVERNAME"); sn != "" {
		sender.tlsServerName = sn
	}

	return sender
}

// SendVerificationCode 发送验证码邮件。
func (s *SMTPSender) SendVerificationCode(to, code string) error {
	addr := net.JoinHostPort(s.host, s.port)

	// 端口 465 使用隐式 TLS（先 TLS 握手，再 SMTP）
	tlsConfig := &tls.Config{
		ServerName: s.tlsServerName,
		MinVersion: tls.VersionTLS12,
	}
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS 连接失败: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("SMTP 客户端创建失败: %w", err)
	}
	defer client.Quit()

	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP 认证失败: %w", err)
	}

	if err := client.Mail(s.from); err != nil {
		return fmt.Errorf("MAIL FROM 失败: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO 失败: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA 失败: %w", err)
	}

	msg := buildVerificationEmail(s.from, to, code)
	if _, err := w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("写入邮件内容失败: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("关闭邮件数据失败: %w", err)
	}

	log.Printf("📧 验证码已发送到 %s", to)
	return nil
}

// rfc2047B 对 UTF-8 字符串做 RFC 2047 B 编码（Base64）。
// 产出格式：=?UTF-8?B?<base64>?=
func rfc2047B(s string) string {
	return "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(s)) + "?="
}

// buildVerificationEmail 构建验证码邮件内容。
func buildVerificationEmail(from, to, code string) string {
	return fmt.Sprintf(
		"From: %s <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			`<div style="max-width:480px;margin:0 auto;padding:32px 24px;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;">
  <div style="text-align:center;margin-bottom:24px;">
    <div style="display:inline-block;width:48px;height:48px;border-radius:12px;background:#6366f1;color:#fff;font-size:24px;line-height:48px;">🎯</div>
  </div>
  <h2 style="text-align:center;color:#1e293b;margin:0 0 8px;">邮箱验证码</h2>
  <p style="text-align:center;color:#64748b;font-size:14px;margin:0 0 24px;">你正在注册面试互助平台，请在注册页面输入以下验证码</p>
  <div style="background:#f8fafc;border-radius:12px;padding:24px;text-align:center;margin-bottom:24px;">
    <div style="font-size:36px;font-weight:700;letter-spacing:8px;color:#6366f1;font-family:'SF Mono',Menlo,monospace;">%s</div>
  </div>
  <p style="text-align:center;color:#94a3b8;font-size:12px;margin:0;">验证码 5 分钟内有效，请勿转发给他人</p>
  <hr style="border:none;border-top:1px solid #e2e8f0;margin:24px 0;">
  <p style="text-align:center;color:#cbd5e1;font-size:11px;margin:0;">面试互助平台 · 电子科技大学</p>
</div>`,
		rfc2047B("面试互助平台"),
		from,
		to,
		rfc2047B("邮箱验证码"),
		code,
	)
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
