package webhooks

import (
	"time"

	"github.com/google/uuid"
)

// WebhookEvent represents the type of webhook event
type WebhookEvent string

const (
	EventGistCreated  WebhookEvent = "gist.created"
	EventGistUpdated  WebhookEvent = "gist.updated"
	EventGistDeleted  WebhookEvent = "gist.deleted"
	EventGistStarred  WebhookEvent = "gist.starred"
	EventGistForked   WebhookEvent = "gist.forked"
	EventUserCreated  WebhookEvent = "user.created"
	EventUserUpdated  WebhookEvent = "user.updated"
	EventUserFollowed WebhookEvent = "user.followed"
)

// Event represents a webhook event payload
type Event struct {
	ID        uuid.UUID      `json:"id"`
	Type      WebhookEvent   `json:"type"`
	Action    string         `json:"action"`
	Timestamp time.Time      `json:"timestamp"`
	Data      map[string]any `json:"data"`
	User      *EventUser     `json:"user,omitempty"`
	Gist      *EventGist     `json:"gist,omitempty"`
}

// EventUser represents user data in webhook events
type EventUser struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email,omitempty"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
}

// EventGist represents gist data in webhook events
type EventGist struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Visibility  string    `json:"visibility"`
	StarCount   int       `json:"star_count"`
	ForkCount   int       `json:"fork_count"`
	FileCount   int       `json:"file_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Webhook represents a webhook configuration
type Webhook struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      *uuid.UUID `json:"user_id" gorm:"type:uuid;index"` // nil for system-wide webhooks
	URL         string     `json:"url" gorm:"type:text;not null"`
	Secret      string     `json:"-" gorm:"type:text"`      // HMAC secret for verification
	Events      string     `json:"events" gorm:"type:text"` // JSON array of subscribed events
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	ContentType string     `json:"content_type" gorm:"default:'application/json'"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	WebhookID  uuid.UUID  `json:"webhook_id" gorm:"type:uuid;not null;index"`
	Event      string     `json:"event" gorm:"type:text;not null"`
	Payload    string     `json:"payload" gorm:"type:text"` // JSON payload
	URL        string     `json:"url" gorm:"type:text;not null"`
	StatusCode int        `json:"status_code" gorm:"default:0"`
	Success    bool       `json:"success" gorm:"default:false"`
	Error      string     `json:"error,omitempty" gorm:"type:text"`
	Duration   int64      `json:"duration" gorm:"default:0"` // milliseconds
	Attempts   int        `json:"attempts" gorm:"default:1"`
	NextRetry  *time.Time `json:"next_retry,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// WebhookPayload represents the structure of webhook payloads
type WebhookPayload struct {
	Event     WebhookEvent `json:"event"`
	Action    string       `json:"action,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
	Data      interface{}  `json:"data"`
	Sender    *SenderInfo  `json:"sender,omitempty"`
}

// SenderInfo represents information about the user who triggered the event
type SenderInfo struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email,omitempty"`
}

// GistEventData represents gist-related event data
type GistEventData struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Visibility  string    `json:"visibility"`
	URL         string    `json:"url"`
	FilesCount  int       `json:"files_count"`
	Tags        []string  `json:"tags,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserEventData represents user-related event data
type UserEventData struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name,omitempty"`
	Bio         string    `json:"bio,omitempty"`
	Location    string    `json:"location,omitempty"`
	Website     string    `json:"website,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FollowEventData represents follow-related event data
type FollowEventData struct {
	Follower  UserEventData `json:"follower"`
	Following UserEventData `json:"following"`
}

// StarEventData represents star-related event data
type StarEventData struct {
	Gist    GistEventData `json:"gist"`
	Starrer UserEventData `json:"starrer"`
}

// ForkEventData represents fork-related event data
type ForkEventData struct {
	OriginalGist GistEventData `json:"original_gist"`
	ForkedGist   GistEventData `json:"forked_gist"`
	Forker       UserEventData `json:"forker"`
}
