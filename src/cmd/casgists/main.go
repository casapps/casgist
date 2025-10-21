package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	// "github.com/casapps/casgists/src/internal/cli" // Temporarily disabled
	"github.com/casapps/casgists/src/internal/config"
	"github.com/casapps/casgists/src/internal/database"
	"github.com/casapps/casgists/src/internal/privileges"
	"github.com/casapps/casgists/src/internal/server"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

var (
	Version = "dev"
)

func main() {
	// Setup logging
	setupLogging()

	args := os.Args[1:]

	// Handle commands first
	if len(args) > 0 {
		switch args[0] {
		case "install":
			if err := handleInstallCommand(args[1:]); err != nil {
				log.Fatalf("Install failed: %v", err)
			}
			return
		case "uninstall":
			if err := handleUninstallCommand(args[1:]); err != nil {
				log.Fatalf("Uninstall failed: %v", err)
			}
			return
		case "setup":
			if err := handleSetupCommand(args[1:]); err != nil {
				log.Fatalf("Setup failed: %v", err)
			}
			return
		case "verify-install":
			if err := handleVerifyInstallCommand(args[1:]); err != nil {
				log.Fatalf("Verification failed: %v", err)
			}
			return
		case "--version", "-v":
			fmt.Printf("CasGists v%s\n", Version)
			os.Exit(0)
		case "--help", "-h":
			printHelp()
			os.Exit(0)
		}
	}

	// Handle remaining flags
	for _, arg := range args {
		switch arg {
		case "--version", "-v":
			fmt.Printf("CasGists v%s\n", Version)
			os.Exit(0)
		case "--help", "-h":
			printHelp()
			os.Exit(0)
		case "--config-check":
			if err := handleConfigCheckCommand(); err != nil {
				log.Fatalf("Configuration check failed: %v", err)
			}
			return
		case "--dry-run":
			if err := handleDryRunCommand(); err != nil {
				log.Fatalf("Dry run failed: %v", err)
			}
			return
		case "--status":
			if err := handleStatusCommand(); err != nil {
				log.Fatalf("Status check failed: %v", err)
			}
			return
		}
	}

	// Check if this requires privilege escalation
	if privileges.RequiresElevation(args) {
		result := privileges.EscalatePrivileges()
		if !result.Success && !result.AlreadyElevated {
			log.Printf("Warning: Failed to escalate privileges: %v", result.Error)
			log.Println("Running in user mode with limited functionality")
		}
	}

	// Determine privilege level for path configuration
	isPrivileged := privileges.IsElevated()
	
	// Initialize path configuration
	pathConfig := config.NewPathConfig(isPrivileged)
	if err := pathConfig.ResolveAll(); err != nil {
		log.Fatalf("Failed to resolve paths: %v", err)
	}

	// Create necessary directories
	if err := pathConfig.CreateDirectories(); err != nil {
		log.Fatalf("Failed to create directories: %v", err)
	}

	// Validate paths
	if err := pathConfig.ValidatePaths(); err != nil {
		log.Fatalf("Path validation failed: %v", err)
	}

	// Initialize main configuration with resolved paths
	cfg, err := config.LoadWithPaths(pathConfig)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.Initialize(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	
	// Get the underlying SQL DB for closing
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	// Run migrations
	if err := database.MigrateDB(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Create Echo instance
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Note: Static files are served from disk in development
	// In production, they would be embedded in the binary
	
	// Initialize server with path config
	srv := server.NewWithPaths(e, cfg, db, pathConfig)

	// Get configured port - check environment first
	port := cfg.GetInt("server.port")
	if port == 0 {
		// No port specified, use port manager to select one
		portManager := server.NewPortManager(db)
		port, err = portManager.GetConfiguredPort()
		if err != nil {
			log.Fatalf("Failed to get configured port: %v", err)
		}
		// Update config with selected port
		cfg.Set("server.port", port)
	} else {
		// Port specified via environment variable, use it directly
		log.Printf("Using configured port: %d", port)
	}

	log.Printf("CasGists v%s starting on port %d", Version, port)
	
	// Set up graceful shutdown
	go func() {
		if err := srv.Start(context.Background(), fmt.Sprintf(":%d", port)); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}

// selectRandomPort selects a random port in the high range as per SPEC
func selectRandomPort() (int, error) {
	return config.SelectRandomPort()
}

func printHelp() {
	fmt.Printf(`CasGists v%s - Self-hosted Git snippet manager

Usage:
  casgists [options] [command]

Commands:
  install     Install CasGists as a system service
  setup       Run the interactive setup wizard
  
Options:
  -h, --help         Show this help message
  -v, --version      Show version information
  --config-check     Validate configuration file
  --dry-run          Test configuration without starting server
  --status           Show server status

Environment Variables:
  CASGISTS_DATA_DIR      Main data directory (default: /var/lib/casgists)
  CASGISTS_LOG_DIR       Log directory (default: /var/log/casgists)
  CASGISTS_LISTEN_PORT   Server port (default: random 64000-64999)
  CASGISTS_DB_TYPE       Database type: sqlite|postgresql|mysql
  CASGISTS_SECRET_KEY    Server secret key (auto-generated if empty)

Examples:
  casgists                    Start the server
  sudo casgists install       Install as system service
  casgists setup              Run setup wizard
  casgists --config-check     Validate configuration

For more information, visit: https://github.com/casapps/casgists
`, Version)
}

// handleConfigCheckCommand validates configuration without starting server
func handleConfigCheckCommand() error {
	fmt.Println("🔍 Checking CasGists configuration...")
	
	// Determine privilege level for path configuration
	isPrivileged := privileges.IsElevated()
	
	// Initialize path configuration
	pathConfig := config.NewPathConfig(isPrivileged)
	if err := pathConfig.ResolveAll(); err != nil {
		fmt.Printf("❌ Path resolution failed: %v\n", err)
		return err
	}
	fmt.Println("✅ Path configuration valid")

	// Load configuration
	cfg, err := config.LoadWithPaths(pathConfig)
	if err != nil {
		fmt.Printf("❌ Configuration loading failed: %v\n", err)
		return err
	}
	fmt.Println("✅ Configuration loaded successfully")

	// Test database connection  
	if err := testDatabaseConnection(cfg); err != nil {
		fmt.Printf("❌ Database connection failed: %v\n", err)
		return err
	}
	fmt.Println("✅ Database connection successful")

	// Validate required directories
	if err := pathConfig.ValidatePaths(); err != nil {
		fmt.Printf("❌ Path validation failed: %v\n", err)
		return err
	}
	fmt.Println("✅ All paths accessible")

	fmt.Println("\n🎉 Configuration is valid and ready!")
	return nil
}

// handleDryRunCommand tests configuration without starting server
func handleDryRunCommand() error {
	fmt.Println("🧪 Running CasGists in dry-run mode...")
	
	// First run config check
	if err := handleConfigCheckCommand(); err != nil {
		return err
	}

	// Initialize configuration
	isPrivileged := privileges.IsElevated()
	pathConfig := config.NewPathConfig(isPrivileged)
	if err := pathConfig.ResolveAll(); err != nil {
		return err
	}

	cfg, err := config.LoadWithPaths(pathConfig)
	if err != nil {
		return err
	}

	// Initialize database (but don't run migrations)
	db, err := database.Initialize(cfg)
	if err != nil {
		fmt.Printf("❌ Database initialization failed: %v\n", err)
		return err
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	fmt.Println("✅ Database initialized successfully")

	// Test port availability
	port := cfg.GetInt("server.port")
	if port == 0 {
		portManager := server.NewPortManager(db)
		port, err = portManager.GetConfiguredPort()
		if err != nil {
			fmt.Printf("❌ Port selection failed: %v\n", err)
			return err
		}
	}

	if err := testPortAvailability(port); err != nil {
		fmt.Printf("❌ Port %d not available: %v\n", port, err)
		return err
	}
	fmt.Printf("✅ Port %d is available\n", port)

	fmt.Printf("\n🎉 Dry run successful! Server would start on port %d\n", port)
	return nil
}

// handleStatusCommand shows server status
func handleStatusCommand() error {
	fmt.Println("📊 Checking CasGists server status...")
	
	// Try to connect to potential running instances
	portRanges := []int{64000, 64001, 64002, 64003, 64004, 64005}
	var runningPort int
	var serverResponse map[string]interface{}

	for _, port := range portRanges {
		if resp, err := checkServerHealth(port); err == nil {
			runningPort = port
			serverResponse = resp
			break
		}
	}

	if runningPort == 0 {
		fmt.Println("❌ No CasGists server detected")
		fmt.Println("💡 Start server with: casgists")
		return nil
	}

	fmt.Printf("✅ CasGists server running on port %d\n", runningPort)
	
	if serverResponse != nil {
		if version, ok := serverResponse["version"].(string); ok {
			fmt.Printf("📦 Version: %s\n", version)
		}
		if uptime, ok := serverResponse["uptime"].(string); ok {
			fmt.Printf("⏱️  Uptime: %s\n", uptime)
		}
		if status, ok := serverResponse["status"].(string); ok {
			fmt.Printf("🟢 Status: %s\n", status)
		}
		if metrics, ok := serverResponse["metrics"].(map[string]interface{}); ok {
			if users, ok := metrics["total_users"].(float64); ok {
				fmt.Printf("👥 Users: %.0f\n", users)
			}
			if gists, ok := metrics["total_gists"].(float64); ok {
				fmt.Printf("📝 Gists: %.0f\n", gists)
			}
		}
	}

	return nil
}

// Helper functions
func testDatabaseConnection(cfg *viper.Viper) error {
	db, err := database.Initialize(cfg)
	if err != nil {
		return err
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	// Simple ping test
	if sqlDB, err := db.DB(); err == nil {
		return sqlDB.Ping()
	}
	return nil
}

func testPortAvailability(port int) error {
	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

func checkServerHealth(port int) (map[string]interface{}, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/healthz", port))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}
// setupLogging configures pretty console output and server.log file
func setupLogging() {
	// Get log directory from environment or use default
	logDir := os.Getenv("CASGISTS_LOG_DIR")
	if logDir == "" {
		logDir = "/var/log/casgists"
	}

	// Try to create log directory and server.log
	if err := os.MkdirAll(logDir, 0755); err == nil {
		serverLogPath := filepath.Join(logDir, "server.log")
		if logFile, err := os.OpenFile(serverLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			// Setup multi-writer for both console and file
			multiWriter := io.MultiWriter(os.Stdout, logFile)
			log.SetOutput(multiWriter)
			log.Printf("✓ Server logging: %s", serverLogPath)
		}
	}

	// Set log format flags for prettier output
	log.SetFlags(log.Ldate | log.Ltime)
}
