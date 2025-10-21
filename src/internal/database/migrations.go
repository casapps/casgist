package database

import (
	"embed"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// MigrationManager handles database migrations
type MigrationManager struct {
	db       *gorm.DB
	migrate  *migrate.Migrate
	dbType   string
	logger   *slog.Logger
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *gorm.DB, dbType string) (*MigrationManager, error) {
	// Get underlying SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Create source driver from embedded files
	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to create source driver: %w", err)
	}

	// Create database driver based on type
	var dbDriver database.Driver
	switch dbType {
	case "sqlite", "sqlite3":
		dbDriver, err = sqlite3.WithInstance(sqlDB, &sqlite3.Config{})
	case "postgres", "postgresql":
		dbDriver, err = postgres.WithInstance(sqlDB, &postgres.Config{})
	case "mysql":
		dbDriver, err = mysql.WithInstance(sqlDB, &mysql.Config{})
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create database driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithInstance("iofs", sourceDriver, dbType, dbDriver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return &MigrationManager{
		db:      db,
		migrate: m,
		dbType:  dbType,
		logger:  slog.Default(),
	}, nil
}

// Up runs all pending migrations
func (m *MigrationManager) Up() error {
	m.logger.Info("Running database migrations")
	
	err := m.migrate.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	
	if err == migrate.ErrNoChange {
		m.logger.Info("No migrations to run")
	} else {
		m.logger.Info("Migrations completed successfully")
	}
	
	return nil
}

// Down rolls back migrations
func (m *MigrationManager) Down(steps int) error {
	m.logger.Info("Rolling back migrations", "steps", steps)
	
	if steps <= 0 {
		return fmt.Errorf("steps must be greater than 0")
	}
	
	err := m.migrate.Steps(-steps)
	if err != nil {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}
	
	m.logger.Info("Rollback completed successfully")
	return nil
}

// Version returns the current migration version
func (m *MigrationManager) Version() (uint, bool, error) {
	return m.migrate.Version()
}

// Force sets the migration version without running migrations
func (m *MigrationManager) Force(version int) error {
	m.logger.Info("Forcing migration version", "version", version)
	
	err := m.migrate.Force(version)
	if err != nil {
		return fmt.Errorf("failed to force migration version: %w", err)
	}
	
	return nil
}

// Close closes the migration manager
func (m *MigrationManager) Close() error {
	sourceErr, dbErr := m.migrate.Close()
	if sourceErr != nil {
		return fmt.Errorf("failed to close source: %w", sourceErr)
	}
	if dbErr != nil {
		return fmt.Errorf("failed to close database: %w", dbErr)
	}
	return nil
}

// RunMigrations is a convenience function to run migrations
func RunMigrations(db *gorm.DB, dbType string) error {
	manager, err := NewMigrationManager(db, dbType)
	if err != nil {
		return fmt.Errorf("failed to create migration manager: %w", err)
	}
	
	// Run migrations
	if err := manager.Up(); err != nil {
		manager.Close()
		return err
	}
	
	// Close migration manager but keep database connection open
	manager.Close()
	return nil
}

// RunCustomMigrations runs any custom migrations after the main migrations
func RunCustomMigrations(db *gorm.DB) error {
	// Get database type
	dbType := db.Dialector.Name()
	
	// Create indexes based on database type
	if err := createIndexes(db, dbType); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}
	
	// Initialize system configuration
	if err := initializeSystemConfig(db); err != nil {
		return fmt.Errorf("failed to initialize system config: %w", err)
	}
	
	return nil
}

// createIndexes creates performance indexes
func createIndexes(db *gorm.DB, dbType string) error {
	var indexes []string
	
	switch dbType {
	case "sqlite":
		indexes = []string{
			// SQLite indexes
			"CREATE INDEX IF NOT EXISTS idx_gists_user_visibility ON gists(user_id, visibility)",
			"CREATE INDEX IF NOT EXISTS idx_gists_created_at ON gists(created_at)",
			"CREATE INDEX IF NOT EXISTS idx_gist_files_gist_id ON gist_files(gist_id)",
			"CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)",
			"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
			"CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token)",
			"CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)",
		}
	case "postgres", "mysql":
		indexes = []string{
			// PostgreSQL/MySQL indexes
			"CREATE INDEX IF NOT EXISTS idx_gists_user_visibility ON gists(user_id, visibility)",
			"CREATE INDEX IF NOT EXISTS idx_gists_org_visibility ON gists(organization_id, visibility)",
			"CREATE INDEX IF NOT EXISTS idx_gists_created_at ON gists(created_at DESC)",
			"CREATE INDEX IF NOT EXISTS idx_gists_updated_at ON gists(updated_at DESC)",
			"CREATE INDEX IF NOT EXISTS idx_gist_files_gist_id ON gist_files(gist_id)",
			"CREATE INDEX IF NOT EXISTS idx_gist_stars_user_gist ON gist_stars(user_id, gist_id)",
			"CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at)",
			"CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC)",
		}
	}
	
	for _, idx := range indexes {
		if err := db.Exec(idx).Error; err != nil {
			// Log warning but don't fail - index might already exist
			slog.Default().Warn("Failed to create index", "index", idx, "error", err)
		}
	}
	
	return nil
}

// initializeSystemConfig creates default system configuration
func initializeSystemConfig(db *gorm.DB) error {
	// This will be called after models are created
	// The models package will handle the actual initialization
	return nil
}

// GetMigrationStatus returns the current migration status
func GetMigrationStatus(db *gorm.DB, dbType string) (map[string]interface{}, error) {
	manager, err := NewMigrationManager(db, dbType)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration manager: %w", err)
	}
	defer manager.Close()

	version, dirty, err := manager.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return nil, fmt.Errorf("failed to get migration version: %w", err)
	}

	status := map[string]interface{}{
		"current_version": version,
		"dirty":           dirty,
		"database_type":   dbType,
	}

	if err == migrate.ErrNilVersion {
		status["current_version"] = 0
		status["message"] = "No migrations have been run yet"
	}

	return status, nil
}