package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// PathConfig manages all path-related configuration with environment variable substitution
type PathConfig struct {
	// Core directory variables
	DataDir  string `mapstructure:"data_dir" env:"CASGISTS_DATA_DIR"`
	LogDir   string `mapstructure:"log_dir" env:"CASGISTS_LOG_DIR"`
	CacheDir string `mapstructure:"cache_dir" env:"CASGISTS_CACHE_DIR"`
	TempDir  string `mapstructure:"temp_dir" env:"CASGISTS_TEMP_DIR"`

	// Data subdirectories (use variable substitution)
	DatabasePath   string `mapstructure:"db_path" env:"CASGISTS_DB_PATH"`
	StoragePath    string `mapstructure:"storage_path" env:"CASGISTS_STORAGE_PATH"`
	RepositoryDir  string `mapstructure:"repo_dir" env:"CASGISTS_REPO_DIR"`
	BackupDir      string `mapstructure:"backup_dir" env:"CASGISTS_BACKUP_DIR"`
	SSLDir         string `mapstructure:"ssl_dir" env:"CASGISTS_SSL_DIR"`

	// External certificate paths (when not using auto-SSL)
	TLSCertPath string `mapstructure:"tls_cert_path" env:"CASGISTS_TLS_CERT_PATH"`
	TLSKeyPath  string `mapstructure:"tls_key_path" env:"CASGISTS_TLS_KEY_PATH"`

	// Runtime resolved paths (after variable substitution)
	resolved map[string]string
}

// DirectorySet represents platform-specific directory paths
type DirectorySet struct {
	Privileged DirectoryPaths
	UserMode   DirectoryPaths
}

type DirectoryPaths struct {
	Data   string
	Config string
	Logs   string
	Run    string
	Cache  string
	Temp   string
}

// Platform-specific directory configurations
var directoryConfig = map[string]DirectorySet{
	"linux": {
		Privileged: DirectoryPaths{
			Data:   "/var/lib/casgists",
			Config: "/etc/casgists",
			Logs:   "/var/log/casgists",
			Run:    "/run/casgists",
			Cache:  "/var/cache/casgists",
			Temp:   "/tmp/casgists",
		},
		UserMode: DirectoryPaths{
			Data:   "~/.local/share/casgists",
			Config: "~/.config/casgists",
			Logs:   "~/.local/share/casgists/logs",
			Run:    "~/.local/share/casgists/run",
			Cache:  "~/.cache/casgists",
			Temp:   "~/.local/share/casgists/tmp",
		},
	},
	"darwin": {
		Privileged: DirectoryPaths{
			Data:   "/usr/local/var/lib/casgists",
			Config: "/usr/local/etc/casgists",
			Logs:   "/usr/local/var/log/casgists",
			Run:    "/usr/local/var/run/casgists",
			Cache:  "/usr/local/var/cache/casgists",
			Temp:   "/tmp/casgists",
		},
		UserMode: DirectoryPaths{
			Data:   "~/Library/Application Support/CasGists",
			Config: "~/Library/Preferences/CasGists",
			Logs:   "~/Library/Logs/CasGists",
			Run:    "~/Library/Application Support/CasGists/run",
			Cache:  "~/Library/Caches/CasGists",
			Temp:   "~/Library/Caches/CasGists/tmp",
		},
	},
	"windows": {
		Privileged: DirectoryPaths{
			Data:   "%PROGRAMDATA%\\casgists",
			Config: "%PROGRAMDATA%\\casgists\\config",
			Logs:   "%PROGRAMDATA%\\casgists\\logs",
			Run:    "%PROGRAMDATA%\\casgists\\run",
			Cache:  "%PROGRAMDATA%\\casgists\\cache",
			Temp:   "%TEMP%\\casgists",
		},
		UserMode: DirectoryPaths{
			Data:   "%USERPROFILE%\\AppData\\Local\\casgists",
			Config: "%USERPROFILE%\\AppData\\Local\\casgists\\config",
			Logs:   "%USERPROFILE%\\AppData\\Local\\casgists\\logs",
			Run:    "%USERPROFILE%\\AppData\\Local\\casgists\\run",
			Cache:  "%USERPROFILE%\\AppData\\Local\\casgists\\cache",
			Temp:   "%USERPROFILE%\\AppData\\Local\\casgists\\temp",
		},
	},
}

