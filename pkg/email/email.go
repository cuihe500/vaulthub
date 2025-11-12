package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/cuihe500/vaulthub/pkg/errors"
	"github.com/cuihe500/vaulthub/pkg/logger"
)

// Config SMTP配置
type Config struct {
	Host     string // SMTP服务器地址
	Port     int    // SMTP服务器端口
	Username string // SMTP用户名
	Password string // SMTP密码
	From     string // 发件人邮箱
	FromName string // 发件人名称
	UseTLS   bool   // 是否使用TLS加密
}

// Sender 邮件发送器
type Sender struct {
	config *Config
}

// NewSender 创建邮件发送器
func NewSender(config *Config) *Sender {
	return &Sender{
		config: config,
	}
}

// SendMail 发送邮件
// to: 收件人邮箱列表
// subject: 邮件主题
// body: 邮件正文（支持HTML）
func (s *Sender) SendMail(to []string, subject, body string) error {
	if len(to) == 0 {
		return errors.New(errors.CodeInvalidEmail, "收件人列表不能为空")
	}

	// 构建邮件内容
	from := s.config.From
	if s.config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", s.config.FromName, s.config.From)
	}

	// 邮件头
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = strings.Join(to, ",")
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// 组装完整邮件
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// 发送邮件
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	var err error
	if s.config.UseTLS {
		err = s.sendMailTLS(addr, auth, s.config.From, to, []byte(message))
	} else {
		err = smtp.SendMail(addr, auth, s.config.From, to, []byte(message))
	}

	if err != nil {
		logger.Error("发送邮件失败",
			logger.Strings("to", to),
			logger.String("subject", subject),
			logger.Err(err),
		)
		return errors.Wrap(errors.CodeEmailSendFailed, err)
	}

	logger.Info("邮件发送成功",
		logger.Strings("to", to),
		logger.String("subject", subject),
	)

	return nil
}

// sendMailTLS 使用TLS加密发送邮件
func (s *Sender) sendMailTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// 连接到SMTP服务器
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName:         s.config.Host,
		InsecureSkipVerify: false, // 生产环境应设为false，验证服务器证书
	})
	if err != nil {
		return fmt.Errorf("TLS连接失败: %w", err)
	}
	defer func() {
		_ = conn.Close() // 连接关闭失败不影响邮件发送结果
	}()

	// 创建SMTP客户端
	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("创建SMTP客户端失败: %w", err)
	}
	defer func() {
		_ = client.Quit() // 客户端退出失败不影响邮件发送结果
	}()

	// 认证
	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP认证失败: %w", err)
		}
	}

	// 设置发件人
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}

	// 设置收件人
	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return fmt.Errorf("设置收件人失败: %w", err)
		}
	}

	// 发送邮件内容
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("获取数据写入器失败: %w", err)
	}
	_, err = writer.Write(msg)
	if err != nil {
		return fmt.Errorf("写入邮件内容失败: %w", err)
	}
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("关闭数据写入器失败: %w", err)
	}

	return nil
}

// SendVerificationCode 发送验证码邮件
// to: 收件人邮箱
// code: 验证码
// purpose: 用途（注册/登录/重置密码等）
func (s *Sender) SendVerificationCode(to, code, purpose string) error {
	subject := fmt.Sprintf("VaultHub - %s验证码", purpose)
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4CAF50; color: white; padding: 10px; text-align: center; }
        .content { background-color: #f9f9f9; padding: 20px; border-radius: 5px; margin-top: 20px; }
        .code { font-size: 32px; font-weight: bold; color: #4CAF50; text-align: center; letter-spacing: 5px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>VaultHub 密钥管理系统</h2>
        </div>
        <div class="content">
            <p>您好，</p>
            <p>您正在进行<strong>%s</strong>操作，验证码为：</p>
            <div class="code">%s</div>
            <p>验证码有效期为 <strong>5分钟</strong>，请尽快使用。</p>
            <p>如果这不是您本人的操作，请忽略此邮件。</p>
        </div>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿回复。</p>
            <p>&copy; 2024 VaultHub. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, purpose, code)

	return s.SendMail([]string{to}, subject, body)
}

// SendPasswordResetLink 发送密码重置链接邮件
// to: 收件人邮箱
// resetURL: 重置密码链接
// expiryMinutes: 链接有效期（分钟）
func (s *Sender) SendPasswordResetLink(to, resetURL string, expiryMinutes int) error {
	subject := "VaultHub - 密码重置"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4CAF50; color: white; padding: 10px; text-align: center; }
        .content { background-color: #f9f9f9; padding: 20px; border-radius: 5px; margin-top: 20px; }
        .button { display: inline-block; padding: 12px 24px; background-color: #4CAF50; color: white; text-decoration: none; border-radius: 4px; margin: 20px 0; }
        .button:hover { background-color: #45a049; }
        .warning { background-color: #fff3cd; border-left: 4px solid #ffc107; padding: 10px; margin: 15px 0; }
        .footer { text-align: center; margin-top: 20px; font-size: 12px; color: #666; }
        .link { word-break: break-all; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>VaultHub 密钥管理系统</h2>
        </div>
        <div class="content">
            <p>您好，</p>
            <p>我们收到了您的密码重置请求。请点击下面的按钮重置您的密码：</p>
            <div style="text-align: center;">
                <a href="%s" class="button">重置密码</a>
            </div>
            <p>或者复制以下链接到浏览器打开：</p>
            <p class="link">%s</p>
            <div class="warning">
                <p style="margin: 0;"><strong>注意：</strong></p>
                <ul style="margin: 5px 0;">
                    <li>此链接有效期为 <strong>%d分钟</strong></li>
                    <li>链接仅可使用一次</li>
                    <li>如果这不是您本人的操作，请忽略此邮件并确保账户安全</li>
                </ul>
            </div>
        </div>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿回复。</p>
            <p>&copy; 2024 VaultHub. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, resetURL, resetURL, expiryMinutes)

	return s.SendMail([]string{to}, subject, body)
}
