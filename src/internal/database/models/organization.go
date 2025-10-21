package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Organization represents a group of users
type Organization struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key"`
	Name        string         `gorm:"uniqueIndex;size:39;not null"`
	DisplayName string         `gorm:"size:100"`
	Description string         `gorm:"size:500"`
	Website     string         `gorm:"size:255"`
	Location    string         `gorm:"size:100"`
	AvatarURL   string         `gorm:"size:255"`
	IsPublic    bool           `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Relations
	Members               []OrganizationMember      `gorm:"constraint:OnDelete:CASCADE"`
	Invitations           []OrganizationInvitation  `gorm:"constraint:OnDelete:CASCADE"`
	Gists                 []Gist                    `gorm:"constraint:OnDelete:CASCADE"`
	CustomDomains         []CustomDomain            `gorm:"constraint:OnDelete:CASCADE"`
	// ImportJobs relation removed - ImportJobs are owned by users, not organizations
	WebhookSubscriptions  []WebhookSubscription     `gorm:"constraint:OnDelete:CASCADE"`
}

// OrganizationMember represents a user's membership in an organization
type OrganizationMember struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null"`
	UserID         uuid.UUID `gorm:"type:uuid;not null"`
	Role           string    `gorm:"size:20;not null;default:'member'"`
	JoinedAt       time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time

	// Relations
	Organization Organization `gorm:"constraint:OnDelete:CASCADE"`
	User         User         `gorm:"constraint:OnDelete:CASCADE"`
}

// OrganizationInvitation represents an invitation to join an organization
type OrganizationInvitation struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null"`
	Email          string     `gorm:"size:255;not null"`
	Role           string     `gorm:"size:20;not null;default:'member'"`
	InvitedByID    uuid.UUID  `gorm:"type:uuid;not null"`
	Token          string     `gorm:"uniqueIndex;size:64;not null"`
	ExpiresAt      time.Time  `gorm:"not null"`
	AcceptedAt     *time.Time
	CreatedAt      time.Time

	// Relations
	Organization Organization `gorm:"constraint:OnDelete:CASCADE"`
	InvitedBy    User         `gorm:"foreignKey:InvitedByID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hooks
func (o *Organization) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}

func (m *OrganizationMember) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	m.JoinedAt = time.Now()
	
	// Check for duplicate membership
	tx.Where("organization_id = ? AND user_id = ?", m.OrganizationID, m.UserID).First(&OrganizationMember{})
	if tx.RowsAffected > 0 {
		return gorm.ErrDuplicatedKey
	}
	return nil
}

func (i *OrganizationInvitation) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	// Default expiry to 7 days
	if i.ExpiresAt.IsZero() {
		i.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	}
	return nil
}