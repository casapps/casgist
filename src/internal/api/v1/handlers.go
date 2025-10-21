package v1

import (
	"net/http"
	"runtime"

	"github.com/labstack/echo/v4"
)

// healthHandler returns service health status
func healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "healthy",
		"time":   c.Request().Header.Get("X-Request-Start"),
	})
}

// versionHandler returns version information
func versionHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"version": "0.1.0",
		"go_version": runtime.Version(),
		"build_time": "development",
	})
}

// exploreHandler returns public gists for exploration
func exploreHandler(c echo.Context) error {
	// TODO: Implement explore functionality
	return c.JSON(http.StatusOK, map[string]interface{}{
		"gists": []interface{}{},
		"total": 0,
	})
}

// trendingHandler returns trending gists
func trendingHandler(c echo.Context) error {
	// TODO: Implement trending functionality
	return c.JSON(http.StatusOK, map[string]interface{}{
		"gists": []interface{}{},
		"total": 0,
	})
}