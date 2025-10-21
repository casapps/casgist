package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SystemConfig stores system-wide configuration values
type SystemConfig struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	Key       string    `gorm:"uniqueIndex;not null"`
	Value     string    `gorm:"type:text"`
	Type      string    `gorm:"size:50;default:'string'"` // string, int, bool, json
	Category  string    `gorm:"size:50;default:'general'"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// BeforeCreate sets default values before creating
func (s *SystemConfig) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// SystemConfigDefaults contains default system configurations
var SystemConfigDefaults = []SystemConfig{
	{
		Key:      "setup_completed",
		Value:    "false",
		Type:     "bool",
		Category: "setup",
	},
	{
		Key:      "server_port",
		Value:    "0", // 0 means not configured yet
		Type:     "int",
		Category: "server",
	},
	{
		Key:      "server_url",
		Value:    "",
		Type:     "string",
		Category: "server",
	},
	{
		Key:      "registration_enabled",
		Value:    "true",
		Type:     "bool",
		Category: "features",
	},
	{
		Key:      "organizations_enabled",
		Value:    "true",
		Type:     "bool",
		Category: "features",
	},
	{
		Key:      "webhooks_enabled",
		Value:    "true",
		Type:     "bool",
		Category: "features",
	},
	{
		Key:      "public_gists_enabled",
		Value:    "true",
		Type:     "bool",
		Category: "features",
	},
	{
		Key:      "email_enabled",
		Value:    "false",
		Type:     "bool",
		Category: "email",
	},
	{
		Key:      "cache_type",
		Value:    "memory",
		Type:     "string",
		Category: "cache",
	},
	{
		Key:      "search_backend",
		Value:    "sqlite",
		Type:     "string",
		Category: "search",
	},
	{
		Key:      "max_file_size",
		Value:    "5242880", // 5MB in bytes
		Type:     "int",
		Category: "limits",
	},
	{
		Key:      "max_files_per_gist",
		Value:    "100",
		Type:     "int",
		Category: "limits",
	},
	{
		Key:      "max_gist_size",
		Value:    "26214400", // 25MB in bytes
		Type:     "int",
		Category: "limits",
	},
	{
		Key:      "password_min_length",
		Value:    "12",
		Type:     "int",
		Category: "security",
	},
	{
		Key:      "password_require_uppercase",
		Value:    "true",
		Type:     "bool",
		Category: "security",
	},
	{
		Key:      "password_require_lowercase",
		Value:    "true",
		Type:     "bool",
		Category: "security",
	},
	{
		Key:      "password_require_numbers",
		Value:    "true",
		Type:     "bool",
		Category: "security",
	},
	{
		Key:      "session_timeout_hours",
		Value:    "8",
		Type:     "int",
		Category: "security",
	},
	{
		Key:      "jwt_access_token_hours",
		Value:    "2",
		Type:     "int",
		Category: "security",
	},
	{
		Key:      "jwt_refresh_token_days",
		Value:    "3",
		Type:     "int",
		Category: "security",
	},
	{
		Key:      "rate_limit_authenticated",
		Value:    "1000",
		Type:     "int",
		Category: "rate_limit",
	},
	{
		Key:      "rate_limit_anonymous",
		Value:    "100",
		Type:     "int",
		Category: "rate_limit",
	},
	{
		Key:      "rate_limit_login_attempts",
		Value:    "5",
		Type:     "int",
		Category: "rate_limit",
	},
}

// GetConfigValue gets a configuration value by key
func GetConfigValue(db *gorm.DB, key string) (string, error) {
	var config SystemConfig
	if err := db.Where("key = ?", key).First(&config).Error; err != nil {
		return "", err
	}
	return config.Value, nil
}

// SetConfigValue sets a configuration value by key
func SetConfigValue(db *gorm.DB, key, value string) error {
	config := SystemConfig{
		Key:       key,
		Value:     value,
		UpdatedAt: time.Now(),
	}
	
	return db.Where("key = ?", key).
		Assign(config).
		FirstOrCreate(&config).Error
}

// GetConfigBool gets a boolean configuration value
func GetConfigBool(db *gorm.DB, key string) (bool, error) {
	value, err := GetConfigValue(db, key)
	if err != nil {
		return false, err
	}
	return value == "true", nil
}

// GetConfigInt gets an integer configuration value
func GetConfigInt(db *gorm.DB, key string) (int, error) {
	value, err := GetConfigValue(db, key)
	if err != nil {
		return 0, err
	}
	
	var result int
	if _, err := fmt.Sscanf(value, "%d", &result); err != nil {
		return 0, err
	}
	return result, nil
}

// InitializeSystemConfig initializes system configuration with defaults
func InitializeSystemConfig(db *gorm.DB) error {
	for _, config := range SystemConfigDefaults {
		var existing SystemConfig
		if err := db.Where("key = ?", config.Key).First(&existing).Error; err != nil {
			// Config doesn't exist, create it
			if err := db.Create(&config).Error; err != nil {
				return fmt.Errorf("failed to create config %s: %w", config.Key, err)
			}
		}
	}
	return nil
}