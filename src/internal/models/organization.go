package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Organization represents an organization
type Organization struct {
	ID            uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	Name          string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"name"`
	DisplayName   string         `gorm:"type:varchar(255)" json:"display_name"`
	Description   string         `gorm:"type:text" json:"description"`
	Email         string         `gorm:"type:varchar(255)" json:"email"`
	Website       string         `gorm:"type:varchar(255)" json:"website"`
	Location      string         `gorm:"type:varchar(255)" json:"location"`
	AvatarURL     string         `gorm:"type:varchar(255)" json:"avatar_url"`
	IsPublic      bool           `gorm:"default:true" json:"is_public"`
	IsVerified    bool           `gorm:"default:false" json:"is_verified"`
	GistCount     int64          `gorm:"default:0" json:"gist_count"`
	MemberCount   int64          `gorm:"default:0" json:"member_count"`
	Settings      string         `gorm:"type:text" json:"-"` // JSON settings
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Members []OrganizationUser `gorm:"foreignKey:OrganizationID" json:"-"`
	Teams   []Team             `gorm:"foreignKey:OrganizationID" json:"-"`
	Gists   []Gist             `gorm:"foreignKey:OrganizationID" json:"-"`
}

// BeforeCreate hook to set UUID
func (o *Organization) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}

// OrganizationUser represents a user's membership in an organization
type OrganizationUser struct {
	ID             uuid.UUID `gorm:"type:char(36);primary_key"`
	UserID         uuid.UUID `gorm:"type:char(36);not null;index"`
	OrganizationID uuid.UUID `gorm:"type:char(36);not null;index"`
	Role           string    `gorm:"type:varchar(50);not null;default:'member'"` // owner, admin, member
	CreatedAt      time.Time
	UpdatedAt      time.Time

	// Relationships
	User         User         `gorm:"foreignKey:UserID"`
	Organization Organization `gorm:"foreignKey:OrganizationID"`
}

// BeforeCreate hook to set UUID
func (ou *OrganizationUser) BeforeCreate(tx *gorm.DB) error {
	if ou.ID == uuid.Nil {
		ou.ID = uuid.New()
	}
	return nil
}

// Team represents a team within an organization
type Team struct {
	ID             uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:char(36);not null;index" json:"organization_id"`
	Name           string         `gorm:"type:varchar(255);not null" json:"name"`
	Slug           string         `gorm:"type:varchar(255);not null;index" json:"slug"`
	Description    string         `gorm:"type:text" json:"description"`
	Permission     string         `gorm:"type:varchar(50);default:'read'" json:"permission"` // read, write, admin
	Permissions    string         `gorm:"type:text" json:"-"` // JSON permissions (for future use)
	MemberCount    int64          `gorm:"default:0" json:"member_count"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Organization Organization `gorm:"foreignKey:OrganizationID" json:"-"`
	Members      []TeamMember `gorm:"foreignKey:TeamID" json:"-"`
}

// BeforeCreate hook to set UUID
func (t *Team) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// TeamMember represents a user's membership in a team
type TeamMember struct {
	ID        uuid.UUID `gorm:"type:char(36);primary_key"`
	TeamID    uuid.UUID `gorm:"type:char(36);not null;index"`
	UserID    uuid.UUID `gorm:"type:char(36);not null;index"`
	Role      string    `gorm:"type:varchar(50);not null;default:'member'"` // maintainer, member
	CreatedAt time.Time
	UpdatedAt time.Time

	// Relationships
	Team Team `gorm:"foreignKey:TeamID"`
	User User `gorm:"foreignKey:UserID"`
}

// BeforeCreate hook to set UUID
func (tm *TeamMember) BeforeCreate(tx *gorm.DB) error {
	if tm.ID == uuid.Nil {
		tm.ID = uuid.New()
	}
	return nil
}