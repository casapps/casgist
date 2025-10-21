package privilege

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

// EscalationService handles privilege escalation operations
type EscalationService struct {
	cfg *viper.Viper
}

// EscalationMethod represents different ways to escalate privileges
type EscalationMethod string

const (
	EscalationMethodSudo   EscalationMethod = "sudo"
	EscalationMethodPkexec EscalationMethod = "pkexec"
	EscalationMethodRunas  EscalationMethod = "runas" // Windows
)

// SystemOperation represents operations that require privilege escalation
type SystemOperation string

const (
	OperationInstallService   SystemOperation = "install_service"
	OperationUninstallService SystemOperation = "uninstall_service"
	OperationStartService     SystemOperation = "start_service"
	OperationStopService      SystemOperation = "stop_service"
	OperationCreateUser       SystemOperation = "create_user"
	OperationSetPermissions   SystemOperation = "set_permissions"
	OperationBindPrivilegedPort SystemOperation = "bind_privileged_port"
	OperationWriteSystemConfig  SystemOperation = "write_system_config"
)

// PrivilegeRequest represents a privilege escalation request
type PrivilegeRequest struct {
	Operation   SystemOperation `json:"operation"`
	Command     string         `json:"command"`
	Arguments   []string       `json:"arguments"`
	WorkingDir  string         `json:"working_dir"`
	Method      EscalationMethod `json:"method"`
	Reason      string         `json:"reason"`
	Interactive bool           `json:"interactive"`
}

// PrivilegeResult represents the result of privilege escalation
type PrivilegeResult struct {
	Success    bool   `json:"success"`
	Output     string `json:"output"`
	Error      string `json:"error"`
	ExitCode   int    `json:"exit_code"`
	Method     EscalationMethod `json:"method_used"`
}

// NewEscalationService creates a new privilege escalation service
func NewEscalationService(cfg *viper.Viper) *EscalationService {
	return &EscalationService{
		cfg: cfg,
	}
}

// IsPrivileged checks if the current process is running with elevated privileges
func (s *EscalationService) IsPrivileged() bool {
	switch runtime.GOOS {
	case "windows":
		return s.isWindowsAdmin()
	case "darwin", "linux":
		return os.Geteuid() == 0
	default:
		return false
	}
}

// CanEscalate checks if privilege escalation is available
func (s *EscalationService) CanEscalate() (bool, []EscalationMethod) {
	var methods []EscalationMethod

	switch runtime.GOOS {
	case "windows":
		// Check if UAC can be used
		if s.isUACAvailable() {
			methods = append(methods, EscalationMethodRunas)
		}
	case "darwin", "linux":
		// Check for sudo
		if s.isCommandAvailable("sudo") {
			methods = append(methods, EscalationMethodSudo)
		}
		// Check for pkexec (PolicyKit)
		if s.isCommandAvailable("pkexec") {
			methods = append(methods, EscalationMethodPkexec)
		}
	}

	return len(methods) > 0, methods
}

// ExecutePrivileged executes a command with elevated privileges
func (s *EscalationService) ExecutePrivileged(req PrivilegeRequest) (*PrivilegeResult, error) {
	if s.IsPrivileged() {
		// Already running as admin/root, execute directly
		return s.executeDirect(req)
	}

	// Check if escalation is available
	canEscalate, methods := s.CanEscalate()
	if !canEscalate {
		return &PrivilegeResult{
			Success:  false,
			Error:    "Privilege escalation not available on this system",
			ExitCode: -1,
		}, nil
	}

	// Use specified method or choose automatically
	method := req.Method
	if method == "" {
		method = methods[0] // Use first available method
	}

	// Validate method is available
	methodAvailable := false
	for _, m := range methods {
		if m == method {
			methodAvailable = true
			break
		}
	}
	if !methodAvailable {
		return &PrivilegeResult{
			Success:  false,
			Error:    fmt.Sprintf("Escalation method '%s' not available", method),
			ExitCode: -1,
		}, nil
	}

	// Execute with escalation
	switch method {
	case EscalationMethodSudo:
		return s.executeWithSudo(req)
	case EscalationMethodPkexec:
		return s.executeWithPkexec(req)
	case EscalationMethodRunas:
		return s.executeWithRunas(req)
	default:
		return &PrivilegeResult{
			Success:  false,
			Error:    fmt.Sprintf("Unknown escalation method: %s", method),
			ExitCode: -1,
		}, nil
	}
}

