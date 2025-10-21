package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// DatabaseInjector injects the database connection into context
func DatabaseInjector(db *gorm.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("db", db)
			return next(c)
		}
	}
}

// ConfigInjector injects the configuration into context
func ConfigInjector(cfg *viper.Viper) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("config", cfg)
			return next(c)
		}
	}
}

// GetDB extracts database from context
func GetDB(c echo.Context) *gorm.DB {
	return c.Get("db").(*gorm.DB)
}

// GetConfig extracts configuration from context
func GetConfig(c echo.Context) *viper.Viper {
	return c.Get("config").(*viper.Viper)
}