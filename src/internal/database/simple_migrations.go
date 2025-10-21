package database

import (
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
	
	"gorm.io/gorm"
)

// SimpleMigrations runs SQL migrations without closing the database
func SimpleMigrations(db *gorm.DB) error {
	logger := slog.Default()
	
	// Get database type
	dbType := db.Dialector.Name()
	logger.Info("Running migrations", "database", dbType)
	
	// Read migration files
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}
	
	// Filter and sort .up.sql files
	var upMigrations []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			upMigrations = append(upMigrations, entry.Name())
		}
	}
	
	// Create migrations tracking table
	createTableSQL := `CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		dirty INTEGER NOT NULL DEFAULT 0,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	
	if err := db.Exec(createTableSQL).Error; err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}
	
	// Run each migration
	for _, filename := range upMigrations {
		// Extract version from filename (e.g., "000001_initial_schema.up.sql" -> 1)
		version := 0
		if _, err := fmt.Sscanf(filename, "%06d_", &version); err != nil {
			logger.Warn("Skipping migration with invalid filename", "filename", filename)
			continue
		}
		
		// Check if already applied
		var count int64
		if err := db.Raw("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count).Error; err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}
		
		if count > 0 {
			logger.Info("Migration already applied", "version", version, "filename", filename)
			continue
		}
		
		// Skip SQLite-specific migrations for other databases
		if strings.Contains(filename, "sqlite") && dbType != "sqlite" {
			logger.Info("Skipping SQLite-specific migration", "version", version, "filename", filename)
			continue
		}
		
		// Read migration content
		content, err := fs.ReadFile(migrationsFS, filepath.Join("migrations", filename))
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", filename, err)
		}
		
		// Execute migration
		logger.Info("Applying migration", "version", version, "filename", filename, "size", len(content))
		
		// Don't use transactions for DDL statements in SQLite
		useTransaction := false
		
		var execDB *gorm.DB = db
		
		// Split and execute statements
		statements := splitSQLStatements(string(content))
		logger.Info("Executing statements", "count", len(statements))
		for i, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" || strings.HasPrefix(stmt, "--") {
				continue
			}
			
			// Log first few statements for debugging
			if i < 3 {
				logger.Info("Statement preview", "index", i, "length", len(stmt), "preview", stmt[:min(50, len(stmt))]+"...")
			}
			
			if err := execDB.Exec(stmt).Error; err != nil {
				// Rollback and mark as dirty
				if useTransaction {
					execDB.Rollback()
				}
				db.Exec("INSERT INTO schema_migrations (version, dirty) VALUES (?, 1)", version)
				return fmt.Errorf("failed to execute migration %d statement %d: %w", version, i+1, err)
			}
		}
		
		// Record successful migration
		if err := db.Exec("INSERT INTO schema_migrations (version, dirty) VALUES (?, 0)", version).Error; err != nil {
			return fmt.Errorf("failed to record migration: %w", err)
		}
		
		// Log table count after migration
		if dbType == "sqlite" {
			var tableCount int
			db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&tableCount)
			logger.Info("Migration completed", "version", version, "table_count", tableCount)
		} else {
			logger.Info("Migration completed", "version", version)
		}
	}
	
	logger.Info("All migrations completed successfully")
	return nil
}

// splitSQLStatements splits SQL content by semicolons while respecting SQL syntax
func splitSQLStatements(content string) []string {
	var statements []string
	var currentStmt strings.Builder
	triggerDepth := 0
	
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		upperLine := strings.ToUpper(trimmedLine)
		
		// Track trigger blocks
		if strings.HasPrefix(upperLine, "CREATE TRIGGER") {
			triggerDepth = 1
		}
		
		// Add line to current statement
		currentStmt.WriteString(line)
		currentStmt.WriteString("\n")
		
		// Check for END; that closes a trigger
		if triggerDepth > 0 && upperLine == "END;" {
			statements = append(statements, strings.TrimSpace(currentStmt.String()))
			currentStmt.Reset()
			triggerDepth = 0
			continue
		}
		
		// For non-trigger statements, split on semicolon at end of line
		if triggerDepth == 0 && strings.HasSuffix(trimmedLine, ";") && !strings.HasPrefix(upperLine, "--") {
			statements = append(statements, strings.TrimSpace(currentStmt.String()))
			currentStmt.Reset()
		}
	}
	
	// Add any remaining content
	if currentStmt.Len() > 0 {
		remaining := strings.TrimSpace(currentStmt.String())
		if remaining != "" && !strings.HasPrefix(remaining, "--") {
			statements = append(statements, remaining)
		}
	}
	
	return statements
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}