// executeDirect executes command directly (already privileged)
func (s *EscalationService) executeDirect(req PrivilegeRequest) (*PrivilegeResult, error) {
	cmd := exec.Command(req.Command, req.Arguments...)
	if req.WorkingDir != "" {
		cmd.Dir = req.WorkingDir
	}

	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return &PrivilegeResult{
		Success:  err == nil,
		Output:   string(output),
		Error:    getErrorString(err),
		ExitCode: exitCode,
		Method:   "", // No escalation used
	}, nil
}

// executeWithSudo executes command using sudo
func (s *EscalationService) executeWithSudo(req PrivilegeRequest) (*PrivilegeResult, error) {
	args := []string{}
	
	// Add sudo options
	if !req.Interactive {
		args = append(args, "-n") // Non-interactive
	}
	
	// Add reason if provided
	if req.Reason != "" {
		args = append(args, "-p", fmt.Sprintf("CasGists needs sudo access for: %s [sudo] password for %%p: ", req.Reason))
	}

	// Add command and arguments
	args = append(args, req.Command)
	args = append(args, req.Arguments...)

	cmd := exec.Command("sudo", args...)
	if req.WorkingDir != "" {
		cmd.Dir = req.WorkingDir
	}

	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return &PrivilegeResult{
		Success:  err == nil,
		Output:   string(output),
		Error:    getErrorString(err),
		ExitCode: exitCode,
		Method:   EscalationMethodSudo,
	}, nil
}

// executeWithPkexec executes command using pkexec
func (s *EscalationService) executeWithPkexec(req PrivilegeRequest) (*PrivilegeResult, error) {
	args := []string{}
	
	// Add pkexec options
	if req.Reason != "" {
		args = append(args, "--message", fmt.Sprintf("CasGists needs administrator access for: %s", req.Reason))
	}

	// Add command and arguments
	args = append(args, req.Command)
	args = append(args, req.Arguments...)

	cmd := exec.Command("pkexec", args...)
	if req.WorkingDir != "" {
		cmd.Dir = req.WorkingDir
	}

	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return &PrivilegeResult{
		Success:  err == nil,
		Output:   string(output),
		Error:    getErrorString(err),
		ExitCode: exitCode,
		Method:   EscalationMethodPkexec,
	}, nil
}

// executeWithRunas executes command using Windows runas
func (s *EscalationService) executeWithRunas(req PrivilegeRequest) (*PrivilegeResult, error) {
	if runtime.GOOS != "windows" {
		return &PrivilegeResult{
			Success:  false,
			Error:    "runas is only available on Windows",
			ExitCode: -1,
		}, nil
	}

	// Use PowerShell Start-Process with -Verb RunAs
	psCommand := fmt.Sprintf("Start-Process -FilePath '%s' -ArgumentList '%s' -Verb RunAs -Wait -PassThru", 
		req.Command, strings.Join(req.Arguments, "', '"))
	
	cmd := exec.Command("powershell", "-Command", psCommand)
	if req.WorkingDir != "" {
		cmd.Dir = req.WorkingDir
	}

	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return &PrivilegeResult{
		Success:  err == nil,
		Output:   string(output),
		Error:    getErrorString(err),
		ExitCode: exitCode,
		Method:   EscalationMethodRunas,
	}, nil
}

// isWindowsAdmin checks if running as Windows administrator
func (s *EscalationService) isWindowsAdmin() bool {
	if runtime.GOOS != "windows" {
		return false
	}

	cmd := exec.Command("net", "session")
	err := cmd.Run()
	return err == nil
}

// isUACAvailable checks if UAC is available on Windows
func (s *EscalationService) isUACAvailable() bool {
	if runtime.GOOS != "windows" {
		return false
	}

	// Check if UAC is enabled
	cmd := exec.Command("reg", "query", "HKEY_LOCAL_MACHINE\\SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Policies\\System", "/v", "EnableLUA")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.Contains(string(output), "0x1")
}

