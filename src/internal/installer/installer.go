package installer

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"text/template"
	"time"
)

// Installer handles the installation of CasGists as a system service
type Installer struct {
	Config InstallerConfig
	writer io.Writer
}

// InstallerConfig holds installation configuration
type InstallerConfig struct {
	ServiceName     string
	BinaryPath      string
	DataDir         string
	ConfigPath      string
	User            string
	Group           string
	Port            int
	InstallPath     string
	WorkingDir      string
	EnvFile         string
	LogFile         string
	Description     string
	NoSystemService bool
}

// NewInstaller creates a new installer instance
func NewInstaller(config InstallerConfig) *Installer {
	if config.ServiceName == "" {
		config.ServiceName = "casgists"
	}
	if config.User == "" {
		config.User = "casgists"
	}
	if config.Group == "" {
		config.Group = "casgists"
	}
	if config.Port == 0 {
		config.Port = 64080
	}
	if config.InstallPath == "" {
		config.InstallPath = "/opt/casgists"
	}
	if config.DataDir == "" {
		config.DataDir = "/var/lib/casgists"
	}
	if config.ConfigPath == "" {
		config.ConfigPath = "/etc/casgists/config.yaml"
	}
	if config.WorkingDir == "" {
		config.WorkingDir = "/var/lib/casgists"
	}
	if config.EnvFile == "" {
		config.EnvFile = "/etc/casgists/environment"
	}
	if config.LogFile == "" {
		config.LogFile = "/var/log/casgists/casgists.log"
	}
	if config.Description == "" {
		config.Description = "CasGists - Self-hosted GitHub Gist Alternative"
	}

	return &Installer{
		Config: config,
		writer: os.Stdout,
	}
}

