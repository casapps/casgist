package email

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EmailType represents the type of email being sent
type EmailType string

const (
	EmailTypeVerification      EmailType = "verification"
	EmailTypePasswordReset     EmailType = "password_reset"
	EmailTypeWelcome           EmailType = "welcome"
	EmailTypeGistCreated       EmailType = "gist_created"
	EmailTypeGistStarred       EmailType = "gist_starred"
	EmailTypeGistForked        EmailType = "gist_forked"
	EmailTypeGistCommented     EmailType = "gist_commented"
	EmailTypeUserFollowed      EmailType = "user_followed"
	EmailTypeWeeklyDigest      EmailType = "weekly_digest"
	EmailTypeSystemAlert       EmailType = "system_alert"
	EmailTypeInvitation        EmailType = "invitation"
	EmailTypeBackupComplete    EmailType = "backup_complete"
	EmailTypeMigrationComplete EmailType = "migration_complete"
)

// EmailTemplate represents an email template
type EmailTemplate struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	Type      EmailType `json:"type" gorm:"type:text;not null;unique"`
	Name      string    `json:"name" gorm:"type:text;not null"`
	Subject   string    `json:"subject" gorm:"type:text;not null"`
	BodyText  string    `json:"body_text" gorm:"type:text"`
	BodyHTML  string    `json:"body_html" gorm:"type:text"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	Variables string    `json:"variables" gorm:"type:text"` // JSON array of available variables
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EmailQueue represents an email in the sending queue
type EmailQueue struct {
	ID          uuid.UUID   `json:"id" gorm:"type:uuid;primary_key"`
	Type        EmailType   `json:"type" gorm:"type:text;not null"`
	ToEmail     string      `json:"to_email" gorm:"type:text;not null"`
	ToName      string      `json:"to_name" gorm:"type:text"`
	FromEmail   string      `json:"from_email" gorm:"type:text;not null"`
	FromName    string      `json:"from_name" gorm:"type:text"`
	Subject     string      `json:"subject" gorm:"type:text;not null"`
	BodyText    string      `json:"body_text" gorm:"type:text"`
	BodyHTML    string      `json:"body_html" gorm:"type:text"`
	Priority    int         `json:"priority" gorm:"default:5"` // 1=highest, 10=lowest
	Status      EmailStatus `json:"status" gorm:"type:text;default:'pending'"`
	Attempts    int         `json:"attempts" gorm:"default:0"`
	MaxAttempts int         `json:"max_attempts" gorm:"default:3"`
	Error       string      `json:"error,omitempty" gorm:"type:text"`
	ScheduledAt *time.Time  `json:"scheduled_at,omitempty"`
	SentAt      *time.Time  `json:"sent_at,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

func (q *EmailQueue) BeforeCreate(tx *gorm.DB) error {
	if q.ID == uuid.Nil {
		q.ID = uuid.New()
	}
	return nil
}

// EmailStatus represents the status of an email
type EmailStatus string

const (
	EmailStatusPending   EmailStatus = "pending"
	EmailStatusSending   EmailStatus = "sending"
	EmailStatusSent      EmailStatus = "sent"
	EmailStatusFailed    EmailStatus = "failed"
	EmailStatusCancelled EmailStatus = "cancelled"
)

// EmailPreference represents user email preferences
type EmailPreference struct {
	ID                  uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	UserID              uuid.UUID `json:"user_id" gorm:"type:uuid;not null;unique"`
	EmailVerified       bool      `json:"email_verified" gorm:"default:false"`
	NotifyGistStarred   bool      `json:"notify_gist_starred" gorm:"default:true"`
	NotifyGistForked    bool      `json:"notify_gist_forked" gorm:"default:true"`
	NotifyGistCommented bool      `json:"notify_gist_commented" gorm:"default:true"`
	NotifyFollowed      bool      `json:"notify_followed" gorm:"default:true"`
	NotifyWeeklyDigest  bool      `json:"notify_weekly_digest" gorm:"default:false"`
	NotifySystemAlerts  bool      `json:"notify_system_alerts" gorm:"default:true"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func (p *EmailPreference) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// EmailVerificationToken represents an email verification token
type EmailVerificationToken struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key"`
	UserID    uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	Email     string     `json:"email" gorm:"type:text;not null"`
	Token     string     `json:"token" gorm:"type:text;not null;unique"`
	Used      bool       `json:"used" gorm:"default:false"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

func (t *EmailVerificationToken) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key"`
	UserID    uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	Email     string     `json:"email" gorm:"type:text;not null"`
	Token     string     `json:"token" gorm:"type:text;not null;unique"`
	Used      bool       `json:"used" gorm:"default:false"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

func (t *PasswordResetToken) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// EmailData represents data for email template rendering
type EmailData map[string]interface{}

// Common email data structures
type WelcomeEmailData struct {
	UserName   string `json:"user_name"`
	LoginURL   string `json:"login_url"`
	SupportURL string `json:"support_url"`
}

type VerificationEmailData struct {
	UserName        string `json:"user_name"`
	VerificationURL string `json:"verification_url"`
	ExpiresAt       string `json:"expires_at"`
}

type PasswordResetEmailData struct {
	UserName  string `json:"user_name"`
	ResetURL  string `json:"reset_url"`
	ExpiresAt string `json:"expires_at"`
}

type GistNotificationEmailData struct {
	RecipientName string `json:"recipient_name"`
	ActorName     string `json:"actor_name"`
	GistTitle     string `json:"gist_title"`
	GistURL       string `json:"gist_url"`
	Action        string `json:"action"` // "starred", "forked", etc.
}

type FollowNotificationEmailData struct {
	RecipientName string `json:"recipient_name"`
	FollowerName  string `json:"follower_name"`
	FollowerURL   string `json:"follower_url"`
	ProfileURL    string `json:"profile_url"`
}

type WeeklyDigestEmailData struct {
	UserName         string       `json:"user_name"`
	Week             string       `json:"week"`
	GistsCreated     int          `json:"gists_created"`
	StarsReceived    int          `json:"stars_received"`
	NewFollowers     int          `json:"new_followers"`
	TrendingGists    []DigestGist `json:"trending_gists"`
	RecommendedUsers []DigestUser `json:"recommended_users"`
	UnsubscribeURL   string       `json:"unsubscribe_url"`
}

type DigestGist struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	StarCount   int    `json:"star_count"`
	Author      string `json:"author"`
}

type DigestUser struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Bio         string `json:"bio"`
	URL         string `json:"url"`
	GistCount   int    `json:"gist_count"`
}

// BackupStats represents backup completion statistics
type BackupStats struct {
	ID              uuid.UUID `json:"id"`
	Date            time.Time `json:"date"`
	Size            int64     `json:"size"`
	GistCount       int       `json:"gist_count"`
	UserCount       int       `json:"user_count"`
	StorageLocation string    `json:"storage_location"`
	Type            string    `json:"type"`
	DownloadURL     string    `json:"download_url,omitempty"`
	NextBackupDate  time.Time `json:"next_backup_date"`
}

// MigrationStats represents migration completion statistics
type MigrationStats struct {
	ID             uuid.UUID     `json:"id"`
	Date           time.Time     `json:"date"`
	SourcePlatform string        `json:"source_platform"`
	Duration       time.Duration `json:"duration"`
	GistCount      int           `json:"gist_count"`
	StarCount      int           `json:"star_count"`
	FollowerCount  int           `json:"follower_count"`
	SkippedItems   int           `json:"skipped_items"`
}
