package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
)

// Mailer handles sending emails
type Mailer struct {
	cfg    *viper.Viper
	dialer *gomail.Dialer
}

// NewMailer creates a new mailer instance
func NewMailer(cfg *viper.Viper) *Mailer {
	var dialer *gomail.Dialer

	if cfg.GetBool("email.enabled") {
		host := cfg.GetString("email.smtp.host")
		port := cfg.GetInt("email.smtp.port")
		username := cfg.GetString("email.smtp.username")
		password := cfg.GetString("email.smtp.password")

		dialer = gomail.NewDialer(host, port, username, password)
		
		// Configure TLS
		if cfg.GetBool("email.smtp.use_tls") {
			dialer.TLSConfig = &tls.Config{
				ServerName:         host,
				InsecureSkipVerify: cfg.GetBool("email.smtp.skip_verify"),
			}
		}

		// Note: gomail.v2 doesn't support setting timeout directly
		// The timeout would need to be handled at a different level
	}

	return &Mailer{
		cfg:    cfg,
		dialer: dialer,
	}
}

// SendEmail sends an email message
func (m *Mailer) SendEmail(email *EmailQueue) error {
	if !m.cfg.GetBool("email.enabled") {
		return fmt.Errorf("email sending is disabled")
	}

	if m.dialer == nil {
		return fmt.Errorf("email dialer not configured")
	}

	message := gomail.NewMessage()

	// Set headers
	message.SetHeader("From", m.formatAddress(email.FromEmail, email.FromName))
	message.SetHeader("To", m.formatAddress(email.ToEmail, email.ToName))
	message.SetHeader("Subject", email.Subject)

	// Set body
	if email.BodyHTML != "" {
		message.SetBody("text/html", email.BodyHTML)
		if email.BodyText != "" {
			message.AddAlternative("text/plain", email.BodyText)
		}
	} else if email.BodyText != "" {
		message.SetBody("text/plain", email.BodyText)
	} else {
		return fmt.Errorf("email body is empty")
	}

	// Add headers for tracking
	message.SetHeader("X-Mailer", "CasGists")
	message.SetHeader("X-Email-Type", string(email.Type))
	message.SetHeader("X-Priority", fmt.Sprintf("%d", email.Priority))

	// Send the email
	if err := m.dialer.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// TestConnection tests the SMTP connection
func (m *Mailer) TestConnection() error {
	if !m.cfg.GetBool("email.enabled") {
		return fmt.Errorf("email is disabled")
	}

	if m.dialer == nil {
		return fmt.Errorf("email dialer not configured")
	}

	// Create a test connection
	closer, err := m.dialer.Dial()
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer closer.Close()

	return nil
}

// SendTestEmail sends a test email
func (m *Mailer) SendTestEmail(toEmail, toName string) error {
	if !m.cfg.GetBool("email.enabled") {
		return fmt.Errorf("email sending is disabled")
	}

	testEmail := &EmailQueue{
		Type:      EmailTypeSystemAlert,
		ToEmail:   toEmail,
		ToName:    toName,
		FromEmail: m.cfg.GetString("email.from_email"),
		FromName:  m.cfg.GetString("email.from_name"),
		Subject:   "CasGists Email Test",
		BodyText:  "This is a test email from CasGists. If you received this, email configuration is working correctly.",
		BodyHTML:  "<p>This is a test email from <strong>CasGists</strong>.</p><p>If you received this, email configuration is working correctly.</p>",
		Priority:  1,
	}

	return m.SendEmail(testEmail)
}

// formatAddress formats an email address with optional name
func (m *Mailer) formatAddress(email, name string) string {
	if name == "" {
		return email
	}
	return fmt.Sprintf("%s <%s>", name, email)
}

// SMTPConfig represents SMTP configuration for validation
type SMTPConfig struct {
	Host       string
	Port       int
	Username   string
	Password   string
	UseTLS     bool
	SkipVerify bool
	Timeout    time.Duration
}

// ValidateSMTPConfig validates SMTP configuration
func ValidateSMTPConfig(config SMTPConfig) error {
	if config.Host == "" {
		return fmt.Errorf("SMTP host is required")
	}

	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("SMTP port must be between 1 and 65535")
	}

	// Test connection
	var auth smtp.Auth
	if config.Username != "" && config.Password != "" {
		auth = smtp.PlainAuth("", config.Username, config.Password, config.Host)
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	
	// Simple connection test
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	// Test TLS if enabled
	if config.UseTLS {
		tlsConfig := &tls.Config{
			ServerName:         config.Host,
			InsecureSkipVerify: config.SkipVerify,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	// Test authentication if credentials provided
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	return nil
}

// EmailProviderTemplates contains common provider configurations
var EmailProviderTemplates = map[string]SMTPConfig{
	"gmail": {
		Host:    "smtp.gmail.com",
		Port:    587,
		UseTLS:  true,
		Timeout: 30 * time.Second,
	},
	"outlook": {
		Host:    "smtp-mail.outlook.com",
		Port:    587,
		UseTLS:  true,
		Timeout: 30 * time.Second,
	},
	"yahoo": {
		Host:    "smtp.mail.yahoo.com",
		Port:    587,
		UseTLS:  true,
		Timeout: 30 * time.Second,
	},
	"sendgrid": {
		Host:    "smtp.sendgrid.net",
		Port:    587,
		UseTLS:  true,
		Timeout: 30 * time.Second,
	},
	"mailgun": {
		Host:    "smtp.mailgun.org",
		Port:    587,
		UseTLS:  true,
		Timeout: 30 * time.Second,
	},
	"ses": {
		Host:    "email-smtp.us-east-1.amazonaws.com",
		Port:    587,
		UseTLS:  true,
		Timeout: 30 * time.Second,
	},
}

// GetProviderConfig returns SMTP configuration for a known provider
func GetProviderConfig(provider string) (SMTPConfig, bool) {
	config, exists := EmailProviderTemplates[strings.ToLower(provider)]
	return config, exists
}