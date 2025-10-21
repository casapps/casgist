package utils

import (
	"os"
	"runtime"
)

// IsElevated checks if the process is running with elevated privileges
func IsElevated() bool {
	switch runtime.GOOS {
	case "windows":
		// On Windows, check if running as admin
		// This is a simplified check
		_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
		return err == nil
	default:
		// On Unix-like systems, check if running as root
		return os.Geteuid() == 0
	}
}