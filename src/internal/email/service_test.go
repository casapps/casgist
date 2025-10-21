package email

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupEmailTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err)

	// For SQLite testing, we need to manually create tables with TEXT for UUID fields
	db.Exec(`CREATE TABLE email_queues (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		to_email TEXT NOT NULL,
		to_name TEXT,
		from_email TEXT NOT NULL,
		from_name TEXT,
		subject TEXT NOT NULL,
		body_text TEXT,
		body_html TEXT,
		priority INTEGER DEFAULT 5,
		status TEXT DEFAULT 'pending',
		attempts INTEGER DEFAULT 0,
		max_attempts INTEGER DEFAULT 3,
		error TEXT,
		scheduled_at DATETIME,
		sent_at DATETIME,
		created_at DATETIME,
		updated_at DATETIME
	)`)

	db.Exec(`CREATE TABLE email_preferences (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL UNIQUE,
		email_verified BOOLEAN DEFAULT FALSE,
		notify_gist_starred BOOLEAN DEFAULT TRUE,
		notify_gist_forked BOOLEAN DEFAULT TRUE,
		notify_gist_commented BOOLEAN DEFAULT TRUE,
		notify_followed BOOLEAN DEFAULT TRUE,
		notify_weekly_digest BOOLEAN DEFAULT FALSE,
		notify_system_alerts BOOLEAN DEFAULT TRUE,
		created_at DATETIME,
		updated_at DATETIME
	)`)

	db.Exec(`CREATE TABLE email_verification_tokens (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		email TEXT NOT NULL,
		token TEXT NOT NULL UNIQUE,
		used BOOLEAN DEFAULT FALSE,
		expires_at DATETIME,
		used_at DATETIME,
		created_at DATETIME
	)`)

	db.Exec(`CREATE TABLE password_reset_tokens (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		email TEXT NOT NULL,
		token TEXT NOT NULL UNIQUE,
		used BOOLEAN DEFAULT FALSE,
		expires_at DATETIME,
		used_at DATETIME,
		created_at DATETIME
	)`)

	return db
}

