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
	SendResetCode(to, code string) error
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

// sendMail 发送一封邮件（TLS + SMTP 公共逻辑）。
func (s *SMTPSender) sendMail(to, subject, htmlBody string) error {
	addr := net.JoinHostPort(s.host, s.port)

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

	msg := fmt.Sprintf(
		"From: %s <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s",
		rfc2047B("面试互助平台"), s.from, to, rfc2047B(subject), htmlBody,
	)
	if _, err := w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("写入邮件内容失败: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("关闭邮件数据失败: %w", err)
	}

	log.Printf("📧 邮件已发送到 %s", to)
	return nil
}

// SendVerificationCode 发送注册验证码邮件。
func (s *SMTPSender) SendVerificationCode(to, code string) error {
	return s.sendMail(to, "邮箱验证码", buildVerificationHTML(code))
}

// SendResetCode 发送密码重置验证码邮件。
func (s *SMTPSender) SendResetCode(to, code string) error {
	return s.sendMail(to, "重置密码验证码", buildResetPasswordHTML(code))
}

// rfc2047B 对 UTF-8 字符串做 RFC 2047 B 编码（Base64）。
// 产出格式：=?UTF-8?B?<base64>?=
func rfc2047B(s string) string {
	return "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(s)) + "?="
}

// emailWrapper 邮件外层布局（素净极简风格）。
const emailWrapper = `
<div style="max-width:420px;margin:0 auto;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI','PingFang SC','Hiragino Sans GB',sans-serif;padding:48px 24px;">
  <div style="text-align:center;margin-bottom:40px;">
    <div style="font-size:13px;font-weight:500;color:#1a1a1a;letter-spacing:0.3px;">面试互助平台</div>
  </div>
  %s
  <div style="border-top:1px solid #eee;margin-top:40px;padding-top:20px;">
    <p style="font-size:11px;color:#bbb;text-align:center;line-height:1.8;margin:0;">
      电子科技大学 · 面试互助平台<br>此邮件由系统自动发送，请勿回复
    </p>
  </div>
</div>`

// buildVerificationHTML 构建注册验证码邮件 HTML。
func buildVerificationHTML(code string) string {
	body := fmt.Sprintf(`
  <h1 style="font-size:22px;font-weight:600;color:#1a1a1a;margin:0 0 8px;text-align:center;letter-spacing:-0.3px;">验证你的邮箱</h1>
  <p style="font-size:14px;color:#888;line-height:1.8;margin:0 0 36px;text-align:center;">请输入以下验证码来完成注册。</p>

  <table width="100%%" cellpadding="0" cellspacing="0" style="margin-bottom:36px;">
    <tr>
      <td align="center" style="padding:28px 0;border:1px solid #eaeaea;border-radius:6px;">
        <span style="font-size:32px;font-weight:500;letter-spacing:14px;color:#1a1a1a;font-family:'SF Mono','SFMono-Regular',Menlo,monospace;padding-left:14px;">%s</span>
      </td>
    </tr>
  </table>

  <p style="font-size:12px;color:#aaa;text-align:center;line-height:1.8;margin:0;">
    验证码 5 分钟内有效<br>如果不是你本人操作，请忽略此邮件
  </p>`, code)

	return fmt.Sprintf(emailWrapper, body)
}

// buildResetPasswordHTML 构建密码重置验证码邮件 HTML。
func buildResetPasswordHTML(code string) string {
	body := fmt.Sprintf(`
  <h1 style="font-size:22px;font-weight:600;color:#1a1a1a;margin:0 0 8px;text-align:center;letter-spacing:-0.3px;">重置登录密码</h1>
  <p style="font-size:14px;color:#888;line-height:1.8;margin:0 0 36px;text-align:center;">你发起了密码重置请求，请输入以下验证码来设置新密码。</p>

  <table width="100%%" cellpadding="0" cellspacing="0" style="margin-bottom:36px;">
    <tr>
      <td align="center" style="padding:28px 0;border:1px solid #eaeaea;border-radius:6px;">
        <span style="font-size:32px;font-weight:500;letter-spacing:14px;color:#1a1a1a;font-family:'SF Mono','SFMono-Regular',Menlo,monospace;padding-left:14px;">%s</span>
      </td>
    </tr>
  </table>

  <p style="font-size:12px;color:#aaa;text-align:center;line-height:1.8;margin:0;">
    验证码 5 分钟内有效<br>如果你没有发起此操作，请忽略此邮件，你的账号是安全的
  </p>`, code)

	return fmt.Sprintf(emailWrapper, body)
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
