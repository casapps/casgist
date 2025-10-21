package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

// CORS returns a CORS middleware configured from settings
func CORS(cfg *viper.Viper) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			origin := req.Header.Get("Origin")

			// Skip CORS for same-origin requests
			if origin == "" {
				return next(c)
			}

			// Get allowed origins from config
			allowedOrigins := cfg.GetStringSlice("cors.allowed_origins")
			if len(allowedOrigins) == 0 {
				allowedOrigins = []string{"*"}
			}

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}

				// Support wildcard subdomains
				if strings.HasPrefix(allowedOrigin, "*.") {
					domain := strings.TrimPrefix(allowedOrigin, "*.")
					if strings.HasSuffix(origin, domain) {
						allowed = true
						break
					}
				}
			}

			// Public read-only endpoints allow any origin
			if !allowed && isPublicReadEndpoint(c) {
				allowed = true
			}

			if !allowed {
				return echo.NewHTTPError(http.StatusForbidden, "CORS: origin not allowed")
			}

			// Set CORS headers
			res.Header().Set("Access-Control-Allow-Origin", origin)
			res.Header().Set("Access-Control-Allow-Methods", 
				cfg.GetString("cors.allowed_methods"))
			res.Header().Set("Access-Control-Allow-Headers", 
				cfg.GetString("cors.allowed_headers"))
			res.Header().Set("Access-Control-Expose-Headers", 
				cfg.GetString("cors.exposed_headers"))
			res.Header().Set("Access-Control-Max-Age", 
				strconv.Itoa(cfg.GetInt("cors.max_age")))

			if cfg.GetBool("cors.allow_credentials") {
				res.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight requests
			if req.Method == "OPTIONS" {
				return c.NoContent(http.StatusNoContent)
			}

			return next(c)
		}
	}
}

func isPublicReadEndpoint(c echo.Context) bool {
	path := c.Request().URL.Path
	method := c.Request().Method

	if method == "GET" {
		publicPaths := []string{
			"/api/v1/gists/public",
			"/api/v1/explore",
			"/api/v1/trending",
			"/healthz",
		}

		for _, publicPath := range publicPaths {
			if strings.HasPrefix(path, publicPath) {
				return true
			}
		}
	}

	return false
}