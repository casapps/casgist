package auth

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TOTPService handles two-factor authentication
type TOTPService struct {
	issuer string
}

// NewTOTPService creates a new TOTP service
func NewTOTPService(issuer string) *TOTPService {
	return &TOTPService{
		issuer: issuer,
	}
}

// TOTPSetup contains the setup information for TOTP
type TOTPSetup struct {
	Secret    string `json:"secret"`
	URL       string `json:"url"`
	QRCode    string `json:"qr_code"`
}

// GenerateTOTP generates a new TOTP secret for a user
func (t *TOTPService) GenerateTOTP(username string) (*TOTPSetup, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      t.issuer,
		AccountName: username,
		Period:      30,
		SecretSize:  32,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP key: %w", err)
	}
	
	// Generate QR code
	var buf bytes.Buffer
	img, err := key.Image(256, 256)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR image: %w", err)
	}
	
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode QR image: %w", err)
	}
	
	// Encode QR code as base64 data URL
	qrCode := fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(buf.Bytes()))
	
	return &TOTPSetup{
		Secret: key.Secret(),
		URL:    key.URL(),
		QRCode: qrCode,
	}, nil
}

// ValidateTOTP validates a TOTP code
func (t *TOTPService) ValidateTOTP(secret, code string) bool {
	return totp.Validate(code, secret)
}

// GenerateRecoveryCodes generates recovery codes for 2FA
func GenerateRecoveryCodes(count int) ([]string, error) {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		code, err := GenerateSecureToken(4)
		if err != nil {
			return nil, err
		}
		// Format as XXXX-XXXX for readability
		if len(code) >= 8 {
			codes[i] = fmt.Sprintf("%s-%s", code[:4], code[4:8])
		} else {
			codes[i] = code
		}
	}
	return codes, nil
}