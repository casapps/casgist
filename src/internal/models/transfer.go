package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Transfer represents a gist ownership transfer request
type Transfer struct {
	ID               uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	GistID           uuid.UUID      `gorm:"type:char(36);not null;index" json:"gist_id"`
	FromUserID       uuid.UUID      `gorm:"type:char(36);not null;index" json:"from_user_id"`
	ToUserID         *uuid.UUID     `gorm:"type:char(36);index" json:"to_user_id,omitempty"`
	ToOrganizationID *uuid.UUID     `gorm:"type:char(36);index" json:"to_organization_id,omitempty"`
	Token            string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"-"`
	Status           string         `gorm:"type:varchar(50);not null;default:'pending'" json:"status"` // pending, accepted, rejected, cancelled, expired
	Message          string         `gorm:"type:text" json:"message"`
	ExpiresAt        time.Time      `gorm:"not null" json:"expires_at"`
	ProcessedAt      *time.Time     `json:"processed_at,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Gist           Gist          `gorm:"foreignKey:GistID" json:"gist,omitempty"`
	FromUser       User          `gorm:"foreignKey:FromUserID" json:"from_user,omitempty"`
	ToUser         *User         `gorm:"foreignKey:ToUserID" json:"to_user,omitempty"`
	ToOrganization *Organization `gorm:"foreignKey:ToOrganizationID" json:"to_organization,omitempty"`
}

// BeforeCreate hook to set UUID
func (t *Transfer) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}