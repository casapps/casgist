package opengist

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OpenGist database models for migration
// These match the OpenGist database schema

type OpenGistUser struct {
	ID        uint      `gorm:"primaryKey"`
	Username  string    `gorm:"size:24"`
	Email     string    `gorm:"size:255"`
	Password  string    `gorm:"size:255"`
	IsAdmin   bool      `gorm:"default:false"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (OpenGistUser) TableName() string {
	return "users"
}

type OpenGistGist struct {
	ID          uint      `gorm:"primaryKey"`
	Uuid        string    `gorm:"size:36"`
	Title       string    `gorm:"size:250"`
	Preview     string    `gorm:"type:text"`
	Description string    `gorm:"type:text"`
	Private     int       `gorm:"size:1;default:0"`
	UserID      uint      `gorm:"column:user_id"`
	User        OpenGistUser
	NbFiles     int       `gorm:"column:nb_files;default:0"`
	NbLikes     int       `gorm:"column:nb_likes;default:0"`
	NbForks     int       `gorm:"column:nb_forks;default:0"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (OpenGistGist) TableName() string {
	return "gists"
}

type OpenGistSSHKey struct {
	ID        uint      `gorm:"primaryKey"`
	Title     string    `gorm:"size:50"`
	Content   string    `gorm:"type:text"`
	SHA       string    `gorm:"size:44;column:sha"`
	UserID    uint      `gorm:"column:user_id"`
	User      OpenGistUser
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (OpenGistSSHKey) TableName() string {
	return "ssh_keys"
}

type OpenGistLike struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"column:user_id"`
	User      OpenGistUser
	GistID    uint      `gorm:"column:gist_id"`
	Gist      OpenGistGist
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (OpenGistLike) TableName() string {
	return "likes"
}

// MigrationResult contains the results of a migration operation
type MigrationResult struct {
	UsersImported     int
	GistsImported     int
	FilesImported     int
	SSHKeysImported   int
	LikesImported     int
	ForksImported     int
	Errors            []error
	SkippedUsers      []string
	SkippedGists      []string
	UserIDMapping     map[uint]uuid.UUID // OpenGist ID -> CasGists UUID
	GistIDMapping     map[uint]uuid.UUID // OpenGist ID -> CasGists UUID
	DurationSeconds   int64
	PasswordsReset    bool
	GeneratedPasswords map[string]string // username -> generated password
}

// MigrationOptions configures the migration process
type MigrationOptions struct {
	// Source database connection
	SourceDB *gorm.DB
	
	// Target database connection
	TargetDB *gorm.DB
	
	// Path to OpenGist repository directory
	RepositoryPath string
	
	// Whether to reset all user passwords
	ResetPasswords bool
	
	// Whether to preserve timestamps
	PreserveTimestamps bool
	
	// Whether to migrate SSH keys
	MigrateSSHKeys bool
	
	// Whether to migrate private gists
	MigratePrivateGists bool
	
	// Batch size for bulk operations
	BatchSize int
	
	// Progress callback
	ProgressCallback func(message string, current, total int)
}