// isCommandAvailable checks if a command is available in PATH
func (s *EscalationService) isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// getErrorString safely converts error to string
func getErrorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// GetCurrentUser returns information about current user privileges
func (s *EscalationService) GetCurrentUser() map[string]interface{} {
	info := map[string]interface{}{
		"uid":         os.Getuid(),
		"gid":         os.Getgid(),
		"privileged":  s.IsPrivileged(),
	}

	if runtime.GOOS != "windows" {
		info["euid"] = os.Geteuid()
		info["egid"] = os.Getegid()
		
		// Get user groups
		if groups, err := os.Getgroups(); err == nil {
			info["groups"] = groups
		}
	}

	return info
}

// RequiresPrivileges checks if an operation requires privilege escalation
func (s *EscalationService) RequiresPrivileges(operation SystemOperation) bool {
	if s.IsPrivileged() {
		return false // Already privileged
	}

	switch operation {
	case OperationInstallService, OperationUninstallService, 
		 OperationStartService, OperationStopService,
		 OperationCreateUser, OperationSetPermissions,
		 OperationBindPrivilegedPort, OperationWriteSystemConfig:
		return true
	default:
		return false
	}
}

// CreateServiceInstallRequest creates a privilege request for service installation
func (s *EscalationService) CreateServiceInstallRequest(serviceName, binaryPath, description string) PrivilegeRequest {
	switch runtime.GOOS {
	case "linux":
		// Create systemd service file
		return PrivilegeRequest{
			Operation:  OperationInstallService,
			Command:    "systemctl",
			Arguments:  []string{"enable", serviceName},
			Reason:     "Installing CasGists system service",
			Interactive: false,
		}
	case "darwin":
		// Create launchd plist
		return PrivilegeRequest{
			Operation:  OperationInstallService,
			Command:    "launchctl",
			Arguments:  []string{"load", "-w", "/Library/LaunchDaemons/com.casapps.casgists.plist"},
			Reason:     "Installing CasGists system service",
			Interactive: false,
		}
	case "windows":
		// Install Windows service
		return PrivilegeRequest{
			Operation:  OperationInstallService,
			Command:    "sc",
			Arguments:  []string{"create", serviceName, "binPath=", binaryPath, "DisplayName=", description},
			Reason:     "Installing CasGists Windows service",
			Interactive: false,
		}
	default:
		return PrivilegeRequest{
			Operation: OperationInstallService,
			Reason:    "Service installation not supported on this platform",
		}
	}
}

// CreatePortBindRequest creates a privilege request for binding to privileged ports
func (s *EscalationService) CreatePortBindRequest(port int) PrivilegeRequest {
	if port >= 1024 {
		// Non-privileged port, no escalation needed
		return PrivilegeRequest{
			Operation: OperationBindPrivilegedPort,
			Reason:    "Port binding does not require privileges",
		}
	}

	switch runtime.GOOS {
	case "linux":
		// Use setcap to allow binding to privileged ports
		return PrivilegeRequest{
			Operation:   OperationBindPrivilegedPort,
			Command:     "setcap",
			Arguments:   []string{"cap_net_bind_service=+ep", os.Args[0]},
			Reason:      fmt.Sprintf("Allowing CasGists to bind to port %d", port),
			Interactive: false,
		}
	default:
		return PrivilegeRequest{
			Operation: OperationBindPrivilegedPort,
			Reason:    fmt.Sprintf("Binding to privileged port %d", port),
		}
	}
}

// TestPrivilegeEscalation tests if privilege escalation is working
func (s *EscalationService) TestPrivilegeEscalation() (*PrivilegeResult, error) {
	if s.IsPrivileged() {
		return &PrivilegeResult{
			Success: true,
			Output:  "Already running with privileges",
			Method:  "",
		}, nil
	}

	// Test with a simple command that requires no actual privileges
	req := PrivilegeRequest{
		Operation:   "test",
		Command:     "echo",
		Arguments:   []string{"privilege escalation test"},
		Reason:      "Testing privilege escalation functionality",
		Interactive: false,
	}

	return s.ExecutePrivileged(req)
}