package domains

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// DNSValidator handles DNS-based domain verification
type DNSValidator struct {
	timeout time.Duration
}

// NewDNSValidator creates a new DNS validator
func NewDNSValidator() *DNSValidator {
	return &DNSValidator{
		timeout: 10 * time.Second,
	}
}

// VerifyOwnership verifies domain ownership through DNS TXT record
func (dv *DNSValidator) VerifyOwnership(domain, token string) error {
	// Check for TXT record at _casgists-challenge subdomain
	challengeDomain := fmt.Sprintf("_casgists-challenge.%s", domain)
	
	txtRecords, err := dv.lookupTXT(challengeDomain)
	if err != nil {
		return fmt.Errorf("failed to lookup TXT records for %s: %w", challengeDomain, err)
	}
	
	// Check if any TXT record contains our token
	for _, record := range txtRecords {
		if strings.Contains(record, token) {
			return nil // Verification successful
		}
	}
	
	return fmt.Errorf("verification token not found in DNS TXT records")
}

// VerifyDNSPointing verifies that domain points to our server
func (dv *DNSValidator) VerifyDNSPointing(domain, expectedIP string) error {
	ips, err := dv.lookupIP(domain)
	if err != nil {
		return fmt.Errorf("failed to lookup IP for %s: %w", domain, err)
	}
	
	for _, ip := range ips {
		if ip == expectedIP {
			return nil // Domain points to our server
		}
	}
	
	return fmt.Errorf("domain does not point to expected IP %s", expectedIP)
}

// CheckCNAME checks if domain has correct CNAME record
func (dv *DNSValidator) CheckCNAME(domain, expectedCNAME string) error {
	cname, err := dv.lookupCNAME(domain)
	if err != nil {
		return fmt.Errorf("failed to lookup CNAME for %s: %w", domain, err)
	}
	
	if cname != expectedCNAME {
		return fmt.Errorf("CNAME record %s does not match expected %s", cname, expectedCNAME)
	}
	
	return nil
}

// lookupTXT performs TXT record lookup with timeout
func (dv *DNSValidator) lookupTXT(domain string) ([]string, error) {
	// Create a channel to receive the result
	resultChan := make(chan []string, 1)
	errorChan := make(chan error, 1)
	
	// Perform lookup in goroutine
	go func() {
		records, err := net.LookupTXT(domain)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- records
	}()
	
	// Wait for result or timeout
	select {
	case records := <-resultChan:
		return records, nil
	case err := <-errorChan:
		return nil, err
	case <-time.After(dv.timeout):
		return nil, fmt.Errorf("DNS lookup timeout for %s", domain)
	}
}

// lookupIP performs IP lookup with timeout
func (dv *DNSValidator) lookupIP(domain string) ([]string, error) {
	// Create a channel to receive the result
	resultChan := make(chan []string, 1)
	errorChan := make(chan error, 1)
	
	// Perform lookup in goroutine
	go func() {
		ips, err := net.LookupHost(domain)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- ips
	}()
	
	// Wait for result or timeout
	select {
	case ips := <-resultChan:
		return ips, nil
	case err := <-errorChan:
		return nil, err
	case <-time.After(dv.timeout):
		return nil, fmt.Errorf("DNS lookup timeout for %s", domain)
	}
}

// lookupCNAME performs CNAME lookup with timeout
func (dv *DNSValidator) lookupCNAME(domain string) (string, error) {
	// Create a channel to receive the result
	resultChan := make(chan string, 1)
	errorChan := make(chan error, 1)
	
	// Perform lookup in goroutine
	go func() {
		cname, err := net.LookupCNAME(domain)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- cname
	}()
	
	// Wait for result or timeout
	select {
	case cname := <-resultChan:
		return cname, nil
	case err := <-errorChan:
		return "", err
	case <-time.After(dv.timeout):
		return "", fmt.Errorf("DNS lookup timeout for %s", domain)
	}
}

// GetDNSInfo gets comprehensive DNS information for a domain
func (dv *DNSValidator) GetDNSInfo(domain string) (*DNSInfo, error) {
	info := &DNSInfo{
		Domain: domain,
	}
	
	// Get A records
	if ips, err := dv.lookupIP(domain); err == nil {
		info.ARecords = ips
	}
	
	// Get TXT records
	if txtRecords, err := dv.lookupTXT(domain); err == nil {
		info.TXTRecords = txtRecords
	}
	
	// Get CNAME record
	if cname, err := dv.lookupCNAME(domain); err == nil {
		info.CNAME = cname
	}
	
	// Get NS records
	if nsRecords, err := net.LookupNS(domain); err == nil {
		nsStrings := make([]string, len(nsRecords))
		for i, ns := range nsRecords {
			nsStrings[i] = ns.Host
		}
		info.NSRecords = nsStrings
	}
	
	// Get MX records
	if mxRecords, err := net.LookupMX(domain); err == nil {
		mxStrings := make([]string, len(mxRecords))
		for i, mx := range mxRecords {
			mxStrings[i] = fmt.Sprintf("%d %s", mx.Pref, mx.Host)
		}
		info.MXRecords = mxStrings
	}
	
	return info, nil
}

// DNSInfo contains comprehensive DNS information
type DNSInfo struct {
	Domain     string   `json:"domain"`
	ARecords   []string `json:"a_records"`
	TXTRecords []string `json:"txt_records"`
	CNAME      string   `json:"cname,omitempty"`
	NSRecords  []string `json:"ns_records"`
	MXRecords  []string `json:"mx_records"`
}

// SetTimeout sets the DNS lookup timeout
func (dv *DNSValidator) SetTimeout(timeout time.Duration) {
	dv.timeout = timeout
}