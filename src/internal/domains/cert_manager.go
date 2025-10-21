package domains

import (
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CertificateManager handles SSL certificate operations
type CertificateManager struct {
	certDir     string
	acmeDir     string
	email       string
	staging     bool
	acmeClient  ACMEClient
}

// ACMEClient interface for ACME protocol operations
type ACMEClient interface {
	ObtainCertificate(domain string) (certPath, keyPath string, expiresAt time.Time, err error)
	RenewCertificate(certPath string) (newCertPath, newKeyPath string, expiresAt time.Time, err error)
	RevokeCertificate(certPath string) error
}

// NewCertificateManager creates a new certificate manager
func NewCertificateManager(certDir, email string, staging bool) *CertificateManager {
	return &CertificateManager{
		certDir:    certDir,
		acmeDir:    filepath.Join(certDir, "acme"),
		email:      email,
		staging:    staging,
		acmeClient: NewLetsEncryptClient(email, staging),
	}
}

// ObtainCertificate obtains a new SSL certificate
func (cm *CertificateManager) ObtainCertificate(domain string) (string, string, time.Time, error) {
	// Ensure directories exist
	if err := os.MkdirAll(cm.certDir, 0700); err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to create cert directory: %w", err)
	}
	
	if err := os.MkdirAll(cm.acmeDir, 0700); err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to create ACME directory: %w", err)
	}
	
	// Use ACME client to obtain certificate
	certPath, keyPath, expiresAt, err := cm.acmeClient.ObtainCertificate(domain)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("ACME certificate request failed: %w", err)
	}
	
	// Move certificates to our directory structure
	finalCertPath := filepath.Join(cm.certDir, fmt.Sprintf("%s.crt", domain))
	finalKeyPath := filepath.Join(cm.certDir, fmt.Sprintf("%s.key", domain))
	
	if err := cm.moveCertificate(certPath, finalCertPath); err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to move certificate: %w", err)
	}
	
	if err := cm.moveCertificate(keyPath, finalKeyPath); err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to move private key: %w", err)
	}
	
	return finalCertPath, finalKeyPath, expiresAt, nil
}

// RenewCertificate renews an existing SSL certificate
func (cm *CertificateManager) RenewCertificate(certPath string) (string, string, time.Time, error) {
	// Use ACME client to renew certificate
	newCertPath, newKeyPath, expiresAt, err := cm.acmeClient.RenewCertificate(certPath)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("ACME certificate renewal failed: %w", err)
	}
	
	return newCertPath, newKeyPath, expiresAt, nil
}

// RevokeCertificate revokes an SSL certificate
func (cm *CertificateManager) RevokeCertificate(certPath string) error {
	// Use ACME client to revoke certificate
	if err := cm.acmeClient.RevokeCertificate(certPath); err != nil {
		return fmt.Errorf("ACME certificate revocation failed: %w", err)
	}
	
	// Remove certificate files
	keyPath := cm.getCertificateKeyPath(certPath)
	
	if err := os.Remove(certPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove certificate file: %w", err)
	}
	
	if err := os.Remove(keyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove key file: %w", err)
	}
	
	return nil
}

// ValidateCertificate validates an SSL certificate
func (cm *CertificateManager) ValidateCertificate(certPath string) error {
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("failed to read certificate: %w", err)
	}
	
	// Parse certificate
	cert, err := x509.ParseCertificate(certData)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}
	
	// Check if certificate is expired
	now := time.Now()
	if now.Before(cert.NotBefore) {
		return fmt.Errorf("certificate is not yet valid")
	}
	
	if now.After(cert.NotAfter) {
		return fmt.Errorf("certificate has expired")
	}
	
	return nil
}

// GetCertificateInfo gets information about a certificate
func (cm *CertificateManager) GetCertificateInfo(certPath string) (*CertificateInfo, error) {
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %w", err)
	}
	
	// Parse certificate
	cert, err := x509.ParseCertificate(certData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}
	
	return &CertificateInfo{
		Subject:    cert.Subject.String(),
		Issuer:     cert.Issuer.String(),
		NotBefore:  cert.NotBefore,
		NotAfter:   cert.NotAfter,
		DNSNames:   cert.DNSNames,
		SerialNumber: cert.SerialNumber.String(),
	}, nil
}