// NewPathConfig creates a new path configuration with platform defaults
func NewPathConfig(privileged bool) *PathConfig {
	config := &PathConfig{
		resolved: make(map[string]string),
	}

	// Get platform-specific defaults
	platformConfig, exists := directoryConfig[runtime.GOOS]
	if !exists {
		// Fallback to Linux defaults
		platformConfig = directoryConfig["linux"]
	}

	var paths DirectoryPaths
	if privileged {
		paths = platformConfig.Privileged
	} else {
		paths = platformConfig.UserMode
	}

	// Set defaults from environment or platform defaults
	config.DataDir = getEnvOrDefault("CASGISTS_DATA_DIR", paths.Data)
	config.LogDir = getEnvOrDefault("CASGISTS_LOG_DIR", paths.Logs)
	config.CacheDir = getEnvOrDefault("CASGISTS_CACHE_DIR", paths.Cache)
	config.TempDir = getEnvOrDefault("CASGISTS_TEMP_DIR", paths.Temp)

	// Set subdirectory defaults with variable substitution
	config.DatabasePath = getEnvOrDefault("CASGISTS_DB_PATH", "{CASGISTS_DATA_DIR}/data.db")
	config.StoragePath = getEnvOrDefault("CASGISTS_STORAGE_PATH", "{CASGISTS_DATA_DIR}/files")
	config.RepositoryDir = getEnvOrDefault("CASGISTS_REPO_DIR", "{CASGISTS_DATA_DIR}/repositories")
	config.BackupDir = getEnvOrDefault("CASGISTS_BACKUP_DIR", "{CASGISTS_DATA_DIR}/backups")
	config.SSLDir = getEnvOrDefault("CASGISTS_SSL_DIR", "{CASGISTS_DATA_DIR}/ssl")

	// External SSL certificates (optional)
	config.TLSCertPath = getEnvOrDefault("CASGISTS_TLS_CERT_PATH", "")
	config.TLSKeyPath = getEnvOrDefault("CASGISTS_TLS_KEY_PATH", "")

	return config
}

// ResolveAll resolves all path variables and returns resolved paths
func (p *PathConfig) ResolveAll() error {
	// Build variable map
	variables := map[string]string{
		"CASGISTS_DATA_DIR":  p.DataDir,
		"CASGISTS_LOG_DIR":   p.LogDir,
		"CASGISTS_CACHE_DIR": p.CacheDir,
		"CASGISTS_TEMP_DIR":  p.TempDir,
	}

	// First pass: resolve core directories
	var err error
	variables["CASGISTS_DATA_DIR"], err = p.expandPath(p.DataDir)
	if err != nil {
		return fmt.Errorf("failed to resolve data dir: %w", err)
	}

	variables["CASGISTS_LOG_DIR"], err = p.expandPath(p.LogDir)
	if err != nil {
		return fmt.Errorf("failed to resolve log dir: %w", err)
	}

	variables["CASGISTS_CACHE_DIR"], err = p.expandPath(p.CacheDir)
	if err != nil {
		return fmt.Errorf("failed to resolve cache dir: %w", err)
	}

	variables["CASGISTS_TEMP_DIR"], err = p.expandPath(p.TempDir)
	if err != nil {
		return fmt.Errorf("failed to resolve temp dir: %w", err)
	}

	// Second pass: resolve subdirectories with variable substitution
	p.resolved["DatabasePath"], err = p.substituteVariables(p.DatabasePath, variables)
	if err != nil {
		return fmt.Errorf("failed to resolve database path: %w", err)
	}

	p.resolved["StoragePath"], err = p.substituteVariables(p.StoragePath, variables)
	if err != nil {
		return fmt.Errorf("failed to resolve storage path: %w", err)
	}

	p.resolved["RepositoryDir"], err = p.substituteVariables(p.RepositoryDir, variables)
	if err != nil {
		return fmt.Errorf("failed to resolve repository dir: %w", err)
	}

	p.resolved["BackupDir"], err = p.substituteVariables(p.BackupDir, variables)
	if err != nil {
		return fmt.Errorf("failed to resolve backup dir: %w", err)
	}

	p.resolved["SSLDir"], err = p.substituteVariables(p.SSLDir, variables)
	if err != nil {
		return fmt.Errorf("failed to resolve ssl dir: %w", err)
	}

	// Resolve external SSL paths if provided
	if p.TLSCertPath != "" {
		p.resolved["TLSCertPath"], err = p.expandPath(p.TLSCertPath)
		if err != nil {
			return fmt.Errorf("failed to resolve TLS cert path: %w", err)
		}
	}

	if p.TLSKeyPath != "" {
		p.resolved["TLSKeyPath"], err = p.expandPath(p.TLSKeyPath)
		if err != nil {
			return fmt.Errorf("failed to resolve TLS key path: %w", err)
		}
	}

	return nil
}

