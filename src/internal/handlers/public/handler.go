package public

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// Handler handles public web pages
type Handler struct {
	db     *gorm.DB
	config *viper.Viper
}

// NewHandler creates a new public handler
func NewHandler(db *gorm.DB, cfg *viper.Viper) *Handler {
	return &Handler{
		db:     db,
		config: cfg,
	}
}

// Home renders the landing page
func (h *Handler) Home(c echo.Context) error {
	data := map[string]interface{}{
		"message":     "Welcome to CasGists",
		"title":       h.config.GetString("ui.title"),
		"description": h.config.GetString("ui.description"),
		"version":     h.config.GetString("version"),
	}
	
	return c.JSON(http.StatusOK, data)
}

// Explore renders the explore page
func (h *Handler) Explore(c echo.Context) error {
	data := map[string]interface{}{
		"message":     "Explore public gists",
		"title":       "Explore - " + h.config.GetString("ui.title"),
		"description": "Discover public gists",
	}
	
	return c.JSON(http.StatusOK, data)
}

// Trending renders the trending page
func (h *Handler) Trending(c echo.Context) error {
	data := map[string]interface{}{
		"message":     "Trending gists",
		"title":       "Trending - " + h.config.GetString("ui.title"),
		"description": "Trending gists",
	}
	
	return c.JSON(http.StatusOK, data)
}