package notifications

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service handles notifications
type Service struct {
	db *gorm.DB
}

// NewService creates a new notifications service
func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

// SendNotification sends a notification (placeholder implementation)
func (s *Service) SendNotification(userID uuid.UUID, title, message string) error {
	// This would implement actual notification delivery
	// For now, it's a placeholder
	return nil
}