// GetDatabasePath returns the resolved database path
func (p *PathConfig) GetDatabasePath() string {
	return p.resolved["DatabasePath"]
}

// GetStoragePath returns the resolved storage path
func (p *PathConfig) GetStoragePath() string {
	return p.resolved["StoragePath"]
}

// GetRepositoryDir returns the resolved repository directory
func (p *PathConfig) GetRepositoryDir() string {
	return p.resolved["RepositoryDir"]
}

// GetBackupDir returns the resolved backup directory
func (p *PathConfig) GetBackupDir() string {
	return p.resolved["BackupDir"]
}

// GetSSLDir returns the resolved SSL directory
func (p *PathConfig) GetSSLDir() string {
	return p.resolved["SSLDir"]
}

// GetTLSCertPath returns the resolved TLS certificate path
func (p *PathConfig) GetTLSCertPath() string {
	return p.resolved["TLSCertPath"]
}

// GetTLSKeyPath returns the resolved TLS key path
func (p *PathConfig) GetTLSKeyPath() string {
	return p.resolved["TLSKeyPath"]
}

// GetLogDir returns the log directory
func (p *PathConfig) GetLogDir() string {
	if dir := os.Getenv("CASGISTS_LOG_DIR"); dir != "" {
		return dir
	}
	return p.LogDir
}

// CreateDirectories creates all necessary directories with proper permissions
func (p *PathConfig) CreateDirectories() error {
	directories := []string{
		p.resolved["DatabasePath"],
		p.resolved["StoragePath"],
		p.resolved["RepositoryDir"],
		p.resolved["BackupDir"],
		p.resolved["SSLDir"],
	}

	for _, dir := range directories {
		if dir == "" {
			continue
		}

		// Get directory part (remove filename if present)
		dirPath := filepath.Dir(dir)
		
		// Create directory with proper permissions
		if err := os.MkdirAll(dirPath, 0750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
		}
	}

	return nil
}

// substituteVariables replaces variables in the format {VARIABLE_NAME} with their values
func (p *PathConfig) substituteVariables(path string, variables map[string]string) (string, error) {
	result := path
	
	for varName, varValue := range variables {
		placeholder := "{" + varName + "}"
		if strings.Contains(result, placeholder) {
			expandedValue, err := p.expandPath(varValue)
			if err != nil {
				return "", fmt.Errorf("failed to expand variable %s: %w", varName, err)
			}
			result = strings.ReplaceAll(result, placeholder, expandedValue)
		}
	}

	return result, nil
}

// expandPath expands environment variables and home directory references
func (p *PathConfig) expandPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	// Expand environment variables
	expanded := os.ExpandEnv(path)

	// Handle home directory expansion
	if strings.HasPrefix(expanded, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		expanded = filepath.Join(homeDir, expanded[2:])
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(expanded)
	if err != nil {
		return "", fmt.Errorf("failed to convert to absolute path: %w", err)
	}

	return absPath, nil
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(envVar, defaultValue string) string {
	if value := os.Getenv(envVar); value != "" {
		return value
	}
	return defaultValue
}

// IsPrivileged detects if the application is running with elevated privileges
func IsPrivileged() bool {
	switch runtime.GOOS {
	case "linux", "darwin":
		return os.Geteuid() == 0
	case "windows":
		// Check if running as admin on Windows
		// This is a simplified check - in practice would need more sophisticated detection
		return false // TODO: Implement proper Windows admin detection
	default:
		return false
	}
}

// ValidatePaths validates that all required paths are accessible
func (p *PathConfig) ValidatePaths() error {
	paths := map[string]string{
		"database": p.GetDatabasePath(),
		"storage":  p.GetStoragePath(),
		"repo":     p.GetRepositoryDir(),
		"backup":   p.GetBackupDir(),
		"ssl":      p.GetSSLDir(),
	}

	for name, path := range paths {
		if path == "" {
			return fmt.Errorf("%s path is empty", name)
		}

		// Check if parent directory is writable
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0750); err != nil {
			return fmt.Errorf("cannot create %s directory %s: %w", name, dir, err)
		}

		// Test write permissions
		testFile := filepath.Join(dir, ".casgists-test")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			return fmt.Errorf("no write permission for %s directory %s: %w", name, dir, err)
		}
		os.Remove(testFile) // Clean up
	}

	return nil
}