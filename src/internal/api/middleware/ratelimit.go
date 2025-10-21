package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

// RateLimit returns a rate limiting middleware
func RateLimit(cfg *viper.Viper) echo.MiddlewareFunc {
	// For now, return a simple middleware that does nothing
	// TODO: Implement proper rate limiting
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}
}