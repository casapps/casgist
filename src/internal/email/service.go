package email

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// Service handles email operations
type Service struct {
	db       *gorm.DB
	cfg      *viper.Viper
	mailer   *Mailer
	renderer *TemplateRenderer
}

// NewService creates a new email service
func NewService(db *gorm.DB, cfg *viper.Viper) *Service {
	service := &Service{
		db:       db,
		cfg:      cfg,
		mailer:   NewMailer(cfg),
		renderer: NewTemplateRenderer(),
	}

	// Load default templates
	if err := service.renderer.LoadDefaultTemplates(); err != nil {
		log.Printf("Failed to load default email templates: %v", err)
	}

	return service
}

// QueueEmail queues an email for sending
func (s *Service) QueueEmail(email *EmailQueue) error {
	if email.ID == uuid.Nil {
		email.ID = uuid.New()
	}

	if email.Status == "" {
		email.Status = EmailStatusPending
	}

	if email.Priority == 0 {
		email.Priority = 5 // Default priority
	}

	if email.MaxAttempts == 0 {
		email.MaxAttempts = 3
	}

	return s.db.Create(email).Error
}

// SendVerificationEmail sends email verification
func (s *Service) SendVerificationEmail(userID uuid.UUID, email, username string) error {
	// Generate verification token
	token, err := s.generateToken()
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	// Save token to database
	verificationToken := &EmailVerificationToken{
		UserID:    userID,
		Email:     email,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.db.Create(verificationToken).Error; err != nil {
		return fmt.Errorf("failed to save verification token: %w", err)
	}

	// Prepare email data
	data := EmailData{
		"UserName":        username,
		"VerificationURL": fmt.Sprintf("%s/verify-email?token=%s", s.cfg.GetString("server.url"), token),
		"ExpiresAt":       verificationToken.ExpiresAt.Format("January 2, 2006 at 3:04 PM"),
	}

	return s.sendTemplatedEmail(EmailTypeVerification, email, username, data)
}

// SendPasswordResetEmail sends password reset email
func (s *Service) SendPasswordResetEmail(userID uuid.UUID, email, username string) error {
	// Generate reset token
	token, err := s.generateToken()
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	// Save token to database
	resetToken := &PasswordResetToken{
		UserID:    userID,
		Email:     email,
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if err := s.db.Create(resetToken).Error; err != nil {
		return fmt.Errorf("failed to save reset token: %w", err)
	}

	// Prepare email data
	data := EmailData{
		"UserName":  username,
		"ResetURL":  fmt.Sprintf("%s/reset-password?token=%s", s.cfg.GetString("server.url"), token),
		"ExpiresAt": resetToken.ExpiresAt.Format("January 2, 2006 at 3:04 PM"),
	}

	return s.sendTemplatedEmail(EmailTypePasswordReset, email, username, data)
}

// SendWelcomeEmail sends welcome email to new users
func (s *Service) SendWelcomeEmail(email, username string) error {
	data := EmailData{
		"UserName":   username,
		"LoginURL":   fmt.Sprintf("%s/login", s.cfg.GetString("server.url")),
		"SupportURL": fmt.Sprintf("%s/support", s.cfg.GetString("server.url")),
	}

	return s.sendTemplatedEmail(EmailTypeWelcome, email, username, data)
}

// SendGistStarredNotification sends notification when gist is starred
func (s *Service) SendGistStarredNotification(recipientID uuid.UUID, recipientEmail, recipientName, actorName, gistTitle string, gistID uuid.UUID) error {
	// Check if user wants this notification
	if !s.userWantsNotification(recipientID, "notify_gist_starred") {
		return nil
	}

	data := EmailData{
		"RecipientName": recipientName,
		"ActorName":     actorName,
		"GistTitle":     gistTitle,
		"GistURL":       fmt.Sprintf("%s/gists/%s", s.cfg.GetString("server.url"), gistID.String()),
	}

	return s.sendTemplatedEmail(EmailTypeGistStarred, recipientEmail, recipientName, data)
}

// SendUserFollowedNotification sends notification when user is followed
func (s *Service) SendUserFollowedNotification(recipientID uuid.UUID, recipientEmail, recipientName, followerName string, followerID uuid.UUID) error {
	// Check if user wants this notification
	if !s.userWantsNotification(recipientID, "notify_followed") {
		return nil
	}

	data := EmailData{
		"RecipientName": recipientName,
		"FollowerName":  followerName,
		"FollowerURL":   fmt.Sprintf("%s/users/%s", s.cfg.GetString("server.url"), followerID.String()),
		"ProfileURL":    fmt.Sprintf("%s/users/%s", s.cfg.GetString("server.url"), recipientID.String()),
	}

	return s.sendTemplatedEmail(EmailTypeUserFollowed, recipientEmail, recipientName, data)
}

// ProcessEmailQueue processes pending emails in the queue
func (s *Service) ProcessEmailQueue(ctx context.Context) error {
	// Get pending emails ordered by priority and creation time
	var emails []EmailQueue
	if err := s.db.Where("status = ? AND (scheduled_at IS NULL OR scheduled_at <= ?)",
		EmailStatusPending, time.Now()).
		Order("priority ASC, created_at ASC").
		Limit(10). // Process 10 emails at a time
		Find(&emails).Error; err != nil {
		return fmt.Errorf("failed to fetch pending emails: %w", err)
	}

	for _, email := range emails {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := s.processEmail(&email); err != nil {
				log.Printf("Failed to process email %s: %v", email.ID, err)
			}
		}
	}

	return nil
}

// processEmail processes a single email
func (s *Service) processEmail(email *EmailQueue) error {
	// Mark as sending
	email.Status = EmailStatusSending
	email.Attempts++
	if err := s.db.Save(email).Error; err != nil {
		return fmt.Errorf("failed to update email status: %w", err)
	}

	// Send the email
	if err := s.mailer.SendEmail(email); err != nil {
		// Mark as failed
		email.Status = EmailStatusFailed
		email.Error = err.Error()

		// Schedule retry if attempts remaining
		if email.Attempts < email.MaxAttempts {
			email.Status = EmailStatusPending
			// Exponential backoff: 5min, 15min, 45min
			retryDelay := time.Duration(5*email.Attempts*email.Attempts) * time.Minute
			scheduledAt := time.Now().Add(retryDelay)
			email.ScheduledAt = &scheduledAt
		}

		return s.db.Save(email).Error
	}

	// Mark as sent
	email.Status = EmailStatusSent
	email.Error = ""
	sentAt := time.Now()
	email.SentAt = &sentAt

	return s.db.Save(email).Error
}

// sendTemplatedEmail sends an email using templates
func (s *Service) sendTemplatedEmail(emailType EmailType, toEmail, toName string, data EmailData) error {
	if !s.cfg.GetBool("email.enabled") {
		log.Printf("Email disabled, skipping %s email to %s", emailType, toEmail)
		return nil
	}

	// Render templates
	htmlBody, err := s.renderer.RenderHTML(emailType, data)
	if err != nil {
		return fmt.Errorf("failed to render HTML template: %w", err)
	}

	textBody, err := s.renderer.RenderText(emailType, data)
	if err != nil {
		return fmt.Errorf("failed to render text template: %w", err)
	}

	// Get subject
	subjects := GetDefaultSubjects()
	subject, exists := subjects[emailType]
	if !exists {
		subject = "CasGists Notification"
	}

	// Render subject with data
	renderedSubject, err := RenderSubject(subject, data)
	if err != nil {
		log.Printf("Failed to render subject template, using default: %v", err)
		renderedSubject = subject
	}

	// Create email queue entry
	email := &EmailQueue{
		Type:      emailType,
		ToEmail:   toEmail,
		ToName:    toName,
		FromEmail: s.cfg.GetString("email.from_email"),
		FromName:  s.cfg.GetString("email.from_name"),
		Subject:   renderedSubject,
		BodyHTML:  htmlBody,
		BodyText:  textBody,
		Priority:  s.getEmailPriority(emailType),
	}

	return s.QueueEmail(email)
}

// getEmailPriority returns priority for email type
func (s *Service) getEmailPriority(emailType EmailType) int {
	switch emailType {
	case EmailTypeVerification, EmailTypePasswordReset:
		return 1 // Highest priority
	case EmailTypeWelcome:
		return 2
	case EmailTypeSystemAlert:
		return 3
	case EmailTypeGistStarred, EmailTypeGistForked, EmailTypeUserFollowed:
		return 5 // Normal priority
	case EmailTypeWeeklyDigest:
		return 8 // Lower priority
	default:
		return 5
	}
}

// userWantsNotification checks if user wants specific notification
func (s *Service) userWantsNotification(userID uuid.UUID, notificationType string) bool {
	var preference EmailPreference
	if err := s.db.Where("user_id = ?", userID).First(&preference).Error; err != nil {
		// If no preference found, assume user wants notifications
		return true
	}

	switch notificationType {
	case "notify_gist_starred":
		return preference.NotifyGistStarred
	case "notify_gist_forked":
		return preference.NotifyGistForked
	case "notify_gist_commented":
		return preference.NotifyGistCommented
	case "notify_followed":
		return preference.NotifyFollowed
	case "notify_weekly_digest":
		return preference.NotifyWeeklyDigest
	case "notify_system_alerts":
		return preference.NotifySystemAlerts
	default:
		return true
	}
}

// CreateEmailPreference creates email preferences for a user
func (s *Service) CreateEmailPreference(userID uuid.UUID) error {
	preference := &EmailPreference{
		UserID:              userID,
		EmailVerified:       false,
		NotifyGistStarred:   true,
		NotifyGistForked:    true,
		NotifyGistCommented: true,
		NotifyFollowed:      true,
		NotifyWeeklyDigest:  false,
		NotifySystemAlerts:  true,
	}

	return s.db.Create(preference).Error
}

// UpdateEmailPreference updates user email preferences
func (s *Service) UpdateEmailPreference(userID uuid.UUID, updates map[string]bool) error {
	var preference EmailPreference
	if err := s.db.Where("user_id = ?", userID).First(&preference).Error; err != nil {
		return err
	}

	if value, ok := updates["notify_gist_starred"]; ok {
		preference.NotifyGistStarred = value
	}
	if value, ok := updates["notify_gist_forked"]; ok {
		preference.NotifyGistForked = value
	}
	if value, ok := updates["notify_gist_commented"]; ok {
		preference.NotifyGistCommented = value
	}
	if value, ok := updates["notify_followed"]; ok {
		preference.NotifyFollowed = value
	}
	if value, ok := updates["notify_weekly_digest"]; ok {
		preference.NotifyWeeklyDigest = value
	}
	if value, ok := updates["notify_system_alerts"]; ok {
		preference.NotifySystemAlerts = value
	}

	return s.db.Save(&preference).Error
}

// GetEmailPreference gets user email preferences
func (s *Service) GetEmailPreference(userID uuid.UUID) (*EmailPreference, error) {
	var preference EmailPreference
	if err := s.db.Where("user_id = ?", userID).First(&preference).Error; err != nil {
		return nil, err
	}
	return &preference, nil
}

// VerifyEmail verifies an email address using token
func (s *Service) VerifyEmail(token string) error {
	var verificationToken EmailVerificationToken
	if err := s.db.Where("token = ? AND used = ? AND expires_at > ?",
		token, false, time.Now()).First(&verificationToken).Error; err != nil {
		return fmt.Errorf("invalid or expired verification token")
	}

	// Mark token as used
	verificationToken.Used = true
	usedAt := time.Now()
	verificationToken.UsedAt = &usedAt

	if err := s.db.Save(&verificationToken).Error; err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	// Update email preference to mark email as verified
	return s.db.Model(&EmailPreference{}).
		Where("user_id = ?", verificationToken.UserID).
		Update("email_verified", true).Error
}

// VerifyPasswordResetToken verifies a password reset token
func (s *Service) VerifyPasswordResetToken(token string) (*PasswordResetToken, error) {
	var resetToken PasswordResetToken
	if err := s.db.Where("token = ? AND used = ? AND expires_at > ?",
		token, false, time.Now()).First(&resetToken).Error; err != nil {
		return nil, fmt.Errorf("invalid or expired reset token")
	}

	return &resetToken, nil
}

// UsePasswordResetToken marks a password reset token as used
func (s *Service) UsePasswordResetToken(token string) error {
	var resetToken PasswordResetToken
	if err := s.db.Where("token = ?", token).First(&resetToken).Error; err != nil {
		return fmt.Errorf("token not found")
	}

	resetToken.Used = true
	usedAt := time.Now()
	resetToken.UsedAt = &usedAt

	return s.db.Save(&resetToken).Error
}

// generateToken generates a secure random token
func (s *Service) generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// TestEmailConfiguration tests email configuration
func (s *Service) TestEmailConfiguration() error {
	return s.mailer.TestConnection()
}

// SendTestEmail sends a test email
func (s *Service) SendTestEmail(toEmail, toName string) error {
	return s.mailer.SendTestEmail(toEmail, toName)
}

// SendGistCommentNotification sends notification when someone comments on a gist
func (s *Service) SendGistCommentNotification(recipientID uuid.UUID, recipientEmail, recipientName, commenterName, gistTitle, commentPreview string, gistID, commentID uuid.UUID) error {
	// Check if user wants this notification
	if !s.userWantsNotification(recipientID, "notify_gist_commented") {
		return nil
	}

	// Truncate comment preview to 200 characters
	if len(commentPreview) > 200 {
		commentPreview = commentPreview[:197] + "..."
	}

	data := EmailData{
		"RecipientName":  recipientName,
		"CommenterName":  commenterName,
		"GistTitle":      gistTitle,
		"CommentPreview": commentPreview,
		"GistURL":        fmt.Sprintf("%s/gists/%s", s.cfg.GetString("server.url"), gistID.String()),
		"CommentID":      commentID.String(),
		"SettingsURL":    fmt.Sprintf("%s/settings/notifications", s.cfg.GetString("server.url")),
	}

	return s.sendTemplatedEmail(EmailTypeGistCommented, recipientEmail, recipientName, data)
}

// SendBackupCompleteNotification sends notification when backup is completed
func (s *Service) SendBackupCompleteNotification(userID uuid.UUID, email, username string, backupStats BackupStats) error {
	// Check if user wants system alerts
	if !s.userWantsNotification(userID, "notify_system_alerts") {
		return nil
	}

	data := EmailData{
		"UserName":        username,
		"BackupDate":      backupStats.Date.Format("January 2, 2006 at 3:04 PM"),
		"BackupSize":      formatSize(backupStats.Size),
		"GistCount":       backupStats.GistCount,
		"UserCount":       backupStats.UserCount,
		"StorageLocation": backupStats.StorageLocation,
		"BackupType":      backupStats.Type,
		"NextBackupDate":  backupStats.NextBackupDate.Format("January 2, 2006 at 3:04 PM"),
		"AdminURL":        fmt.Sprintf("%s/admin/backup", s.cfg.GetString("server.url")),
	}

	// Add download URL if available
	if backupStats.DownloadURL != "" {
		data["DownloadURL"] = backupStats.DownloadURL
	}

	return s.sendTemplatedEmail(EmailTypeBackupComplete, email, username, data)
}

// SendMigrationCompleteNotification sends notification when migration is completed
func (s *Service) SendMigrationCompleteNotification(userID uuid.UUID, email, username string, migrationStats MigrationStats) error {
	data := EmailData{
		"UserName":       username,
		"MigrationDate":  migrationStats.Date.Format("January 2, 2006 at 3:04 PM"),
		"SourcePlatform": migrationStats.SourcePlatform,
		"Duration":       formatDuration(migrationStats.Duration),
		"GistCount":      migrationStats.GistCount,
		"StarCount":      migrationStats.StarCount,
		"FollowerCount":  migrationStats.FollowerCount,
		"SkippedItems":   migrationStats.SkippedItems,
		"DashboardURL":   fmt.Sprintf("%s/dashboard", s.cfg.GetString("server.url")),
		"SupportURL":     fmt.Sprintf("%s/support", s.cfg.GetString("server.url")),
	}

	// Add migration report URL if there are skipped items
	if migrationStats.SkippedItems > 0 {
		data["MigrationReportURL"] = fmt.Sprintf("%s/migrations/%s/report", s.cfg.GetString("server.url"), migrationStats.ID.String())
	}

	return s.sendTemplatedEmail(EmailTypeMigrationComplete, email, username, data)
}

// formatSize formats bytes to human readable format
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatDuration formats duration to human readable format
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	} else {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		if minutes > 0 {
			return fmt.Sprintf("%d hours %d minutes", hours, minutes)
		}
		return fmt.Sprintf("%d hours", hours)
	}
}
