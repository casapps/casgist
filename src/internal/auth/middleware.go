package auth

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// Middleware provides authentication middleware
type Middleware struct {
	authService *AuthService
	skipper     func(c echo.Context) bool
}

// NewMiddleware creates a new authentication middleware
func NewMiddleware(authService *AuthService) *Middleware {
	return &Middleware{
		authService: authService,
		skipper:     DefaultSkipper,
	}
}

// DefaultSkipper returns true for paths that don't require authentication
func DefaultSkipper(c echo.Context) bool {
	path := c.Path()
	
	// Public paths that don't require authentication
	publicPaths := []string{
		"/",
		"/login",
		"/register",
		"/forgot-password",
		"/reset-password",
		"/health",
		"/metrics",
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/refresh",
		"/static/*",
		"/favicon.ico",
		"/robots.txt",
		"/manifest.json",
		"/service-worker.js",
	}
	
	// Check exact matches
	for _, p := range publicPaths {
		if p == path {
			return true
		}
		// Check wildcard matches
		if strings.HasSuffix(p, "*") && strings.HasPrefix(path, strings.TrimSuffix(p, "*")) {
			return true
		}
	}
	
	// Public gist viewing
	if strings.HasPrefix(path, "/g/") && c.Request().Method == http.MethodGet {
		return true
	}
	
	return false
}

// Auth returns the authentication middleware handler
func (m *Middleware) Auth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip authentication for certain paths
			if m.skipper != nil && m.skipper(c) {
				return next(c)
			}
			
			// Extract token from header
			auth := c.Request().Header.Get("Authorization")
			if auth == "" {
				// Check for token in cookie
				cookie, err := c.Cookie("access_token")
				if err != nil {
					return echo.NewHTTPError(http.StatusUnauthorized, "missing authentication")
				}
				auth = "Bearer " + cookie.Value
			}
			
			// Validate Bearer token format
			parts := strings.Split(auth, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authentication format")
			}
			
			// Validate token
			claims, err := m.authService.ValidateToken(parts[1])
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
			}
			
			// Store user information in context
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("email", claims.Email)
			c.Set("is_admin", claims.IsAdmin)
			c.Set("session_id", claims.SessionID)
			
			return next(c)
		}
	}
}

// RequireAdmin returns middleware that requires admin privileges
func (m *Middleware) RequireAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			isAdmin, ok := c.Get("is_admin").(bool)
			if !ok || !isAdmin {
				return echo.NewHTTPError(http.StatusForbidden, "admin privileges required")
			}
			return next(c)
		}
	}
}

// OptionalAuth returns middleware that sets user context if authenticated but doesn't require it
func (m *Middleware) OptionalAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract token from header
			auth := c.Request().Header.Get("Authorization")
			if auth == "" {
				// Check for token in cookie
				cookie, err := c.Cookie("access_token")
				if err == nil {
					auth = "Bearer " + cookie.Value
				}
			}
			
			// If no auth provided, continue without setting user context
			if auth == "" {
				return next(c)
			}
			
			// Validate Bearer token format
			parts := strings.Split(auth, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				// Validate token
				claims, err := m.authService.ValidateToken(parts[1])
				if err == nil {
					// Store user information in context
					c.Set("user_id", claims.UserID)
					c.Set("username", claims.Username)
					c.Set("email", claims.Email)
					c.Set("is_admin", claims.IsAdmin)
					c.Set("session_id", claims.SessionID)
				}
			}
			
			return next(c)
		}
	}
}