func TestEmailService(t *testing.T) {
	db := setupEmailTestDB(t)
	cfg := viper.New()
	cfg.Set("email.enabled", true)
	cfg.Set("email.from_email", "test@example.com")
	cfg.Set("email.from_name", "Test Service")
	cfg.Set("server.url", "https://test.example.com")

	service := NewService(db, cfg)
	require.NotNil(t, service)

	userID := uuid.New()

	t.Run("QueueEmail", func(t *testing.T) {
		email := &EmailQueue{
			Type:      EmailTypeWelcome,
			ToEmail:   "user@example.com",
			ToName:    "Test User",
			FromEmail: "test@example.com",
			FromName:  "Test Service",
			Subject:   "Welcome!",
			BodyText:  "Welcome to the service",
			BodyHTML:  "<h1>Welcome to the service</h1>",
			Priority:  1,
		}

		err := service.QueueEmail(email)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, email.ID)
		assert.Equal(t, EmailStatusPending, email.Status)
	})

	t.Run("CreateEmailPreference", func(t *testing.T) {
		err := service.CreateEmailPreference(userID)
		assert.NoError(t, err)

		pref, err := service.GetEmailPreference(userID)
		assert.NoError(t, err)
		assert.Equal(t, userID, pref.UserID)
		assert.True(t, pref.NotifyGistStarred)
		assert.True(t, pref.NotifyFollowed)
		assert.False(t, pref.NotifyWeeklyDigest)
	})

	t.Run("UpdateEmailPreference", func(t *testing.T) {
		updates := map[string]bool{
			"notify_gist_starred":  false,
			"notify_weekly_digest": true,
		}

		err := service.UpdateEmailPreference(userID, updates)
		assert.NoError(t, err)

		pref, err := service.GetEmailPreference(userID)
		assert.NoError(t, err)
		assert.False(t, pref.NotifyGistStarred)
		assert.True(t, pref.NotifyWeeklyDigest)

		// Restore starred notifications for subsequent tests
		restore := map[string]bool{"notify_gist_starred": true}
		require.NoError(t, service.UpdateEmailPreference(userID, restore))
	})

	t.Run("SendVerificationEmail", func(t *testing.T) {
		err := service.SendVerificationEmail(userID, "test@example.com", "testuser")
		assert.NoError(t, err)

		// Check that token was created
		var token EmailVerificationToken
		err = db.Where("user_id = ?", userID).First(&token).Error
		assert.NoError(t, err)
		assert.False(t, token.Used)
		assert.True(t, token.ExpiresAt.After(time.Now()))

		// Check that email was queued
		var email EmailQueue
		err = db.Where("type = ? AND to_email = ?", EmailTypeVerification, "test@example.com").First(&email).Error
		assert.NoError(t, err)
		assert.Equal(t, EmailStatusPending, email.Status)
		assert.Contains(t, email.BodyHTML, token.Token)
	})

	t.Run("VerifyEmail", func(t *testing.T) {
		// Create a verification token
		token := &EmailVerificationToken{
			UserID:    userID,
			Email:     "test@example.com",
			Token:     "test-token-123",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		require.NoError(t, db.Create(token).Error)

		err := service.VerifyEmail("test-token-123")
		assert.NoError(t, err)

		// Check that token was marked as used
		var updatedToken EmailVerificationToken
		err = db.Where("token = ?", "test-token-123").First(&updatedToken).Error
		assert.NoError(t, err)
		assert.True(t, updatedToken.Used)
		assert.NotNil(t, updatedToken.UsedAt)
	})

	t.Run("SendPasswordResetEmail", func(t *testing.T) {
		err := service.SendPasswordResetEmail(userID, "test@example.com", "testuser")
		assert.NoError(t, err)

		// Check that token was created
		var token PasswordResetToken
		err = db.Where("user_id = ?", userID).First(&token).Error
		assert.NoError(t, err)
		assert.False(t, token.Used)
		assert.True(t, token.ExpiresAt.After(time.Now()))

		// Check that email was queued
		var email EmailQueue
		err = db.Where("type = ? AND to_email = ?", EmailTypePasswordReset, "test@example.com").First(&email).Error
		assert.NoError(t, err)
		assert.Equal(t, EmailStatusPending, email.Status)
	})

	t.Run("VerifyPasswordResetToken", func(t *testing.T) {
		// Create a password reset token
		token := &PasswordResetToken{
			UserID:    userID,
			Email:     "test@example.com",
			Token:     "reset-token-123",
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		require.NoError(t, db.Create(token).Error)

		verifiedToken, err := service.VerifyPasswordResetToken("reset-token-123")
		assert.NoError(t, err)
		assert.Equal(t, userID, verifiedToken.UserID)

		// Use the token
		err = service.UsePasswordResetToken("reset-token-123")
		assert.NoError(t, err)

		// Check that token was marked as used
		var updatedToken PasswordResetToken
		err = db.Where("token = ?", "reset-token-123").First(&updatedToken).Error
		assert.NoError(t, err)
		assert.True(t, updatedToken.Used)
		assert.NotNil(t, updatedToken.UsedAt)
	})

	t.Run("SendGistStarredNotification", func(t *testing.T) {
		gistID := uuid.New()
		err := service.SendGistStarredNotification(
			userID,
			"recipient@example.com",
			"Recipient Name",
			"Actor Name",
			"My Awesome Gist",
			gistID,
		)
		assert.NoError(t, err)

		// Check that email was queued
		var email EmailQueue
		err = db.Where("type = ? AND to_email = ?", EmailTypeGistStarred, "recipient@example.com").First(&email).Error
		assert.NoError(t, err)
		assert.Equal(t, EmailStatusPending, email.Status)
		assert.Contains(t, email.BodyHTML, "Actor Name")
		assert.Contains(t, email.BodyHTML, "My Awesome Gist")
	})

	t.Run("ProcessEmailQueue", func(t *testing.T) {
		// Create a test email
		email := &EmailQueue{
			Type:      EmailTypeSystemAlert,
			ToEmail:   "test@example.com",
			FromEmail: "system@example.com",
			Subject:   "Test Email",
			BodyText:  "This is a test",
			Priority:  1,
			Status:    EmailStatusPending,
		}
		require.NoError(t, service.QueueEmail(email))

		ctx := context.Background()
		err := service.ProcessEmailQueue(ctx)

		// Should not error even if SMTP is not configured (it will fail to send but handle the error)
		assert.NoError(t, err)

		// Check that email status was updated
		var updatedEmail EmailQueue
		err = db.First(&updatedEmail, email.ID).Error
		assert.NoError(t, err)
		assert.Greater(t, updatedEmail.Attempts, 0)
		// Without SMTP configured the mailer will fail; depending on retry policy the status
		// may remain pending while a retry is scheduled. Accept either outcome.
		if updatedEmail.Status == EmailStatusPending {
			assert.NotNil(t, updatedEmail.ScheduledAt)
		} else {
			assert.Equal(t, EmailStatusFailed, updatedEmail.Status)
		}
	})
}

func TestTemplateRenderer(t *testing.T) {
	renderer := NewTemplateRenderer()
	require.NotNil(t, renderer)

	err := renderer.LoadDefaultTemplates()
	assert.NoError(t, err)

	t.Run("RenderHTML", func(t *testing.T) {
		data := EmailData{
			"UserName":        "Test User",
			"VerificationURL": "https://example.com/verify?token=abc123",
			"ExpiresAt":       "January 1, 2024 at 12:00 PM",
		}

		html, err := renderer.RenderHTML(EmailTypeVerification, data)
		assert.NoError(t, err)
		assert.Contains(t, html, "Test User")
		assert.Contains(t, html, "https://example.com/verify?token=abc123")
		assert.Contains(t, html, "January 1, 2024 at 12:00 PM")
	})

	t.Run("RenderText", func(t *testing.T) {
		data := EmailData{
			"UserName":        "Test User",
			"VerificationURL": "https://example.com/verify?token=abc123",
			"ExpiresAt":       "January 1, 2024 at 12:00 PM",
		}

		text, err := renderer.RenderText(EmailTypeVerification, data)
		assert.NoError(t, err)
		assert.Contains(t, text, "Test User")
		assert.Contains(t, text, "https://example.com/verify?token=abc123")
		assert.Contains(t, text, "January 1, 2024 at 12:00 PM")
	})

	t.Run("RenderNonexistentTemplate", func(t *testing.T) {
		_, err := renderer.RenderHTML("nonexistent", EmailData{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "template not found")
	})
}

func TestEmailPriority(t *testing.T) {
	cfg := viper.New()
	service := &Service{cfg: cfg}

	tests := []struct {
		emailType EmailType
		expected  int
	}{
		{EmailTypeVerification, 1},
		{EmailTypePasswordReset, 1},
		{EmailTypeWelcome, 2},
		{EmailTypeSystemAlert, 3},
		{EmailTypeGistStarred, 5},
		{EmailTypeUserFollowed, 5},
		{EmailTypeWeeklyDigest, 8},
		{EmailType("unknown"), 5},
	}

	for _, test := range tests {
		priority := service.getEmailPriority(test.emailType)
		assert.Equal(t, test.expected, priority, "Priority mismatch for %s", test.emailType)
	}
}

func TestRenderSubject(t *testing.T) {
	data := EmailData{
		"UserName":  "John Doe",
		"GistTitle": "My Cool Script",
	}

	t.Run("SimpleSubject", func(t *testing.T) {
		subject, err := RenderSubject("Welcome {{.UserName}}!", data)
		assert.NoError(t, err)
		assert.Equal(t, "Welcome John Doe!", subject)
	})

	t.Run("InvalidTemplate", func(t *testing.T) {
		subject, err := RenderSubject("Invalid {{.NonExistent", data)
		assert.NoError(t, err) // Should return original subject
		assert.Equal(t, "Invalid {{.NonExistent", subject)
	})

	t.Run("TemplateExecutionError", func(t *testing.T) {
		subject, err := RenderSubject("{{.NonExistent}}", data)
		assert.NoError(t, err)                 // Should return original subject
		assert.Equal(t, "<no value>", subject) // Go template renders missing values as <no value>
	})
}
