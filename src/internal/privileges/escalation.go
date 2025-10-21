package privileges

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

// EscalationResult represents the result of privilege escalation attempt
type EscalationResult struct {
	Success        bool
	AlreadyElevated bool
	Method         string
	Error          error
}

// RequiresElevation checks if the given command requires privilege escalation
func RequiresElevation(args []string) bool {
	if len(args) == 0 {
		return false
	}

	// Information commands that don't require escalation
	infoCommands := []string{
		"--help", "-h",
		"--version", "-v",
		"--config-check",
		"--validate",
		"--dry-run",
		"--check-deps",
		"--status",
		"version",
		"help",
	}

	for _, arg := range args {
		for _, infoCmd := range infoCommands {
			if arg == infoCmd {
				return false
			}
		}
	}

	// All other commands require privilege escalation for system directory access
	return true
}

// EscalatePrivileges attempts to escalate privileges using platform-appropriate method
func EscalatePrivileges() *EscalationResult {
	result := &EscalationResult{}

	// Check if already running with elevated privileges
	if IsElevated() {
		result.Success = true
		result.AlreadyElevated = true
		result.Method = "already-elevated"
		return result
	}

	switch runtime.GOOS {
	case "linux", "darwin":
		return escalateUnix(result)
	case "windows":
		return escalateWindows(result)
	default:
		result.Error = fmt.Errorf("privilege escalation not supported on %s", runtime.GOOS)
		return result
	}
}

// IsElevated checks if the current process has elevated privileges
func IsElevated() bool {
	switch runtime.GOOS {
	case "linux", "darwin":
		return os.Geteuid() == 0
	case "windows":
		return isWindowsAdmin()
	default:
		return false
	}
}

// escalateUnix handles privilege escalation on Unix-like systems (Linux, macOS)
func escalateUnix(result *EscalationResult) *EscalationResult {
	// Try sudo first
	if _, err := exec.LookPath("sudo"); err == nil {
		result.Method = "sudo"
		result.Success, result.Error = reexecuteWithSudo()
		return result
	}

	// Try pkexec as fallback
	if _, err := exec.LookPath("pkexec"); err == nil {
		result.Method = "pkexec"
		result.Success, result.Error = reexecuteWithPkexec()
		return result
	}

	result.Error = fmt.Errorf("no elevation method available (sudo/pkexec not found)")
	return result
}

// escalateWindows handles privilege escalation on Windows using UAC
func escalateWindows(result *EscalationResult) *EscalationResult {
	result.Method = "uac"
	result.Success, result.Error = reexecuteWithUAC()
	return result
}

// reexecuteWithSudo re-executes the current command with sudo
func reexecuteWithSudo() (bool, error) {
	executable, err := os.Executable()
	if err != nil {
		return false, fmt.Errorf("failed to get executable path: %w", err)
	}

	// Prepare sudo command with original arguments
	args := append([]string{executable}, os.Args[1:]...)
	cmd := exec.Command("sudo", args...)
	
	// Preserve stdio
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute and wait
	err = cmd.Run()
	if err == nil {
		// Successful execution - original process should exit
		os.Exit(0)
	}

	return false, fmt.Errorf("sudo execution failed: %w", err)
}

// reexecuteWithPkexec re-executes the current command with pkexec
func reexecuteWithPkexec() (bool, error) {
	executable, err := os.Executable()
	if err != nil {
		return false, fmt.Errorf("failed to get executable path: %w", err)
	}

	// Prepare pkexec command
	args := append([]string{executable}, os.Args[1:]...)
	cmd := exec.Command("pkexec", args...)
	
	// Preserve stdio
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute and wait
	err = cmd.Run()
	if err == nil {
		// Successful execution - original process should exit
		os.Exit(0)
	}

	return false, fmt.Errorf("pkexec execution failed: %w", err)
}

// reexecuteWithUAC re-executes the current command with Windows UAC
func reexecuteWithUAC() (bool, error) {
	executable, err := os.Executable()
	if err != nil {
		return false, fmt.Errorf("failed to get executable path: %w", err)
	}

	// Prepare arguments string for PowerShell
	args := strings.Join(os.Args[1:], " ")
	psCmd := fmt.Sprintf(`Start-Process -FilePath "%s" -ArgumentList "%s" -Verb RunAs -Wait`, 
		executable, args)

	// Execute via PowerShell with RunAs verb for UAC
	cmd := exec.Command("powershell", "-Command", psCmd)
	err = cmd.Run()
	
	if err == nil {
		// Successful execution - original process should exit
		os.Exit(0)
	}

	return false, fmt.Errorf("UAC elevation failed: %w", err)
}

// isWindowsAdmin checks if running as administrator on Windows
func isWindowsAdmin() bool {
	// This is a simplified implementation
	// In a full implementation, you'd use Windows APIs to check token privileges
	
	// Try to create a file in a restricted location as a test
	testPath := `C:\Windows\Temp\casgists-admin-test`
	err := os.WriteFile(testPath, []byte("test"), 0644)
	if err == nil {
		os.Remove(testPath)
		return true
	}
	
	return false
}

// CreateSystemUser creates a system user for CasGists (Unix only)
func CreateSystemUser() error {
	if runtime.GOOS == "windows" {
		return nil // No user creation needed on Windows
	}

	if !IsElevated() {
		return fmt.Errorf("elevated privileges required to create system user")
	}

	return createUnixSystemUser()
}

