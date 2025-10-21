package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Webhook represents a webhook configuration
type Webhook struct {
	ID               uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	UserID           *uuid.UUID     `gorm:"type:char(36);index" json:"user_id,omitempty"`
	OrganizationID   *uuid.UUID     `gorm:"type:char(36);index" json:"organization_id,omitempty"`
	URL              string         `gorm:"type:varchar(255);not null" json:"url"`
	Secret           string         `gorm:"type:varchar(255)" json:"-"`
	Events           string         `gorm:"type:text" json:"events"` // JSON array of events
	IsActive         bool           `gorm:"default:true" json:"is_active"`
	ContentType      string         `gorm:"type:varchar(50);default:'application/json'" json:"content_type"`
	InsecureSSL      bool           `gorm:"default:false" json:"insecure_ssl"`
	LastDeliveredAt  *time.Time     `json:"last_delivered_at,omitempty"`
	LastStatus       int            `json:"last_status,omitempty"`
	LastError        string         `gorm:"type:text" json:"last_error,omitempty"`
	DeliveryCount    int64          `gorm:"default:0" json:"delivery_count"`
	FailureCount     int64          `gorm:"default:0" json:"failure_count"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User         *User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Organization *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Deliveries   []WebhookDelivery `gorm:"foreignKey:WebhookID" json:"-"`
}

// BeforeCreate hook to set UUID
func (w *Webhook) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID             uuid.UUID `gorm:"type:char(36);primary_key" json:"id"`
	WebhookID      uuid.UUID `gorm:"type:char(36);not null;index" json:"webhook_id"`
	Event          string    `gorm:"type:varchar(100);not null" json:"event"`
	Payload        string    `gorm:"type:longtext" json:"payload"`
	RequestHeaders string    `gorm:"type:text" json:"request_headers"`
	ResponseStatus int       `json:"response_status"`
	ResponseHeaders string    `gorm:"type:text" json:"response_headers,omitempty"`
	ResponseBody   string    `gorm:"type:text" json:"response_body,omitempty"`
	Duration       int64     `json:"duration"` // milliseconds
	Success        bool      `json:"success"`
	Error          string    `gorm:"type:text" json:"error,omitempty"`
	CreatedAt      time.Time `json:"created_at"`

	// Relationships
	Webhook Webhook `gorm:"foreignKey:WebhookID" json:"-"`
}

// BeforeCreate hook to set UUID
func (d *WebhookDelivery) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}