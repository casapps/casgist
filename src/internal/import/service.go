package importer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/casapps/casgists/src/internal/database/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ImportService defines the interface for importing from external platforms
type ImportService interface {
	ValidateToken(ctx context.Context) error
	ImportGists(ctx context.Context, targetUserID uuid.UUID) ([]*models.Gist, []error)
}

// ImportManager manages import operations
type ImportManager struct {
	db *gorm.DB
}

// NewImportManager creates a new import manager
func NewImportManager(db *gorm.DB) *ImportManager {
	return &ImportManager{db: db}
}

// ImportRequest represents an import request
type ImportRequest struct {
	Platform string `json:"platform" validate:"required"`
	Token    string `json:"token" validate:"required"`
	Options  struct {
		ImportAsPrivate   bool     `json:"import_as_private"`
		AddPlatformTag    bool     `json:"add_platform_tag"`
		PreserveDates     bool     `json:"preserve_dates"`
		OrganizationName  string   `json:"organization_name"`
		SelectedGistIDs   []string `json:"selected_gist_ids"`
		GitLabURL         string   `json:"gitlab_url"`
		GiteaURL          string   `json:"gitea_url"`
		ForgejoURL        string   `json:"forgejo_url"`
	} `json:"options"`
}

// ImportResult represents the result of an import operation
type ImportResult struct {
	JobID            uuid.UUID `json:"job_id"`
	Platform         string    `json:"platform"`
	TotalGists       int       `json:"total_gists"`
	SuccessfulImports int       `json:"successful_imports"`
	FailedImports    int       `json:"failed_imports"`
	Errors           []string  `json:"errors"`
	Duration         string    `json:"duration"`
	ImportedGistIDs  []uuid.UUID `json:"imported_gist_ids"`
}

// StartImport starts an import operation
func (im *ImportManager) StartImport(ctx context.Context, userID uuid.UUID, req *ImportRequest) (*ImportResult, error) {
	// Create import job record
	job := &models.ImportJob{
		ID:         uuid.New(),
		UserID:     &userID,
		SourceType: req.Platform,
		Status:     "processing",
		StartedAt:  &time.Time{},
	}

	if err := im.db.Create(job).Error; err != nil {
		return nil, fmt.Errorf("failed to create import job: %w", err)
	}

	// Create appropriate importer
	var importer ImportService
	switch strings.ToLower(req.Platform) {
	case "github":
		importer = NewGitHubImporter(req.Token)
	case "gitlab":
		baseURL := req.Options.GitLabURL
		if baseURL == "" {
			baseURL = "https://gitlab.com"
		}
		importer = NewGitLabImporter(req.Token, baseURL)
	case "gitea":
		if req.Options.GiteaURL == "" {
			return nil, fmt.Errorf("Gitea instance URL required")
		}
		importer = NewGiteaImporter(req.Token, req.Options.GiteaURL)
	case "forgejo":
		if req.Options.ForgejoURL == "" {
			return nil, fmt.Errorf("Forgejo instance URL required")
		}
		importer = NewForgejoImporter(req.Token, req.Options.ForgejoURL)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", req.Platform)
	}

	// Validate token first
	if err := importer.ValidateToken(ctx); err != nil {
		job.Status = "failed"
		job.ErrorMessage = err.Error()
		im.db.Save(job)
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Start import
	startTime := time.Now()
	importedGists, importErrors := importer.ImportGists(ctx, userID)

	// Save imported gists to database
	var savedGistIDs []uuid.UUID
	var successCount int

	for _, gist := range importedGists {
		// Apply import options
		if req.Options.ImportAsPrivate {
			gist.Visibility = models.VisibilityPrivate
		}

		// Add platform tag
		if req.Options.AddPlatformTag {
			if gist.Tags == "" {
				gist.Tags = req.Platform
			} else {
				gist.Tags = fmt.Sprintf("%s,%s", gist.Tags, req.Platform)
			}
		}

		// Preserve creation dates if requested
		if !req.Options.PreserveDates {
			now := time.Now()
			gist.CreatedAt = now
			gist.UpdatedAt = now
		}

		// Save to database
		if err := im.db.Create(gist).Error; err != nil {
			importErrors = append(importErrors, fmt.Errorf("failed to save gist %s: %w", gist.Name, err))
			continue
		}

		savedGistIDs = append(savedGistIDs, gist.ID)
		successCount++

		// Create import item record
		item := &models.ImportItem{
			ID:          uuid.New(),
			ImportJobID: job.ID,
			GistID:      &gist.ID,
			Status:      "completed",
			ProcessedAt: &time.Time{},
		}
		im.db.Create(item)
	}

	// Update job status
	duration := time.Since(startTime)
	job.Status = "completed"
	job.ProcessedItems = len(importedGists) + len(importErrors)
	job.SuccessfulItems = successCount
	job.FailedItems = len(importErrors)
	job.CompletedAt = &time.Time{}
	
	if len(importErrors) > 0 {
		job.ErrorMessage = fmt.Sprintf("%d errors occurred during import", len(importErrors))
	}

	im.db.Save(job)

	// Prepare error strings
	var errorStrings []string
	for _, err := range importErrors {
		errorStrings = append(errorStrings, err.Error())
	}

	return &ImportResult{
		JobID:             job.ID,
		Platform:          req.Platform,
		TotalGists:        len(importedGists) + len(importErrors),
		SuccessfulImports: successCount,
		FailedImports:     len(importErrors),
		Errors:           errorStrings,
		Duration:         duration.String(),
		ImportedGistIDs:  savedGistIDs,
	}, nil
}

// GetImportJob gets import job status
func (im *ImportManager) GetImportJob(jobID uuid.UUID) (*models.ImportJob, error) {
	var job models.ImportJob
	if err := im.db.Preload("Items").First(&job, "id = ?", jobID).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

// ListImportJobs lists import jobs for a user
func (im *ImportManager) ListImportJobs(userID uuid.UUID) ([]models.ImportJob, error) {
	var jobs []models.ImportJob
	if err := im.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}