// createUnixSystemUser creates a system user on Unix-like systems
func createUnixSystemUser() error {
	// Check if user already exists
	if userExists("casgists") {
		return nil // User already exists
	}

	// Find unique UID/GID in system range
	uid, gid, err := findUniqueUserIDs(850, 999)
	if err != nil {
		return fmt.Errorf("failed to find available UID/GID: %w", err)
	}

	// Create group first
	cmd := exec.Command("groupadd", "-g", strconv.Itoa(gid), "-r", "casgists")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}

	// Create user
	args := []string{
		"-u", strconv.Itoa(uid),
		"-g", strconv.Itoa(gid),
		"-r",                    // System user
		"-s", "/bin/false",      // No shell access
		"-d", "/var/lib/casgists", // Home directory
		"-c", "CasGists System User", // Comment
		"casgists",
	}
	
	cmd = exec.Command("useradd", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// userExists checks if a user exists on the system
func userExists(username string) bool {
	cmd := exec.Command("id", username)
	return cmd.Run() == nil
}

// findUniqueUserIDs finds available UID and GID in the specified range
func findUniqueUserIDs(min, max int) (int, int, error) {
	for attempts := 0; attempts < 50; attempts++ {
		uid := min + (attempts * 3) // Simple progression instead of random
		gid := min + (attempts * 3) + 1
		
		if uid > max || gid > max {
			break
		}

		// Check if UID is available
		cmd := exec.Command("id", strconv.Itoa(uid))
		if cmd.Run() != nil { // User doesn't exist
			// Check if GID is available
			cmd = exec.Command("getent", "group", strconv.Itoa(gid))
			if cmd.Run() != nil { // Group doesn't exist
				return uid, gid, nil
			}
		}
	}
	
	return 0, 0, fmt.Errorf("no available UID/GID in range %d-%d", min, max)
}

// InstallSystemService installs CasGists as a system service
func InstallSystemService() error {
	if !IsElevated() {
		return fmt.Errorf("elevated privileges required to install system service")
	}

	switch runtime.GOOS {
	case "linux":
		return installSystemdService()
	case "darwin":
		return installLaunchdService()
	case "windows":
		return installWindowsService()
	default:
		return fmt.Errorf("system service installation not supported on %s", runtime.GOOS)
	}
}

// installSystemdService installs systemd service on Linux
func installSystemdService() error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	serviceContent := fmt.Sprintf(`[Unit]
Description=CasGists - Self-hosted Git snippet manager
Documentation=https://docs.casgists.com
After=network.target
Wants=network.target

[Service]
Type=simple
User=casgists
Group=casgists
ExecStart=%s server
Restart=always
RestartSec=10
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ReadWritePaths=/var/lib/casgists /var/log/casgists /run/casgists
Environment=CASGISTS_DATA_DIR=/var/lib/casgists
Environment=CASGISTS_LOG_DIR=/var/log/casgists

[Install]
WantedBy=multi-user.target
`, executable)

	servicePath := "/etc/systemd/system/casgists.service"
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write systemd service file: %w", err)
	}

	// Reload systemd and enable service
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	if err := exec.Command("systemctl", "enable", "casgists").Run(); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}

	return nil
}

// installLaunchdService installs launchd service on macOS
func installLaunchdService() error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.casgists.server</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
		<string>server</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
	<key>StandardErrorPath</key>
	<string>/usr/local/var/log/casgists/casgists.log</string>
	<key>StandardOutPath</key>
	<string>/usr/local/var/log/casgists/casgists.log</string>
	<key>WorkingDirectory</key>
	<string>/usr/local/var/lib/casgists</string>
</dict>
</plist>
`, executable)

	plistPath := "/Library/LaunchDaemons/com.casgists.server.plist"
	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		return fmt.Errorf("failed to write launchd plist: %w", err)
	}

	// Load the service
	if err := exec.Command("launchctl", "load", plistPath).Run(); err != nil {
		return fmt.Errorf("failed to load launchd service: %w", err)
	}

	return nil
}

// installWindowsService installs Windows service (stub implementation)
func installWindowsService() error {
	// TODO: Implement Windows service installation
	// This would typically use the Windows Service Manager APIs
	return fmt.Errorf("Windows service installation not yet implemented")
}

// DropPrivileges drops elevated privileges after setup is complete
func DropPrivileges(username string) error {
	if runtime.GOOS == "windows" {
		return nil // Privilege dropping works differently on Windows
	}

	if !IsElevated() {
		return nil // Already running without elevated privileges
	}

	// Look up the user
	cmd := exec.Command("id", "-u", username)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get UID for user %s: %w", username, err)
	}

	uid, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return fmt.Errorf("failed to parse UID: %w", err)
	}

	cmd = exec.Command("id", "-g", username)
	output, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get GID for user %s: %w", username, err)
	}

	gid, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return fmt.Errorf("failed to parse GID: %w", err)
	}

	// Drop privileges
	if err := syscall.Setgid(gid); err != nil {
		return fmt.Errorf("failed to set GID: %w", err)
	}

	if err := syscall.Setuid(uid); err != nil {
		return fmt.Errorf("failed to set UID: %w", err)
	}

	return nil
}