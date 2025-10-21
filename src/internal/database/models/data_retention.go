package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DataRetentionPolicy defines retention rules for different data types
type DataRetentionPolicy struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DataType       string    `gorm:"type:varchar(100);unique;not null" json:"data_type"`
	RetentionDays  int       `gorm:"not null" json:"retention_days"`
	Description    string    `gorm:"type:text" json:"description"`
	IsActive       bool      `gorm:"default:true" json:"is_active"`
	DeleteStrategy string    `gorm:"type:varchar(50);default:'soft'" json:"delete_strategy"` // soft, hard, archive
	LastRunAt      *time.Time `json:"last_run_at,omitempty"`
	CreatedAt      time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// DataRetentionLog tracks retention policy executions
type DataRetentionLog struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PolicyID      uuid.UUID  `gorm:"type:uuid;not null" json:"policy_id"`
	Policy        *DataRetentionPolicy `gorm:"foreignKey:PolicyID" json:"policy,omitempty"`
	ExecutedAt    time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"executed_at"`
	RecordsAffected int64    `gorm:"default:0" json:"records_affected"`
	Status        string     `gorm:"type:varchar(50);not null" json:"status"` // success, failed, partial
	ErrorMessage  string     `gorm:"type:text" json:"error_message,omitempty"`
	Duration      int        `json:"duration"` // in milliseconds
}

// DefaultRetentionPolicies returns the default data retention policies
func DefaultRetentionPolicies() []DataRetentionPolicy {
	return []DataRetentionPolicy{
		{
			DataType:       "activity_feeds",
			RetentionDays:  90,
			Description:    "Activity feed entries older than 90 days",
			DeleteStrategy: "hard",
		},
		{
			DataType:       "search_history",
			RetentionDays:  30,
			Description:    "Search history older than 30 days",
			DeleteStrategy: "hard",
		},
		{
			DataType:       "webhook_deliveries",
			RetentionDays:  30,
			Description:    "Webhook delivery logs older than 30 days",
			DeleteStrategy: "hard",
		},
		{
			DataType:       "audit_logs",
			RetentionDays:  365,
			Description:    "Audit logs older than 1 year",
			DeleteStrategy: "archive",
		},
		{
			DataType:       "deleted_gists",
			RetentionDays:  30,
			Description:    "Soft-deleted gists older than 30 days",
			DeleteStrategy: "hard",
		},
		{
			DataType:       "deleted_users",
			RetentionDays:  90,
			Description:    "Soft-deleted user accounts older than 90 days",
			DeleteStrategy: "hard",
		},
		{
			DataType:       "sessions",
			RetentionDays:  7,
			Description:    "Expired sessions older than 7 days",
			DeleteStrategy: "hard",
		},
		{
			DataType:       "temp_files",
			RetentionDays:  1,
			Description:    "Temporary files older than 1 day",
			DeleteStrategy: "hard",
		},
	}
}

// InitializeRetentionPolicies creates default retention policies if they don't exist
func InitializeRetentionPolicies(db *gorm.DB) error {
	policies := DefaultRetentionPolicies()

	for _, policy := range policies {
		var existing DataRetentionPolicy
		err := db.Where("data_type = ?", policy.DataType).First(&existing).Error

		if err == gorm.ErrRecordNotFound {
			// Create new policy
			if err := db.Create(&policy).Error; err != nil {
				return fmt.Errorf("failed to create retention policy for %s: %w", policy.DataType, err)
			}
		}
	}

	return nil
}

// GetActiveRetentionPolicies retrieves all active retention policies
func GetActiveRetentionPolicies(db *gorm.DB) ([]DataRetentionPolicy, error) {
	var policies []DataRetentionPolicy
	err := db.Where("is_active = ?", true).Find(&policies).Error
	return policies, err
}

// ExecuteRetentionPolicy executes a specific retention policy
func ExecuteRetentionPolicy(db *gorm.DB, policyID uuid.UUID) (*DataRetentionLog, error) {
	startTime := time.Now()

	// Get the policy
	var policy DataRetentionPolicy
	if err := db.First(&policy, "id = ?", policyID).Error; err != nil {
		return nil, fmt.Errorf("policy not found: %w", err)
	}

	if !policy.IsActive {
		return nil, fmt.Errorf("policy is not active")
	}

	// Calculate cutoff date
	cutoff := time.Now().AddDate(0, 0, -policy.RetentionDays)

	// Execute retention based on data type
	var recordsAffected int64
	var err error

	switch policy.DataType {
	case "activity_feeds":
		recordsAffected, err = cleanupActivityFeeds(db, cutoff, policy.DeleteStrategy)
	case "search_history":
		recordsAffected, err = cleanupSearchHistory(db, cutoff, policy.DeleteStrategy)
	case "webhook_deliveries":
		recordsAffected, err = cleanupWebhookDeliveries(db, cutoff, policy.DeleteStrategy)
	case "audit_logs":
		recordsAffected, err = cleanupAuditLogs(db, cutoff, policy.DeleteStrategy)
	case "deleted_gists":
		recordsAffected, err = cleanupDeletedGists(db, cutoff, policy.DeleteStrategy)
	case "deleted_users":
		recordsAffected, err = cleanupDeletedUsers(db, cutoff, policy.DeleteStrategy)
	case "sessions":
		recordsAffected, err = cleanupSessions(db, cutoff, policy.DeleteStrategy)
	case "temp_files":
		recordsAffected, err = cleanupTempFiles(db, cutoff, policy.DeleteStrategy)
	default:
		err = fmt.Errorf("unknown data type: %s", policy.DataType)
	}

	// Create log entry
	log := &DataRetentionLog{
		PolicyID:        policyID,
		ExecutedAt:      time.Now(),
		RecordsAffected: recordsAffected,
		Duration:        int(time.Since(startTime).Milliseconds()),
	}

	if err != nil {
		log.Status = "failed"
		log.ErrorMessage = err.Error()
	} else {
		log.Status = "success"
		// Update policy last run time
		now := time.Now()
		db.Model(&policy).Update("last_run_at", now)
	}

	// Save log
	if err := db.Create(log).Error; err != nil {
		return log, fmt.Errorf("failed to save log: %w", err)
	}

	return log, nil
}

// Cleanup functions for each data type

func cleanupActivityFeeds(db *gorm.DB, cutoff time.Time, strategy string) (int64, error) {
	query := db.Where("created_at < ?", cutoff)

	if strategy == "hard" {
		result := query.Delete(&ActivityFeed{})
		return result.RowsAffected, result.Error
	}

	// For soft delete, GORM handles it automatically
	result := query.Delete(&ActivityFeed{})
	return result.RowsAffected, result.Error
}

func cleanupSearchHistory(db *gorm.DB, cutoff time.Time, strategy string) (int64, error) {
	result := db.Where("created_at < ?", cutoff).Delete(&SearchHistory{})
	return result.RowsAffected, result.Error
}

func cleanupWebhookDeliveries(db *gorm.DB, cutoff time.Time, strategy string) (int64, error) {
	result := db.Where("created_at < ?", cutoff).Delete(&WebhookDelivery{})
	return result.RowsAffected, result.Error
}

func cleanupAuditLogs(db *gorm.DB, cutoff time.Time, strategy string) (int64, error) {
	if strategy == "archive" {
		// Archive to separate table or export to file
		// For now, just count the records that would be archived
		var count int64
		err := db.Model(&AuditLog{}).Where("created_at < ?", cutoff).Count(&count).Error
		return count, err
	}

	result := db.Where("created_at < ?", cutoff).Delete(&AuditLog{})
	return result.RowsAffected, result.Error
}

func cleanupDeletedGists(db *gorm.DB, cutoff time.Time, strategy string) (int64, error) {
	// Permanently delete soft-deleted gists
	result := db.Unscoped().Where("deleted_at < ?", cutoff).Delete(&Gist{})
	return result.RowsAffected, result.Error
}

func cleanupDeletedUsers(db *gorm.DB, cutoff time.Time, strategy string) (int64, error) {
	// Permanently delete soft-deleted users
	result := db.Unscoped().Where("deleted_at < ?", cutoff).Delete(&User{})
	return result.RowsAffected, result.Error
}

func cleanupSessions(db *gorm.DB, cutoff time.Time, strategy string) (int64, error) {
	// Clean up expired sessions
	result := db.Where("expires_at < ?", cutoff).Delete(&Session{})
	return result.RowsAffected, result.Error
}

func cleanupTempFiles(db *gorm.DB, cutoff time.Time, strategy string) (int64, error) {
	// This would typically clean up actual temp files from filesystem
	// For now, just return 0 as placeholder
	return 0, nil
}

// RunRetentionPolicies runs all active retention policies
func RunRetentionPolicies(db *gorm.DB) ([]DataRetentionLog, error) {
	policies, err := GetActiveRetentionPolicies(db)
	if err != nil {
		return nil, err
	}

	var logs []DataRetentionLog

	for _, policy := range policies {
		// Check if policy should run (based on last run time)
		if policy.LastRunAt != nil {
			nextRun := policy.LastRunAt.AddDate(0, 0, 1) // Run daily
			if time.Now().Before(nextRun) {
				continue // Skip this policy
			}
		}

		log, err := ExecuteRetentionPolicy(db, policy.ID)
		if err != nil {
			// Log error but continue with other policies
			log = &DataRetentionLog{
				PolicyID:        policy.ID,
				ExecutedAt:      time.Now(),
				Status:          "failed",
				ErrorMessage:    err.Error(),
			}
			db.Create(log)
		}

		if log != nil {
			logs = append(logs, *log)
		}
	}

	return logs, nil
}

// GetRetentionLogs retrieves retention policy execution logs
func GetRetentionLogs(db *gorm.DB, limit, offset int) ([]DataRetentionLog, error) {
	var logs []DataRetentionLog
	err := db.Preload("Policy").
		Order("executed_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

// UpdateRetentionPolicy updates a retention policy
func UpdateRetentionPolicy(db *gorm.DB, policyID uuid.UUID, retentionDays int, isActive bool) error {
	return db.Model(&DataRetentionPolicy{}).
		Where("id = ?", policyID).
		Updates(map[string]interface{}{
			"retention_days": retentionDays,
			"is_active":      isActive,
			"updated_at":     time.Now(),
		}).Error
}