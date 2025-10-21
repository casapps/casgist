package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

// Version information set at build time
var (
	Version   = "dev"
	BuildDate = "unknown"
	GitCommit = "unknown"
	GoVersion = runtime.Version()
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Paths    PathsConfig    `mapstructure:"paths"`
	Security SecurityConfig `mapstructure:"security"`
	Email    EmailConfig    `mapstructure:"email"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Webhook  WebhookConfig  `mapstructure:"webhook"`
	Debug    bool           `mapstructure:"debug"`
}

// ServerConfig contains server configuration
type ServerConfig struct {
	Port       int    `mapstructure:"port"`
	Host       string `mapstructure:"host"`
	PublicURL  string `mapstructure:"public_url"`
	EnableHTTPS bool   `mapstructure:"enable_https"`
	CertFile   string `mapstructure:"cert_file"`
	KeyFile    string `mapstructure:"key_file"`
}

// DatabaseConfig contains database configuration
type DatabaseConfig struct {
	Type           string `mapstructure:"type"`
	DSN            string `mapstructure:"dsn"`
	MaxConnections int    `mapstructure:"max_connections"`
	MaxIdleTime    int    `mapstructure:"max_idle_time"`
}

// PathsConfig contains directory paths
type PathsConfig struct {
	DataDir    string `mapstructure:"data_dir"`
	LogDir     string `mapstructure:"log_dir"`
	CacheDir   string `mapstructure:"cache_dir"`
	RepoDir    string `mapstructure:"repo_dir"`
	BackupDir  string `mapstructure:"backup_dir"`
	UploadDir  string `mapstructure:"upload_dir"`
}

// SecurityConfig contains security settings
type SecurityConfig struct {
	SecretKey      string `mapstructure:"secret_key"`
	Enable2FA      bool   `mapstructure:"enable_2fa"`
	EnableWebAuthn bool   `mapstructure:"enable_webauthn"`
	SessionTimeout int    `mapstructure:"session_timeout"`
}

// EmailConfig contains email settings
type EmailConfig struct {
	SMTPHost     string `mapstructure:"smtp_host"`
	SMTPPort     int    `mapstructure:"smtp_port"`
	SMTPUser     string `mapstructure:"smtp_user"`
	SMTPPassword string `mapstructure:"smtp_password"`
	FromAddress  string `mapstructure:"from_address"`
	FromName     string `mapstructure:"from_name"`
	EnableTLS    bool   `mapstructure:"enable_tls"`
}

// RedisConfig contains Redis/Valkey settings
type RedisConfig struct {
	URL      string `mapstructure:"url"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// WebhookConfig contains webhook settings
type WebhookConfig struct {
	Secret  string `mapstructure:"secret"`
	Timeout int    `mapstructure:"timeout"`
}

// Load loads configuration from file and environment
func Load() (*Config, error) {
	v := viper.New()
	
	// Set config name and paths
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	
	// Add config paths
	v.AddConfigPath(".")
	v.AddConfigPath("/etc/casgists")
	if home, err := os.UserHomeDir(); err == nil {
		v.AddConfigPath(filepath.Join(home, ".casgists"))
	}
	
	// Set environment prefix
	v.SetEnvPrefix("CASGISTS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	
	// Set defaults
	setDefaults(v)
	
	// Read config file if exists
	if err := v.ReadInConfig(); err != nil {
		// It's ok if config file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}
	
	// Unmarshal config
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	// Post-process configuration
	if err := postProcess(&cfg); err != nil {
		return nil, fmt.Errorf("failed to post-process config: %w", err)
	}
	
	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", 0) // Will be randomized in postProcess
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.enable_https", false)
	
	// Database defaults
	v.SetDefault("database.type", "sqlite")
	v.SetDefault("database.dsn", "${DATA_DIR}/casgists.db")
	v.SetDefault("database.max_connections", 25)
	v.SetDefault("database.max_idle_time", 300)
	
	// Path defaults
	v.SetDefault("paths.data_dir", getDefaultDataDir())
	v.SetDefault("paths.log_dir", "${DATA_DIR}/logs")
	v.SetDefault("paths.cache_dir", "${DATA_DIR}/cache")
	v.SetDefault("paths.repo_dir", "${DATA_DIR}/repositories")
	v.SetDefault("paths.backup_dir", "${DATA_DIR}/backups")
	v.SetDefault("paths.upload_dir", "${DATA_DIR}/uploads")
	
	// Security defaults
	v.SetDefault("security.enable_2fa", true)
	v.SetDefault("security.enable_webauthn", true)
	v.SetDefault("security.session_timeout", 86400) // 24 hours
	
	// Email defaults
	v.SetDefault("email.smtp_port", 587)
	v.SetDefault("email.enable_tls", true)
	
	// Redis defaults
	v.SetDefault("redis.url", "")
	v.SetDefault("redis.db", 0)
	
	// Webhook defaults
	v.SetDefault("webhook.timeout", 30)
}

// postProcess performs post-processing on configuration
func postProcess(cfg *Config) error {
	// Generate secret key if not set
	if cfg.Security.SecretKey == "" {
		key, err := generateSecretKey()
		if err != nil {
			return fmt.Errorf("failed to generate secret key: %w", err)
		}
		cfg.Security.SecretKey = key
	}
	
	// Select random port if not set
	if cfg.Server.Port == 0 {
		cfg.Server.Port = SelectRandomPort()
	}
	
	// Detect public URL if not set
	if cfg.Server.PublicURL == "" {
		cfg.Server.PublicURL = detectPublicURL(cfg.Server.Port)
	}
	
	// Expand path variables
	cfg.Paths = expandPaths(cfg.Paths)
	cfg.Database.DSN = expandPath(cfg.Database.DSN, cfg.Paths.DataDir)
	
	return nil
}

// SelectRandomPort selects a random port in the range 64000-64999
func SelectRandomPort() int {
	max := big.NewInt(1000)
	n, _ := rand.Int(rand.Reader, max)
	return int(n.Int64()) + 64000
}

// generateSecretKey generates a random secret key
func generateSecretKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// detectPublicURL attempts to detect the public URL
func detectPublicURL(port int) string {
	// First try to get FQDN
	hostname, err := os.Hostname()
	if err == nil && strings.Contains(hostname, ".") {
		return fmt.Sprintf("http://%s:%d", hostname, port)
	}
	
	// Try to get external IP
	if ip := getExternalIP(); ip != "" {
		return fmt.Sprintf("http://%s:%d", ip, port)
	}
	
	// Fall back to hostname
	if hostname != "" {
		return fmt.Sprintf("http://%s:%d", hostname, port)
	}
	
	// Last resort
	return fmt.Sprintf("http://localhost:%d", port)
}

// getExternalIP attempts to get the external IP address
func getExternalIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			
			if ip == nil || ip.IsLoopback() || ip.IsPrivate() {
				continue
			}
			
			if ip.To4() != nil {
				return ip.String()
			}
		}
	}
	
	return ""
}

// expandPaths expands path variables in PathsConfig
func expandPaths(paths PathsConfig) PathsConfig {
	dataDir := paths.DataDir
	paths.LogDir = expandPath(paths.LogDir, dataDir)
	paths.CacheDir = expandPath(paths.CacheDir, dataDir)
	paths.RepoDir = expandPath(paths.RepoDir, dataDir)
	paths.BackupDir = expandPath(paths.BackupDir, dataDir)
	paths.UploadDir = expandPath(paths.UploadDir, dataDir)
	return paths
}

// expandPath expands path variables
func expandPath(path, dataDir string) string {
	path = strings.ReplaceAll(path, "${DATA_DIR}", dataDir)
	path = os.ExpandEnv(path)
	return filepath.Clean(path)
}

// getDefaultDataDir returns the default data directory
func getDefaultDataDir() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("ProgramData"), "casgists")
	case "darwin":
		return "/usr/local/var/casgists"
	default:
		return "/var/lib/casgists"
	}
}