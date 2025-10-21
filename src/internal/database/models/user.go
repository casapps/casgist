package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user account
type User struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key"`
	Username         string         `gorm:"uniqueIndex;size:39;not null"`
	Email            string         `gorm:"uniqueIndex;size:255;not null"`
	PasswordHash     string         `gorm:"size:255;not null"`
	DisplayName      string         `gorm:"size:100"`
	Bio              string         `gorm:"size:500"`
	Website          string         `gorm:"size:255"`
	Location         string         `gorm:"size:100"`
	AvatarURL        string         `gorm:"size:255"`
	IsAdmin          bool           `gorm:"default:false"`
	IsActive         bool           `gorm:"default:true"`
	EmailVerified    bool           `gorm:"default:false"`
	TwoFactorEnabled bool           `gorm:"default:false"`
	TwoFactorSecret  string         `gorm:"size:32"`
	IsSuspended      bool           `gorm:"default:false"`
	IsEmailVerified  bool           `gorm:"default:false"`
	LastLoginAt      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`
	
	// Computed fields for templates (not stored in DB)
	Status      string `gorm:"-"`
	GistCount   int    `gorm:"-"`
	StorageUsed string `gorm:"-"`

	// Relations
	Preferences        *UserPreference       `gorm:"constraint:OnDelete:CASCADE"`
	Gists              []Gist                `gorm:"constraint:OnDelete:CASCADE"`
	Sessions           []Session             `gorm:"constraint:OnDelete:CASCADE"`
	APITokens          []APIToken            `gorm:"constraint:OnDelete:CASCADE"`
	Organizations      []OrganizationMember  `gorm:"constraint:OnDelete:CASCADE"`
	StarredGists       []GistStar            `gorm:"constraint:OnDelete:CASCADE"`
	Comments           []GistComment         `gorm:"constraint:OnDelete:CASCADE"`
	Followers          []UserFollow          `gorm:"foreignKey:FollowingID;constraint:OnDelete:CASCADE"`
	Following          []UserFollow          `gorm:"foreignKey:FollowerID;constraint:OnDelete:CASCADE"`
	AuditLogs          []AuditLog            `gorm:"constraint:OnDelete:SET NULL"`
	SecurityEvents     []SecurityEvent       `gorm:"constraint:OnDelete:SET NULL"`
	ImportJobs         []ImportJob           `gorm:"foreignKey:CreatedBy;constraint:OnDelete:CASCADE"`
	WebhookSubscriptions []WebhookSubscription `gorm:"constraint:OnDelete:CASCADE"`
}

// UserPreference stores user preferences
type UserPreference struct {
	ID                    uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID                uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"`
	Theme                 string    `gorm:"size:20;default:'dracula'"`
	Language              string    `gorm:"size:10;default:'en'"`
	EmailNotifications    bool      `gorm:"default:true"`
	EmailDigest           bool      `gorm:"default:true"`
	PublicProfile         bool      `gorm:"default:true"`
	PublicEmail           bool      `gorm:"default:false"`
	DefaultGistVisibility string    `gorm:"size:20;default:'private'"`
	DefaultSortOrder      string    `gorm:"size:20;default:'recently_updated'"`
	EditorFontSize        int       `gorm:"default:14"`
	EditorTabSize         int       `gorm:"default:4"`
	EditorWordWrap        bool      `gorm:"default:true"`
	EditorTheme           string    `gorm:"size:20;default:'dracula'"`
	Timezone              string    `gorm:"size:50;default:'UTC'"`
	DateFormat            string    `gorm:"size:20;default:'YYYY-MM-DD'"`
	CreatedAt             time.Time
	UpdatedAt             time.Time

	// Relations
	User User `gorm:"constraint:OnDelete:CASCADE"`
}

// Session represents a user session
type Session struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID       uuid.UUID `gorm:"type:uuid;not null"`
	Token        string    `gorm:"uniqueIndex;size:128;not null"`
	RefreshToken string    `gorm:"uniqueIndex;size:128;not null"`
	IPAddress    string    `gorm:"size:45"`
	UserAgent    string    `gorm:"size:500"`
	ExpiresAt    time.Time `gorm:"not null"`
	LastUsedAt   time.Time
	CreatedAt    time.Time

	// Relations
	User User `gorm:"constraint:OnDelete:CASCADE"`
}

// APIToken represents an API access token
type APIToken struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null"`
	Name        string     `gorm:"size:100;not null"`
	TokenHash   string     `gorm:"uniqueIndex;size:64;not null"`
	TokenPrefix string     `gorm:"size:8;not null"`
	Permissions string     `gorm:"type:text"` // JSON array
	ExpiresAt   *time.Time
	LastUsedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// Relations
	User User `gorm:"constraint:OnDelete:CASCADE"`
}

// UserFollow represents a follow relationship between users
type UserFollow struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key"`
	FollowerID  uuid.UUID `gorm:"type:uuid;not null"`
	FollowingID uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt   time.Time

	// Relations
	Follower  User `gorm:"foreignKey:FollowerID;constraint:OnDelete:CASCADE"`
	Following User `gorm:"foreignKey:FollowingID;constraint:OnDelete:CASCADE"`
}

// UserBlock represents a block relationship between users
type UserBlock struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	BlockerID uuid.UUID `gorm:"type:uuid;not null"`
	BlockedID uuid.UUID `gorm:"type:uuid;not null"`
	Reason    string    `gorm:"size:500"`
	CreatedAt time.Time

	// Relations
	Blocker User `gorm:"foreignKey:BlockerID;constraint:OnDelete:CASCADE"`
	Blocked User `gorm:"foreignKey:BlockedID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hooks for UUID generation
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (p *UserPreference) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (t *APIToken) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (f *UserFollow) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	// Add unique constraint check
	tx.Where("follower_id = ? AND following_id = ?", f.FollowerID, f.FollowingID).First(&UserFollow{})
	if tx.RowsAffected > 0 {
		return gorm.ErrDuplicatedKey
	}
	return nil
}

func (b *UserBlock) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	// Add unique constraint check
	tx.Where("blocker_id = ? AND blocked_id = ?", b.BlockerID, b.BlockedID).First(&UserBlock{})
	if tx.RowsAffected > 0 {
		return gorm.ErrDuplicatedKey
	}
	return nil
}