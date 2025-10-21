package middleware

import (
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

// Security returns security headers middleware
func Security(cfg *viper.Viper) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			res := c.Response()

			// Content Security Policy
			csp := buildCSP(cfg)
			res.Header().Set("Content-Security-Policy", csp)

			// Other security headers
			res.Header().Set("X-Content-Type-Options", "nosniff")
			res.Header().Set("X-Frame-Options", "DENY")
			res.Header().Set("X-XSS-Protection", "1; mode=block")
			res.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			res.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

			// HSTS for HTTPS
			if c.Request().TLS != nil || c.Request().Header.Get("X-Forwarded-Proto") == "https" {
				res.Header().Set("Strict-Transport-Security", 
					"max-age=31536000; includeSubDomains")
			}

			return next(c)
		}
	}
}

func buildCSP(cfg *viper.Viper) string {
	policies := map[string]string{
		"default-src":   "'self'",
		"script-src":    "'self' 'unsafe-inline'",
		"style-src":     "'self' 'unsafe-inline'",
		"img-src":       "'self' data: https:",
		"font-src":      "'self'",
		"connect-src":   "'self'",
		"media-src":     "'none'",
		"object-src":    "'none'",
		"frame-src":     "'none'",
		"base-uri":      "'self'",
		"form-action":   "'self'",
		"upgrade-insecure-requests": "",
	}

	// Allow additional domains for CDNs if configured
	cdnDomains := cfg.GetStringSlice("security.csp.cdn_domains")
	if len(cdnDomains) > 0 {
		scriptSrc := policies["script-src"] + " " + strings.Join(cdnDomains, " ")
		policies["script-src"] = scriptSrc
	}

	var parts []string
	for directive, value := range policies {
		if value == "" {
			parts = append(parts, directive)
		} else {
			parts = append(parts, fmt.Sprintf("%s %s", directive, value))
		}
	}

	return strings.Join(parts, "; ")
}