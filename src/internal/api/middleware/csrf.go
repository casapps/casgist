package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

const (
	csrfTokenKey   = "csrf_token"
	csrfCookieName = "csrf_token"
	csrfHeaderName = "X-CSRF-Token"
	csrfFormField  = "csrf_token"
)

// CSRFConfig holds CSRF middleware configuration
type CSRFConfig struct {
	TokenLength    int
	CookiePath     string
	CookieDomain   string
	CookieSecure   bool
	CookieHTTPOnly bool
	CookieSameSite http.SameSite
	Expiration     time.Duration
}

// DefaultCSRFConfig returns default CSRF configuration
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		TokenLength:    32,
		CookiePath:     "/",
		CookieSecure:   false, // Set to true in production with HTTPS
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteStrictMode,
		Expiration:     24 * time.Hour,
	}
}

// CSRF returns CSRF protection middleware
func CSRF(config *viper.Viper) echo.MiddlewareFunc {
	// Check if CSRF is disabled
	if config != nil && config.GetBool("security.disable_csrf") {
		// Return a no-op middleware that just sets a token for templates
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				// Set a dummy token for templates that might need it
				c.Set(csrfTokenKey, "disabled")
				return next(c)
			}
		}
	}
	
	cfg := DefaultCSRFConfig()
	
	// Override with config values if available
	if config != nil {
		if config.GetBool("server.enable_https") {
			cfg.CookieSecure = true
		}
		if domain := config.GetString("server.cookie_domain"); domain != "" {
			cfg.CookieDomain = domain
		}
	}
	
	return CSRFWithConfig(cfg)
}

// CSRFWithConfig returns CSRF middleware with custom config
func CSRFWithConfig(config CSRFConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			
			// Skip CSRF for safe methods
			if req.Method == http.MethodGet || req.Method == http.MethodHead || 
			   req.Method == http.MethodOptions || req.Method == http.MethodTrace {
				// Generate token for forms
				token, err := generateCSRFToken(config.TokenLength)
				if err != nil {
					return err
				}
				
				// Set cookie
				cookie := &http.Cookie{
					Name:     csrfCookieName,
					Value:    token,
					Path:     config.CookiePath,
					Domain:   config.CookieDomain,
					Expires:  time.Now().Add(config.Expiration),
					Secure:   config.CookieSecure,
					HttpOnly: config.CookieHTTPOnly,
					SameSite: config.CookieSameSite,
				}
				c.SetCookie(cookie)
				
				// Set token in context for templates
				c.Set(csrfTokenKey, token)
				
				return next(c)
			}
			
			// For unsafe methods, verify CSRF token
			cookie, err := c.Cookie(csrfCookieName)
			if err != nil {
				return echo.NewHTTPError(http.StatusForbidden, "CSRF cookie not found")
			}
			
			// Check token from multiple sources
			token := extractCSRFToken(c)
			if token == "" {
				return echo.NewHTTPError(http.StatusForbidden, "CSRF token not found")
			}
			
			// Validate token
			if cookie.Value != token {
				return echo.NewHTTPError(http.StatusForbidden, "Invalid CSRF token")
			}
			
			// Set token in context for next handlers
			c.Set(csrfTokenKey, token)
			
			return next(c)
		}
	}
}

// generateCSRFToken generates a random CSRF token
func generateCSRFToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// extractCSRFToken extracts CSRF token from request
func extractCSRFToken(c echo.Context) string {
	// Try header first
	if token := c.Request().Header.Get(csrfHeaderName); token != "" {
		return token
	}
	
	// Try form value
	if token := c.FormValue(csrfFormField); token != "" {
		return token
	}
	
	// Try query parameter (for AJAX requests)
	if token := c.QueryParam(csrfFormField); token != "" {
		return token
	}
	
	return ""
}

// GetCSRFToken returns the CSRF token from context
func GetCSRFToken(c echo.Context) string {
	if token, ok := c.Get(csrfTokenKey).(string); ok {
		return token
	}
	return ""
}

// CSRFTokenFunc returns a template function for CSRF tokens
func CSRFTokenFunc(c echo.Context) func() string {
	return func() string {
		return GetCSRFToken(c)
	}
}