// moveCertificate moves a certificate file to the target location
func (cm *CertificateManager) moveCertificate(source, target string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	targetFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer targetFile.Close()
	
	_, err = targetFile.ReadFrom(sourceFile)
	if err != nil {
		return err
	}
	
	// Remove source file
	return os.Remove(source)
}

// getCertificateKeyPath gets the key path for a certificate
func (cm *CertificateManager) getCertificateKeyPath(certPath string) string {
	dir := filepath.Dir(certPath)
	name := filepath.Base(certPath)
	keyName := name[:len(name)-len(filepath.Ext(name))] + ".key"
	return filepath.Join(dir, keyName)
}

// CertificateInfo contains information about a certificate
type CertificateInfo struct {
	Subject      string    `json:"subject"`
	Issuer       string    `json:"issuer"`
	NotBefore    time.Time `json:"not_before"`
	NotAfter     time.Time `json:"not_after"`
	DNSNames     []string  `json:"dns_names"`
	SerialNumber string    `json:"serial_number"`
}

// MockACMEClient is a mock implementation for testing
type MockACMEClient struct {
	certDir string
}

// NewMockACMEClient creates a new mock ACME client
func NewMockACMEClient(certDir string) *MockACMEClient {
	return &MockACMEClient{
		certDir: certDir,
	}
}

// ObtainCertificate mocks certificate generation
func (m *MockACMEClient) ObtainCertificate(domain string) (string, string, time.Time, error) {
	certPath := filepath.Join(m.certDir, fmt.Sprintf("%s.crt", domain))
	keyPath := filepath.Join(m.certDir, fmt.Sprintf("%s.key", domain))
	expiresAt := time.Now().AddDate(0, 3, 0) // 3 months from now
	
	// Create mock certificate files
	certContent := fmt.Sprintf("-----BEGIN CERTIFICATE-----\nMock certificate for %s\n-----END CERTIFICATE-----\n", domain)
	keyContent := fmt.Sprintf("-----BEGIN PRIVATE KEY-----\nMock private key for %s\n-----END PRIVATE KEY-----\n", domain)
	
	if err := os.WriteFile(certPath, []byte(certContent), 0600); err != nil {
		return "", "", time.Time{}, err
	}
	
	if err := os.WriteFile(keyPath, []byte(keyContent), 0600); err != nil {
		return "", "", time.Time{}, err
	}
	
	return certPath, keyPath, expiresAt, nil
}

// RenewCertificate mocks certificate renewal
func (m *MockACMEClient) RenewCertificate(certPath string) (string, string, time.Time, error) {
	// For mock, just update the existing files
	domain := filepath.Base(certPath)
	domain = domain[:len(domain)-4] // Remove .crt extension
	
	return m.ObtainCertificate(domain)
}

// RevokeCertificate mocks certificate revocation
func (m *MockACMEClient) RevokeCertificate(certPath string) error {
	// For mock, this is always successful
	return nil
}

// LetsEncryptClient implements ACME client for Let's Encrypt
type LetsEncryptClient struct {
	email   string
	staging bool
}

// NewLetsEncryptClient creates a new Let's Encrypt client
func NewLetsEncryptClient(email string, staging bool) *LetsEncryptClient {
	return &LetsEncryptClient{
		email:   email,
		staging: staging,
	}
}

// ObtainCertificate obtains a certificate from Let's Encrypt
func (l *LetsEncryptClient) ObtainCertificate(domain string) (string, string, time.Time, error) {
	// This would integrate with a real ACME library like lego or autocert
	// For now, return mock implementation
	return "", "", time.Time{}, fmt.Errorf("Let's Encrypt integration not implemented")
}

// RenewCertificate renews a certificate with Let's Encrypt
func (l *LetsEncryptClient) RenewCertificate(certPath string) (string, string, time.Time, error) {
	// This would integrate with a real ACME library
	return "", "", time.Time{}, fmt.Errorf("Let's Encrypt integration not implemented")
}

// RevokeCertificate revokes a certificate with Let's Encrypt
func (l *LetsEncryptClient) RevokeCertificate(certPath string) error {
	// This would integrate with a real ACME library
	return fmt.Errorf("Let's Encrypt integration not implemented")
}