package database

import (
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"
	
	"gorm.io/gorm"
)

// FastMigrations runs migrations without complex parsing
func FastMigrations(db *gorm.DB) error {
	logger := slog.Default()
	dbType := db.Dialector.Name()
	
	logger.Info("Running fast migrations", "database", dbType)
	
	// Create migrations table
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY)`).Error; err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}
	
	// Read migration files
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}
	
	// Filter and sort .up.sql files
	var migrations []string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".up.sql") {
			migrations = append(migrations, entry.Name())
		}
	}
	sort.Strings(migrations)
	
	// Run each migration
	for _, filename := range migrations {
		// Extract version
		var version int
		fmt.Sscanf(filename, "%06d_", &version)
		
		// Check if already applied
		var count int64
		db.Raw("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count)
		if count > 0 {
			logger.Info("Migration already applied", "version", version)
			continue
		}
		
		// Skip SQLite-specific migrations for other databases
		if strings.Contains(filename, "sqlite") && dbType != "sqlite" {
			logger.Info("Skipping SQLite-specific migration", "version", version)
			continue
		}
		
		// Read content
		content, err := fs.ReadFile(migrationsFS, filepath.Join("migrations", filename))
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", filename, err)
		}
		
		logger.Info("Applying migration", "version", version, "filename", filename)

		// Execute statements one by one for all databases to avoid issues with multiple commands
		if err := executeStatements(db, string(content)); err != nil {
			return fmt.Errorf("migration %d failed: %w", version, err)
		}
		
		// Record success
		db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version)
		logger.Info("Migration completed", "version", version)
	}
	
	logger.Info("All migrations completed")
	return nil
}

// FastMigrationsSkipFTS runs migrations but skips FTS-related migrations for testing
func FastMigrationsSkipFTS(db *gorm.DB) error {
	logger := slog.Default()
	dbType := db.Dialector.Name()
	
	logger.Info("Running fast migrations (skipping FTS)", "database", dbType)
	
	// Create migrations table
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY)`).Error; err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}
	
	// Read migration files
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}
	
	// Filter and sort .up.sql files, excluding FTS ones
	var migrations []string
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(name, ".up.sql") && !strings.Contains(name, "fts") && !strings.Contains(name, "search") {
			migrations = append(migrations, name)
		}
	}
	sort.Strings(migrations)
	
	// Apply each migration
	for _, migrationFile := range migrations {
		// Extract version from filename (e.g., "000001_initial_schema.up.sql" -> 1)
		parts := strings.Split(migrationFile, "_")
		if len(parts) == 0 {
			continue
		}
		
		versionStr := strings.TrimLeft(parts[0], "0")
		if versionStr == "" {
			versionStr = "0"
		}
		
		var version int
		if _, err := fmt.Sscanf(versionStr, "%d", &version); err != nil {
			continue // Skip malformed filenames
		}
		
		// Check if already applied
		var count int64
		db.Raw("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count)
		if count > 0 {
			continue // Already applied
		}
		
		logger.Info("Applying migration", "version", version, "filename", migrationFile)
		
		// Read and execute migration
		content, err := migrationsFS.ReadFile(filepath.Join("migrations", migrationFile))
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", migrationFile, err)
		}
		
		if dbType == "sqlite" {
			if err := executeStatements(db, string(content)); err != nil {
				return fmt.Errorf("migration %d failed: %w", version, err)
			}
		} else {
			if err := db.Exec(string(content)).Error; err != nil {
				return fmt.Errorf("migration %d failed: %w", version, err)
			}
		}
		
		// Record success
		db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version)
		logger.Info("Migration completed", "version", version)
	}
	
	logger.Info("All migrations completed (FTS skipped)")
	return nil
}

// executeStatements executes SQL statements one by one for all database types
func executeStatements(db *gorm.DB, content string) error {
	// Remove comments and empty lines
	lines := strings.Split(content, "\n")
	var cleanContent strings.Builder
	inTrigger := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Track trigger blocks
		if strings.HasPrefix(strings.ToUpper(trimmed), "CREATE TRIGGER") {
			inTrigger = true
		}
		
		// Skip comments
		if strings.HasPrefix(trimmed, "--") {
			continue
		}
		
		cleanContent.WriteString(line)
		cleanContent.WriteString("\n")
		
		// Execute when we hit a semicolon at the end of a line (not in trigger)
		if strings.HasSuffix(trimmed, ";") {
			if inTrigger && strings.ToUpper(trimmed) == "END;" {
				inTrigger = false
			}
			
			if !inTrigger {
				stmt := strings.TrimSpace(cleanContent.String())
				if stmt != "" && !strings.HasPrefix(stmt, "--") {
					if err := db.Exec(stmt).Error; err != nil {
						return fmt.Errorf("failed to execute: %s: %w", stmt[:min(100, len(stmt))], err)
					}
				}
				cleanContent.Reset()
			}
		}
	}
	
	return nil
}