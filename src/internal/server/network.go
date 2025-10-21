package server

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
)

// NetworkDetector detects network configuration and reverse proxy setup
type NetworkDetector struct {
	detectedIP   string
	detectedFQDN string
	reverseProxy string
}

// NewNetworkDetector creates a new network detector
func NewNetworkDetector() *NetworkDetector {
	nd := &NetworkDetector{}
	nd.detectNetwork()
	return nd
}

// detectNetwork detects the server's network configuration
func (nd *NetworkDetector) detectNetwork() {
	// Get default route IP
	nd.detectedIP = nd.GetDefaultRouteIP()
	
	// Try to get FQDN
	nd.detectedFQDN = nd.GetFQDN()
}

// GetDefaultRouteIP gets the IP address from the default route
func (nd *NetworkDetector) GetDefaultRouteIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()
	
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// GetFQDN attempts to get the fully qualified domain name
func (nd *NetworkDetector) GetFQDN() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	
	// Try to resolve FQDN
	if addrs, err := net.LookupAddr(nd.detectedIP); err == nil && len(addrs) > 0 {
		return strings.TrimSuffix(addrs[0], ".")
	}
	
	return hostname
}

// DetectReverseProxy detects if there's a reverse proxy
func (nd *NetworkDetector) DetectReverseProxy(c echo.Context) string {
	// Check common reverse proxy headers
	headers := []string{
		"X-Forwarded-Host",
		"X-Original-Host",
		"Host",
	}
	
	for _, header := range headers {
		if host := c.Request().Header.Get(header); host != "" {
			// Check if it's different from our detected address
			if host != nd.detectedIP && host != nd.detectedFQDN {
				return host
			}
		}
	}
	
	return ""
}

// GetBestURL returns the best URL for the service
func (nd *NetworkDetector) GetBestURL(c echo.Context, port int) string {
	// Check for reverse proxy first
	if proxy := nd.DetectReverseProxy(c); proxy != "" {
		scheme := "http"
		if c.Request().TLS != nil || c.Request().Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		return fmt.Sprintf("%s://%s", scheme, proxy)
	}
	
	// Use FQDN if available
	if nd.detectedFQDN != "" && nd.detectedFQDN != "localhost" {
		return fmt.Sprintf("http://%s:%d", nd.detectedFQDN, port)
	}
	
	// Fall back to IP
	return fmt.Sprintf("http://%s:%d", nd.detectedIP, port)
}

// GetDetectedIP returns the detected IP address
func (nd *NetworkDetector) GetDetectedIP() string {
	return nd.detectedIP
}

// GetDetectedFQDN returns the detected FQDN
func (nd *NetworkDetector) GetDetectedFQDN() string {
	return nd.detectedFQDN
}