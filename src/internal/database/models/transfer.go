package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TransferStatus represents the status of a transfer request
type TransferStatus string

const (
	TransferStatusPending   TransferStatus = "pending"
	TransferStatusAccepted  TransferStatus = "accepted"
	TransferStatusRejected  TransferStatus = "rejected"
	TransferStatusCancelled TransferStatus = "cancelled"
	TransferStatusExpired   TransferStatus = "expired"
	TransferStatusCompleted TransferStatus = "completed"
	TransferStatusFailed    TransferStatus = "failed"
)

// TransferReason represents the reason for a transfer
type TransferReason string

const (
	TransferReasonReorganization  TransferReason = "reorganization"
	TransferReasonOwnershipChange TransferReason = "ownership_change"
	TransferReasonTeamRestructure TransferReason = "team_restructure"
	TransferReasonProjectMigration TransferReason = "project_migration"
	TransferReasonOther           TransferReason = "other"
)

// TransferRequest represents a request to transfer gist ownership
type TransferRequest struct {
	ID                   uuid.UUID       `gorm:"type:uuid;primary_key"`
	GistID               uuid.UUID       `gorm:"type:uuid;not null"`
	FromUserID           *uuid.UUID      `gorm:"type:uuid"`
	FromOrganizationID   *uuid.UUID      `gorm:"type:uuid"`
	ToUserID             *uuid.UUID      `gorm:"type:uuid"`
	ToOrganizationID     *uuid.UUID      `gorm:"type:uuid"`
	RequestedByID        uuid.UUID       `gorm:"type:uuid;not null"`
	RequestedAt          time.Time
	ExpiresAt            time.Time
	Status               TransferStatus  `gorm:"size:20;not null;default:'pending'"`
	ProcessedByID        *uuid.UUID      `gorm:"type:uuid"`
	ProcessedAt          *time.Time
	Message              string          `gorm:"size:500"`
	TransferReason       TransferReason  `gorm:"size:50"`
	PreserveHistory      bool            `gorm:"default:true"`
	NotifyWatchers       bool            `gorm:"default:true"`
	CreatedAt            time.Time
	UpdatedAt            time.Time

	// Relations
	Gist               Gist          `gorm:"constraint:OnDelete:CASCADE"`
	FromUser           *User         `gorm:"foreignKey:FromUserID;constraint:OnDelete:CASCADE"`
	FromOrganization   *Organization `gorm:"foreignKey:FromOrganizationID;constraint:OnDelete:CASCADE"`
	ToUser             *User         `gorm:"foreignKey:ToUserID;constraint:OnDelete:CASCADE"`
	ToOrganization     *Organization `gorm:"foreignKey:ToOrganizationID;constraint:OnDelete:CASCADE"`
	RequestedBy        User          `gorm:"foreignKey:RequestedByID;constraint:OnDelete:CASCADE"`
	ProcessedBy        *User         `gorm:"foreignKey:ProcessedByID;constraint:OnDelete:SET NULL"`
}

// TransferHistory records completed transfers
type TransferHistory struct {
	ID                      uuid.UUID      `gorm:"type:uuid;primary_key"`
	GistID                  uuid.UUID      `gorm:"type:uuid;not null"`
	TransferRequestID       *uuid.UUID     `gorm:"type:uuid"`
	PreviousUserID          *uuid.UUID     `gorm:"type:uuid"`
	PreviousOrganizationID  *uuid.UUID     `gorm:"type:uuid"`
	NewUserID               *uuid.UUID     `gorm:"type:uuid"`
	NewOrganizationID       *uuid.UUID     `gorm:"type:uuid"`
	TransferredByID         uuid.UUID      `gorm:"type:uuid;not null"`
	TransferredAt           time.Time
	TransferReason          TransferReason `gorm:"size:50"`
	Notes                   string         `gorm:"size:500"`
	RepositoryPathBefore    string         `gorm:"size:255"`
	RepositoryPathAfter     string         `gorm:"size:255"`
	GitCommitHash           string         `gorm:"size:40"`

	// Relations
	Gist                   Gist             `gorm:"constraint:OnDelete:CASCADE"`
	TransferRequest        *TransferRequest `gorm:"constraint:OnDelete:SET NULL"`
	PreviousUser           *User            `gorm:"foreignKey:PreviousUserID;constraint:OnDelete:SET NULL"`
	PreviousOrganization   *Organization    `gorm:"foreignKey:PreviousOrganizationID;constraint:OnDelete:SET NULL"`
	NewUser                *User            `gorm:"foreignKey:NewUserID;constraint:OnDelete:SET NULL"`
	NewOrganization        *Organization    `gorm:"foreignKey:NewOrganizationID;constraint:OnDelete:SET NULL"`
	TransferredBy          User             `gorm:"foreignKey:TransferredByID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hooks
func (t *TransferRequest) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	t.RequestedAt = time.Now()
	if t.ExpiresAt.IsZero() {
		t.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	}
	
	// Validate source and destination
	if !t.validateOwnership() {
		return gorm.ErrInvalidData
	}
	return nil
}

func (h *TransferHistory) BeforeCreate(tx *gorm.DB) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	h.TransferredAt = time.Now()
	return nil
}

// validateOwnership checks that source and destination are valid
func (t *TransferRequest) validateOwnership() bool {
	// Must have exactly one source
	hasFromUser := t.FromUserID != nil
	hasFromOrg := t.FromOrganizationID != nil
	if (hasFromUser && hasFromOrg) || (!hasFromUser && !hasFromOrg) {
		return false
	}

	// Must have exactly one destination
	hasToUser := t.ToUserID != nil
	hasToOrg := t.ToOrganizationID != nil
	if (hasToUser && hasToOrg) || (!hasToUser && !hasToOrg) {
		return false
	}

	// Cannot transfer to same owner
	if hasFromUser && hasToUser && *t.FromUserID == *t.ToUserID {
		return false
	}
	if hasFromOrg && hasToOrg && *t.FromOrganizationID == *t.ToOrganizationID {
		return false
	}

	return true
}