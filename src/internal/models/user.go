package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID                 uuid.UUID  `gorm:"type:char(36);primary_key" json:"id"`
	Username           string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"username"`
	Email              string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash       string     `gorm:"type:varchar(255);not null" json:"-"`
	DisplayName        string     `gorm:"type:varchar(255)" json:"display_name"`
	Bio                string     `gorm:"type:text" json:"bio"`
	AvatarURL          string     `gorm:"type:varchar(255)" json:"avatar_url"`
	IsAdmin            bool       `gorm:"default:false" json:"is_admin"`
	IsActive           bool       `gorm:"default:true" json:"is_active"`
	IsSuspended        bool       `gorm:"default:false" json:"is_suspended"`
	EmailVerified      bool       `gorm:"default:false" json:"email_verified"`
	TwoFactorEnabled   bool       `gorm:"default:false" json:"two_factor_enabled"`
	TwoFactorSecret    string     `gorm:"type:varchar(255)" json:"-"`
	WebAuthnEnabled    bool       `gorm:"default:false" json:"webauthn_enabled"`
	LastLoginAt        *time.Time `json:"last_login_at"`
	LastLoginIP        string     `gorm:"type:varchar(45)" json:"-"`
	PasswordChangedAt  *time.Time `json:"-"`
	RecoveryEmail      string     `gorm:"type:varchar(255)" json:"-"`
	Locale             string     `gorm:"type:varchar(10);default:'en'" json:"locale"`
	Timezone           string     `gorm:"type:varchar(50);default:'UTC'" json:"timezone"`
	MaxGists           int        `gorm:"default:1000" json:"max_gists"`
	MaxFileSize        int64      `gorm:"default:10485760" json:"max_file_size"` // 10MB default
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Gists              []Gist              `gorm:"foreignKey:UserID" json:"-"`
	Sessions           []Session           `gorm:"foreignKey:UserID" json:"-"`
	APITokens          []APIToken          `gorm:"foreignKey:UserID" json:"-"`
	OrganizationUsers  []OrganizationUser  `gorm:"foreignKey:UserID" json:"-"`
	TeamMembers        []TeamMember        `gorm:"foreignKey:UserID" json:"-"`
	StarredGists       []Star              `gorm:"foreignKey:UserID" json:"-"`
	Comments           []Comment           `gorm:"foreignKey:UserID" json:"-"`
	AuditLogs          []AuditLog          `gorm:"foreignKey:UserID" json:"-"`
	WebAuthnCredentials []WebAuthnCredential `gorm:"foreignKey:UserID" json:"-"`
	Followers          []Follow            `gorm:"foreignKey:FollowerID" json:"-"`
	Following          []Follow            `gorm:"foreignKey:FollowingID" json:"-"`
}

// BeforeCreate hook to set UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// Session represents a user session
type Session struct {
	ID           uuid.UUID `gorm:"type:char(36);primary_key"`
	UserID       uuid.UUID `gorm:"type:char(36);not null;index"`
	Token        string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	RefreshToken string    `gorm:"type:varchar(255);uniqueIndex"`
	UserAgent    string    `gorm:"type:text"`
	IPAddress    string    `gorm:"type:varchar(45)"`
	ExpiresAt    time.Time `gorm:"not null"`
	LastAccessAt time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time

	// Relationships
	User User `gorm:"foreignKey:UserID"`
}

// BeforeCreate hook to set UUID
func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// APIToken represents an API token for programmatic access
type APIToken struct {
	ID          uuid.UUID `gorm:"type:char(36);primary_key"`
	UserID      uuid.UUID `gorm:"type:char(36);not null;index"`
	Name        string    `gorm:"type:varchar(255);not null"`
	Token       string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	Scopes      string    `gorm:"type:text"` // JSON array of scopes
	LastUsedAt  *time.Time
	LastUsedIP  string    `gorm:"type:varchar(45)"`
	ExpiresAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// Relationships
	User User `gorm:"foreignKey:UserID"`
}

// BeforeCreate hook to set UUID
func (t *APIToken) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// WebAuthnCredential represents a WebAuthn credential
type WebAuthnCredential struct {
	ID               uuid.UUID `gorm:"type:char(36);primary_key"`
	UserID           uuid.UUID `gorm:"type:char(36);not null;index"`
	CredentialID     []byte    `gorm:"type:blob;not null"`
	PublicKey        []byte    `gorm:"type:blob;not null"`
	AttestationType  string    `gorm:"type:varchar(50)"`
	Transport        string    `gorm:"type:varchar(255)"` // JSON array
	Flags            uint32
	Authenticator    []byte    `gorm:"type:blob"`
	Counter          uint32
	CreatedAt        time.Time
	UpdatedAt        time.Time
	LastUsedAt       *time.Time

	// Relationships
	User User `gorm:"foreignKey:UserID"`
}

// BeforeCreate hook to set UUID
func (w *WebAuthnCredential) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}

// Follow represents a follower relationship
type Follow struct {
	ID          uuid.UUID `gorm:"type:char(36);primary_key"`
	FollowerID  uuid.UUID `gorm:"type:char(36);not null;index"`
	FollowingID uuid.UUID `gorm:"type:char(36);not null;index"`
	CreatedAt   time.Time

	// Relationships
	Follower  User `gorm:"foreignKey:FollowerID"`
	Following User `gorm:"foreignKey:FollowingID"`
}

// BeforeCreate hook to set UUID
func (f *Follow) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

// UserPreference represents a user preference key-value pair
type UserPreference struct {
	ID        uuid.UUID `gorm:"type:char(36);primary_key" json:"id"`
	UserID    uuid.UUID `gorm:"type:char(36);not null;index" json:"user_id"`
	Key       string    `gorm:"type:varchar(100);not null" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// BeforeCreate hook for UserPreference
func (up *UserPreference) BeforeCreate(tx *gorm.DB) error {
	if up.ID == uuid.Nil {
		up.ID = uuid.New()
	}
	return nil
}