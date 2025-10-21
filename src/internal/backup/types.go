package backup

import (
	"time"
)

// BackupMetadata contains metadata about the backup
type BackupMetadata struct {
	Version        string    `json:"version"`
	CreatedAt      time.Time `json:"created_at"`
	CasGistVersion string    `json:"casgist_version"`
	DatabaseType   string    `json:"database_type"`
	TotalUsers     int64     `json:"total_users"`
	TotalGists     int64     `json:"total_gists"`
	TotalOrgs      int64     `json:"total_orgs"`
	Hostname       string    `json:"hostname"`
	Encrypted      bool      `json:"encrypted"`
}

// BackupOptions contains options for creating a backup
type BackupOptions struct {
	IncludeGitRepos    bool
	IncludeAttachments bool
	IncludeLogs        bool
	IncludeConfigs     bool
	EncryptionKey      string
	OutputPath         string
	MaxSizeMB          int64
}

// BackupResult contains the result of a backup operation
type BackupResult struct {
	ID                  string
	Success             bool
	StartTime           time.Time
	EndTime             time.Time
	Size                int64
	OutputPath          string
	DatabaseExported    bool
	GitReposExported    bool
	AttachmentsExported bool
	Errors              []string
}