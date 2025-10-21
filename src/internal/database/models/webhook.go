package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Webhook represents a webhook configuration
type Webhook struct {
	ID          uuid.UUID    `json:"id" gorm:"type:uuid;primary_key"`
	UserID      *uuid.UUID   `json:"user_id" gorm:"type:uuid;index"` // nil for system-wide webhooks
	URL         string       `json:"url" gorm:"type:text;not null"`
	Secret      string       `json:"-" gorm:"type:text"` // HMAC secret for verification
	Events      string       `json:"events" gorm:"type:text"` // JSON array of subscribed events
	IsActive    bool         `json:"is_active" gorm:"default:true"`
	ContentType string       `json:"content_type" gorm:"default:'application/json'"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	DeletedAt   *time.Time   `json:"deleted_at,omitempty" gorm:"index"`

	// Relations
	User *User `gorm:"constraint:OnDelete:CASCADE"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key"`
	WebhookID   uuid.UUID  `json:"webhook_id" gorm:"type:uuid;not null;index"`
	Event       string     `json:"event" gorm:"type:text;not null"`
	Payload     string     `json:"payload" gorm:"type:text"` // JSON payload
	URL         string     `json:"url" gorm:"type:text;not null"`
	StatusCode  int        `json:"status_code" gorm:"default:0"`
	Success     bool       `json:"success" gorm:"default:false"`
	Error       string     `json:"error,omitempty" gorm:"type:text"`
	Duration    int64      `json:"duration" gorm:"default:0"` // milliseconds
	Attempts    int        `json:"attempts" gorm:"default:1"`
	NextRetry   *time.Time `json:"next_retry,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Relations
	Webhook *Webhook `gorm:"constraint:OnDelete:CASCADE"`
}

// Legacy WebhookSubscription for backward compatibility
type WebhookSubscription struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key"`
	UserID         *uuid.UUID     `gorm:"type:uuid"`
	OrganizationID *uuid.UUID     `gorm:"type:uuid"`
	URL            string         `gorm:"size:500;not null"`
	Secret         string         `gorm:"size:64"`
	Events         []string       `gorm:"type:text"`
	Active         bool           `gorm:"default:true"`
	CreatedAt      time.Time
	UpdatedAt      time.Time

	// Relations
	User         *User         `gorm:"constraint:OnDelete:CASCADE"`
	Organization *Organization `gorm:"constraint:OnDelete:CASCADE"`
}

// BeforeCreate hooks
func (w *Webhook) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}

func (wd *WebhookDelivery) BeforeCreate(tx *gorm.DB) error {
	if wd.ID == uuid.Nil {
		wd.ID = uuid.New()
	}
	return nil
}

func (ws *WebhookSubscription) BeforeCreate(tx *gorm.DB) error {
	if ws.ID == uuid.Nil {
		ws.ID = uuid.New()
	}
	// Validate ownership
	if (ws.UserID == nil && ws.OrganizationID == nil) || 
	   (ws.UserID != nil && ws.OrganizationID != nil) {
		return gorm.ErrInvalidData
	}
	return nil
}