// Install performs the system installation
func (i *Installer) Install(ctx context.Context) error {
	// Check if running as root
	if os.Geteuid() != 0 {
		return fmt.Errorf("installation must be run as root (use sudo)")
	}

	// Detect OS and choose installation method
	switch runtime.GOOS {
	case "linux":
		return i.installLinux(ctx)
	case "darwin":
		return i.installDarwin(ctx)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// installLinux handles Linux system installation
func (i *Installer) installLinux(ctx context.Context) error {
	fmt.Fprintln(i.writer, "Installing CasGists for Linux...")

	// Create system user
	if err := i.createSystemUser(); err != nil {
		return fmt.Errorf("failed to create system user: %w", err)
	}

	// Create directory structure
	if err := i.createDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Copy binary to install location
	if err := i.installBinary(); err != nil {
		return fmt.Errorf("failed to install binary: %w", err)
	}

	// Setup systemd service
	serviceType := detectInitSystem()
	switch serviceType {
	case "systemd":
		if err := i.installSystemdService(); err != nil {
			return fmt.Errorf("failed to install systemd service: %w", err)
		}
	case "sysvinit":
		if err := i.installSysVInitService(); err != nil {
			return fmt.Errorf("failed to install SysV init service: %w", err)
		}
	default:
		fmt.Fprintln(i.writer, "Warning: Unknown init system. Skipping service installation.")
		fmt.Fprintln(i.writer, "You'll need to manually configure CasGists to start at boot.")
	}

	// Set up port capability (for binding to privileged ports without root)
	if err := i.setupPortCapability(); err != nil {
		fmt.Fprintf(i.writer, "Warning: Failed to set port capability: %v\n", err)
		fmt.Fprintln(i.writer, "CasGists will need to run as root to bind to ports below 1024")
	}

	// Create initial configuration
	if err := i.createInitialConfig(); err != nil {
		return fmt.Errorf("failed to create initial configuration: %w", err)
	}

	// Set proper permissions
	if err := i.setPermissions(); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	fmt.Fprintln(i.writer, "\nInstallation completed successfully!")
	i.printPostInstallInstructions()

	return nil
}

// installDarwin handles macOS system installation
func (i *Installer) installDarwin(ctx context.Context) error {
	fmt.Fprintln(i.writer, "Installing CasGists for macOS...")

	// Create directory structure
	if err := i.createDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Copy binary to install location
	if err := i.installBinary(); err != nil {
		return fmt.Errorf("failed to install binary: %w", err)
	}

	// Install launchd service
	if err := i.installLaunchdService(); err != nil {
		return fmt.Errorf("failed to install launchd service: %w", err)
	}

	// Create initial configuration
	if err := i.createInitialConfig(); err != nil {
		return fmt.Errorf("failed to create initial configuration: %w", err)
	}

	// Set proper permissions
	if err := i.setPermissions(); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	fmt.Fprintln(i.writer, "\nInstallation completed successfully!")
	i.printPostInstallInstructions()

	return nil
}

// createSystemUser creates the CasGists system user
func (i *Installer) createSystemUser() error {
	// Check if user already exists
	if _, err := user.Lookup(i.Config.User); err == nil {
		fmt.Fprintf(i.writer, "User '%s' already exists, skipping creation\n", i.Config.User)
		return nil
	}

	fmt.Fprintf(i.writer, "Creating system user '%s'...\n", i.Config.User)

	// Create system user with no login shell
	cmd := exec.Command("useradd",
		"--system",
		"--shell", "/bin/false",
		"--home", i.Config.DataDir,
		"--user-group",
		"--comment", "CasGists Service User",
		i.Config.User,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// createDirectories creates the required directory structure
func (i *Installer) createDirectories() error {
	dirs := []string{
		i.Config.InstallPath,
		i.Config.DataDir,
		filepath.Dir(i.Config.ConfigPath),
		filepath.Dir(i.Config.LogFile),
		filepath.Join(i.Config.DataDir, "gists"),
		filepath.Join(i.Config.DataDir, "repos"),
		filepath.Join(i.Config.DataDir, "cache"),
		filepath.Join(i.Config.DataDir, "uploads"),
		filepath.Join(i.Config.DataDir, "backups"),
		filepath.Join(i.Config.DataDir, "gdpr_exports"),
	}

	for _, dir := range dirs {
		fmt.Fprintf(i.writer, "Creating directory: %s\n", dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// installBinary copies the CasGists binary to the installation location
func (i *Installer) installBinary() error {
	targetPath := filepath.Join(i.Config.InstallPath, "bin", "casgists")

	// Create bin directory
	binDir := filepath.Join(i.Config.InstallPath, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	fmt.Fprintf(i.writer, "Installing binary to %s...\n", targetPath)

	// Copy binary
	source, err := os.Open(i.Config.BinaryPath)
	if err != nil {
		return fmt.Errorf("failed to open source binary: %w", err)
	}
	defer source.Close()

	target, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to create target binary: %w", err)
	}
	defer target.Close()

	if _, err := io.Copy(target, source); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	// Create symlink in /usr/local/bin for CLI access
	symlinkPath := "/usr/local/bin/casgists"
	os.Remove(symlinkPath) // Remove existing symlink if any
	if err := os.Symlink(targetPath, symlinkPath); err != nil {
		fmt.Fprintf(i.writer, "Warning: Failed to create symlink at %s: %v\n", symlinkPath, err)
	}

	return nil
}

// setupPortCapability sets CAP_NET_BIND_SERVICE capability for binding to privileged ports
func (i *Installer) setupPortCapability() error {
	if runtime.GOOS != "linux" {
		return nil
	}

	binaryPath := filepath.Join(i.Config.InstallPath, "bin", "casgists")

	// Check if setcap is available
	if _, err := exec.LookPath("setcap"); err != nil {
		return fmt.Errorf("setcap not found: %w", err)
	}

	fmt.Fprintln(i.writer, "Setting port binding capability...")

	// Grant capability to bind to privileged ports
	cmd := exec.Command("setcap", "cap_net_bind_service=+ep", binaryPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set capability: %w", err)
	}

	return nil
}

// installSystemdService creates and installs the systemd service file
func (i *Installer) installSystemdService() error {
	fmt.Fprintln(i.writer, "Installing systemd service...")

	serviceContent := `[Unit]
Description={{.Description}}
After=network.target

[Service]
Type=notify
User={{.User}}
Group={{.Group}}
WorkingDirectory={{.WorkingDir}}
ExecStart={{.InstallPath}}/bin/casgists serve --config {{.ConfigPath}}
Restart=on-failure
RestartSec=5s

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths={{.DataDir}} {{.LogDir}}
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true
RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX
RestrictNamespaces=true
LockPersonality=true
MemoryDenyWriteExecute=true
RestrictRealtime=true
RestrictSUIDSGID=true
RemoveIPC=true

# Environment
Environment="HOME={{.DataDir}}"
Environment="USER={{.User}}"
EnvironmentFile=-{{.EnvFile}}

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier={{.ServiceName}}

[Install]
WantedBy=multi-user.target
`

	// Parse and execute template
	tmpl, err := template.New("systemd").Parse(serviceContent)
	if err != nil {
		return fmt.Errorf("failed to parse systemd template: %w", err)
	}

	var buf bytes.Buffer
	data := map[string]string{
		"Description":  i.Config.Description,
		"User":         i.Config.User,
		"Group":        i.Config.Group,
		"WorkingDir":   i.Config.WorkingDir,
		"InstallPath":  i.Config.InstallPath,
		"ConfigPath":   i.Config.ConfigPath,
		"DataDir":      i.Config.DataDir,
		"LogDir":       filepath.Dir(i.Config.LogFile),
		"EnvFile":      i.Config.EnvFile,
		"ServiceName":  i.Config.ServiceName,
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute systemd template: %w", err)
	}

	// Write service file
	servicePath := fmt.Sprintf("/etc/systemd/system/%s.service", i.Config.ServiceName)
	if err := os.WriteFile(servicePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	// Reload systemd
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	// Enable service
	if err := exec.Command("systemctl", "enable", i.Config.ServiceName).Run(); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}

	return nil
}

// installSysVInitService creates and installs SysV init script
func (i *Installer) installSysVInitService() error {
	fmt.Fprintln(i.writer, "Installing SysV init service...")

	scriptContent := `#!/bin/sh
### BEGIN INIT INFO
# Provides:          {{.ServiceName}}
# Required-Start:    $local_fs $network $named $time $syslog
# Required-Stop:     $local_fs $network $named $time $syslog
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Description:       {{.Description}}
### END INIT INFO

NAME={{.ServiceName}}
DAEMON={{.InstallPath}}/bin/casgists
DAEMON_ARGS="serve --config {{.ConfigPath}}"
PIDFILE=/var/run/$NAME.pid
USER={{.User}}
GROUP={{.Group}}

. /lib/init/vars.sh
. /lib/lsb/init-functions

do_start() {
    start-stop-daemon --start --quiet --pidfile $PIDFILE --exec $DAEMON \
        --chuid $USER:$GROUP --make-pidfile --background -- $DAEMON_ARGS
}

do_stop() {
    start-stop-daemon --stop --quiet --retry=TERM/30/KILL/5 --pidfile $PIDFILE
    rm -f $PIDFILE
}

case "$1" in
  start)
    log_daemon_msg "Starting $NAME"
    do_start
    log_end_msg $?
    ;;
  stop)
    log_daemon_msg "Stopping $NAME"
    do_stop
    log_end_msg $?
    ;;
  restart)
    log_daemon_msg "Restarting $NAME"
    do_stop
    do_start
    log_end_msg $?
    ;;
  status)
    status_of_proc -p $PIDFILE "$DAEMON" "$NAME" && exit 0 || exit $?
    ;;
  *)
    echo "Usage: /etc/init.d/$NAME {start|stop|restart|status}"
    exit 1
    ;;
esac

exit 0
`

	// Parse and execute template
	tmpl, err := template.New("sysvinit").Parse(scriptContent)
	if err != nil {
		return fmt.Errorf("failed to parse sysvinit template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, i.Config); err != nil {
		return fmt.Errorf("failed to execute sysvinit template: %w", err)
	}

	// Write init script
	scriptPath := fmt.Sprintf("/etc/init.d/%s", i.Config.ServiceName)
	if err := os.WriteFile(scriptPath, buf.Bytes(), 0755); err != nil {
		return fmt.Errorf("failed to write init script: %w", err)
	}

	// Update rc.d
	if err := exec.Command("update-rc.d", i.Config.ServiceName, "defaults").Run(); err != nil {
		return fmt.Errorf("failed to update rc.d: %w", err)
	}

	return nil
}

// installLaunchdService creates and installs launchd service for macOS
func (i *Installer) installLaunchdService() error {
	fmt.Fprintln(i.writer, "Installing launchd service...")

	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.casapps.{{.ServiceName}}</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{.InstallPath}}/bin/casgists</string>
        <string>serve</string>
        <string>--config</string>
        <string>{{.ConfigPath}}</string>
    </array>
    <key>WorkingDirectory</key>
    <string>{{.WorkingDir}}</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>HOME</key>
        <string>{{.DataDir}}</string>
        <key>USER</key>
        <string>{{.User}}</string>
    </dict>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>{{.LogFile}}</string>
    <key>StandardErrorPath</key>
    <string>{{.LogFile}}.error</string>
    <key>UserName</key>
    <string>{{.User}}</string>
</dict>
</plist>
`

	// Parse and execute template
	tmpl, err := template.New("launchd").Parse(plistContent)
	if err != nil {
		return fmt.Errorf("failed to parse launchd template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, i.Config); err != nil {
		return fmt.Errorf("failed to execute launchd template: %w", err)
	}

	// Write plist file
	plistPath := fmt.Sprintf("/Library/LaunchDaemons/com.casapps.%s.plist", i.Config.ServiceName)
	if err := os.WriteFile(plistPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write plist file: %w", err)
	}

	// Load service
	if err := exec.Command("launchctl", "load", plistPath).Run(); err != nil {
		return fmt.Errorf("failed to load launchd service: %w", err)
	}

	return nil
}

// createInitialConfig creates the initial configuration file
func (i *Installer) createInitialConfig() error {
	fmt.Fprintln(i.writer, "Creating initial configuration...")

	configContent := fmt.Sprintf(`# CasGists Configuration
# Generated by installer at %s

server:
  host: 0.0.0.0
  port: %d
  base_url: http://localhost:%d

database:
  type: sqlite
  path: ${DATA_DIR}/casgists.db

paths:
  data_dir: %s
  repo_dir: ${DATA_DIR}/repos
  cache_dir: ${DATA_DIR}/cache
  upload_dir: ${DATA_DIR}/uploads
  backup_dir: ${DATA_DIR}/backups
  gdpr_exports: ${DATA_DIR}/gdpr_exports

security:
  secret_key: %s
  
logging:
  level: info
  file: %s

# Default limits
limits:
  max_gist_size: 10485760  # 10MB
  max_file_size: 1048576   # 1MB per file
  max_files_per_gist: 100

# Features
features:
  registration: true
  anonymous_gists: true
  webhooks: true
  email_notifications: true
`, time.Now().Format(time.RFC3339), i.Config.Port, i.Config.Port, i.Config.DataDir, generateSecretKey(), i.Config.LogFile)

	// Ensure config directory exists
	configDir := filepath.Dir(i.Config.ConfigPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config file
	if err := os.WriteFile(i.Config.ConfigPath, []byte(configContent), 0640); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Create environment file
	envContent := fmt.Sprintf(`# CasGists Environment Variables
# Generated by installer

DATA_DIR=%s
CONFIG_PATH=%s
LOG_FILE=%s
`, i.Config.DataDir, i.Config.ConfigPath, i.Config.LogFile)

	if err := os.WriteFile(i.Config.EnvFile, []byte(envContent), 0640); err != nil {
		fmt.Fprintf(i.writer, "Warning: Failed to create environment file: %v\n", err)
	}

	return nil
}

// setPermissions sets proper file permissions
func (i *Installer) setPermissions() error {
	fmt.Fprintln(i.writer, "Setting file permissions...")

	// Only chown on Linux (macOS handles differently)
	if runtime.GOOS == "linux" {
		// Get user and group IDs
		u, err := user.Lookup(i.Config.User)
		if err != nil {
			return fmt.Errorf("failed to lookup user: %w", err)
		}

		uid, err := strconv.Atoi(u.Uid)
		if err != nil {
			return fmt.Errorf("failed to parse UID: %w", err)
		}

		gid, err := strconv.Atoi(u.Gid)
		if err != nil {
			return fmt.Errorf("failed to parse GID: %w", err)
		}

		// Set ownership
		paths := []string{
			i.Config.DataDir,
			i.Config.ConfigPath,
			i.Config.EnvFile,
			filepath.Dir(i.Config.LogFile),
		}

		for _, path := range paths {
			if err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				return os.Chown(p, uid, gid)
			}); err != nil {
				fmt.Fprintf(i.writer, "Warning: Failed to set ownership on %s: %v\n", path, err)
			}
		}

		// Set config file permissions (readable by group)
		if err := os.Chmod(i.Config.ConfigPath, 0640); err != nil {
			fmt.Fprintf(i.writer, "Warning: Failed to set config permissions: %v\n", err)
		}
	}

	return nil
}

// Uninstall removes CasGists from the system
func (i *Installer) Uninstall(ctx context.Context) error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("uninstallation must be run as root (use sudo)")
	}

	fmt.Fprintln(i.writer, "Uninstalling CasGists...")

	// Stop and disable service
	switch runtime.GOOS {
	case "linux":
		serviceType := detectInitSystem()
		switch serviceType {
		case "systemd":
			exec.Command("systemctl", "stop", i.Config.ServiceName).Run()
			exec.Command("systemctl", "disable", i.Config.ServiceName).Run()
			os.Remove(fmt.Sprintf("/etc/systemd/system/%s.service", i.Config.ServiceName))
			exec.Command("systemctl", "daemon-reload").Run()
		case "sysvinit":
			exec.Command("service", i.Config.ServiceName, "stop").Run()
			exec.Command("update-rc.d", "-f", i.Config.ServiceName, "remove").Run()
			os.Remove(fmt.Sprintf("/etc/init.d/%s", i.Config.ServiceName))
		}
	case "darwin":
		plistPath := fmt.Sprintf("/Library/LaunchDaemons/com.casapps.%s.plist", i.Config.ServiceName)
		exec.Command("launchctl", "unload", plistPath).Run()
		os.Remove(plistPath)
	}

	// Remove symlink
	os.Remove("/usr/local/bin/casgists")

	// Ask about data removal
	fmt.Fprintln(i.writer, "\nDo you want to remove all CasGists data? (This cannot be undone)")
	fmt.Fprint(i.writer, "Remove data? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))

	if response == "y" || response == "yes" {
		fmt.Fprintln(i.writer, "Removing all data...")
		os.RemoveAll(i.Config.DataDir)
		os.RemoveAll(i.Config.InstallPath)
		os.RemoveAll(filepath.Dir(i.Config.ConfigPath))
		os.RemoveAll(filepath.Dir(i.Config.LogFile))
	} else {
		fmt.Fprintln(i.writer, "Data preserved. Only removing binaries...")
		os.RemoveAll(i.Config.InstallPath)
	}

	// Remove user (Linux only)
	if runtime.GOOS == "linux" && response == "y" {
		exec.Command("userdel", i.Config.User).Run()
		exec.Command("groupdel", i.Config.Group).Run()
	}

	fmt.Fprintln(i.writer, "\nUninstallation completed.")
	return nil
}

// printPostInstallInstructions prints post-installation instructions
func (i *Installer) printPostInstallInstructions() {
	fmt.Fprintln(i.writer, "\n"+strings.Repeat("=", 60))
	fmt.Fprintln(i.writer, "CasGists Installation Complete!")
	fmt.Fprintln(i.writer, strings.Repeat("=", 60))
	fmt.Fprintln(i.writer, "\nNext steps:")
	fmt.Fprintln(i.writer, "1. Review and edit the configuration file:")
	fmt.Fprintf(i.writer, "   sudo nano %s\n", i.Config.ConfigPath)
	fmt.Fprintln(i.writer, "\n2. Start the CasGists service:")

	switch runtime.GOOS {
	case "linux":
		if detectInitSystem() == "systemd" {
			fmt.Fprintf(i.writer, "   sudo systemctl start %s\n", i.Config.ServiceName)
			fmt.Fprintf(i.writer, "   sudo systemctl status %s\n", i.Config.ServiceName)
		} else {
			fmt.Fprintf(i.writer, "   sudo service %s start\n", i.Config.ServiceName)
			fmt.Fprintf(i.writer, "   sudo service %s status\n", i.Config.ServiceName)
		}
	case "darwin":
		fmt.Fprintf(i.writer, "   sudo launchctl start com.casapps.%s\n", i.Config.ServiceName)
	}

	fmt.Fprintln(i.writer, "\n3. Access CasGists at:")
	fmt.Fprintf(i.writer, "   http://localhost:%d\n", i.Config.Port)
	fmt.Fprintln(i.writer, "\n4. Complete the setup wizard on first access")
	fmt.Fprintln(i.writer, "\nView logs:")
	fmt.Fprintf(i.writer, "   tail -f %s\n", i.Config.LogFile)

	if runtime.GOOS == "linux" && detectInitSystem() == "systemd" {
		fmt.Fprintf(i.writer, "   journalctl -u %s -f\n", i.Config.ServiceName)
	}

	fmt.Fprintln(i.writer, "\n"+strings.Repeat("=", 60))
}

// detectInitSystem detects the init system on Linux
func detectInitSystem() string {
	// Check for systemd
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		return "systemd"
	}

	// Check for sysvinit
	if _, err := os.Stat("/etc/init.d"); err == nil {
		return "sysvinit"
	}

	return "unknown"
}

// generateSecretKey generates a random secret key
func generateSecretKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 64)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

// CheckPort checks if a port is available
func CheckPort(port int) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("port %d is not available: %w", port, err)
	}
	ln.Close()
	return nil
}

// VerifySystemRequirements checks system requirements
func VerifySystemRequirements() error {
	// Check OS
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	// Check architecture
	if runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64" {
		return fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}

	// Check available disk space (require at least 1GB)
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err == nil {
		availableGB := stat.Bavail * uint64(stat.Bsize) / (1024 * 1024 * 1024)
		if availableGB < 1 {
			return fmt.Errorf("insufficient disk space: %d GB available (minimum 1 GB required)", availableGB)
		}
	}

	return nil
}