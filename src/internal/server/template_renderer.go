package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// TemplateRenderer is a custom HTML template renderer for Echo
type TemplateRenderer struct {
	templates *template.Template
	debug     bool
}

// NewTemplateRenderer creates a new template renderer
func NewTemplateRenderer(templatesPath string, debug bool) (*TemplateRenderer, error) {
	fmt.Printf("NewTemplateRenderer called with path: %s, debug: %v\n", templatesPath, debug)
	funcMap := template.FuncMap{
		"timeago": timeAgo,
		"timeAgo": timeAgo,  // Add camelCase version
		"formatDate": formatDate,
		"filesize": formatFileSize,
		"default": defaultValue,
		"truncate": truncate,
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
		"title": strings.Title,
		"contains": strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
		"split": strings.Split,
		"join": strings.Join,
		"trim": strings.TrimSpace,
		"add": add,
		"sub": sub,
		"mul": mul,
		"div": div,
		"mod": mod,
		"dict": dict,
		"list": list,
		"safe": safe,
		"json": jsonEncode,
		"pluralize": pluralize,
		"substr": substr,
		"now": now,
	}

	// In development, we'll reload templates on each request
	if debug {
		return &TemplateRenderer{
			templates: nil,
			debug:     true,
		}, nil
	}

	// In production, load all templates once
	tmpl, err := loadTemplates(templatesPath, funcMap)
	if err != nil {
		return nil, err
	}

	return &TemplateRenderer{
		templates: tmpl,
		debug:     false,
	}, nil
}

// Render renders a template with the provided data
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// In debug mode, reload templates on each request
	if t.debug {
		templatesPath := filepath.Join("src", "web", "templates")
		tmpl, err := loadTemplates(templatesPath, getFuncMap())
		if err != nil {
			return err
		}
		t.templates = tmpl
	}

	// Add context values to template data
	if viewData, ok := data.(map[string]interface{}); ok {
		// Add request context
		viewData["Request"] = c.Request()
		viewData["Path"] = c.Path()
		viewData["URL"] = c.Request().URL.String()
		
		// Add user if authenticated
		if userID := c.Get("user_id"); userID != nil {
			viewData["UserID"] = userID
		}
		if username := c.Get("username"); username != nil {
			viewData["Username"] = username
		}
		
		// Add CSRF token
		viewData["CSRFToken"] = c.Get("csrf_token")
		
		// Add flash messages
		if flash := c.Get("flash"); flash != nil {
			viewData["Flash"] = flash
		}
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

// loadTemplates loads all templates from the templates directory
func loadTemplates(templatesPath string, funcMap template.FuncMap) (*template.Template, error) {
	// Create a new template
	tmpl := template.New("").Funcs(funcMap)

	// Load layouts first
	layoutFiles, err := filepath.Glob(filepath.Join(templatesPath, "layouts", "*.html"))
	if err != nil {
		return nil, err
	}
	if len(layoutFiles) > 0 {
		tmpl, err = tmpl.ParseFiles(layoutFiles...)
		if err != nil {
			return nil, fmt.Errorf("failed to parse layouts: %w", err)
		}
	}

	// Load all page templates
	pageFiles, err := filepath.Glob(filepath.Join(templatesPath, "pages", "*.html"))
	if err != nil {
		return nil, err
	}
	if len(pageFiles) > 0 {
		tmpl, err = tmpl.ParseFiles(pageFiles...)
		if err != nil {
			return nil, fmt.Errorf("failed to parse pages: %w", err)
		}
	}

	// Load setup templates
	setupFiles, err := filepath.Glob(filepath.Join(templatesPath, "pages", "setup", "*.html"))
	if err != nil {
		return nil, err
	}
	if len(setupFiles) > 0 {
		tmpl, err = tmpl.ParseFiles(setupFiles...)
		if err != nil {
			return nil, fmt.Errorf("failed to parse setup templates: %w", err)
		}
	}

	// Load all other subdirectory templates
	subdirs := []string{"auth", "gist", "admin", "user", "organization", "search", "public", "support"}
	for _, subdir := range subdirs {
		subdirFiles, err := filepath.Glob(filepath.Join(templatesPath, "pages", subdir, "*.html"))
		if err != nil {
			continue // Skip if directory doesn't exist
		}
		if len(subdirFiles) > 0 {
			tmpl, err = tmpl.ParseFiles(subdirFiles...)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %s templates: %w", subdir, err)
			}
		}
	}

	// Load partials
	partialFiles, err := filepath.Glob(filepath.Join(templatesPath, "partials", "*.html"))
	if err != nil {
		return nil, err
	}
	if len(partialFiles) > 0 {
		tmpl, err = tmpl.ParseFiles(partialFiles...)
		if err != nil {
			return nil, fmt.Errorf("failed to parse partials: %w", err)
		}
	}

	// Load components
	componentFiles, err := filepath.Glob(filepath.Join(templatesPath, "components", "*.html"))
	if err != nil {
		return nil, err
	}
	if len(componentFiles) > 0 {
		tmpl, err = tmpl.ParseFiles(componentFiles...)
		if err != nil {
			return nil, fmt.Errorf("failed to parse components: %w", err)
		}
	}

	return tmpl, nil
}

// getFuncMap returns the template function map
func getFuncMap() template.FuncMap {
	return template.FuncMap{
		"timeago": timeAgo,
		"timeAgo": timeAgo,  // Add camelCase version
		"formatDate": formatDate,
		"filesize": formatFileSize,
		"default": defaultValue,
		"truncate": truncate,
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
		"title": strings.Title,
		"contains": strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
		"split": strings.Split,
		"join": strings.Join,
		"trim": strings.TrimSpace,
		"add": add,
		"sub": sub,
		"mul": mul,
		"div": div,
		"mod": mod,
		"dict": dict,
		"list": list,
		"safe": safe,
		"json": jsonEncode,
		"pluralize": pluralize,
		"substr": substr,
		"now": now,
	}
}

// Template helper functions

func timeAgo(t time.Time) string {
	duration := time.Since(t)
	
	if duration.Seconds() < 60 {
		return "just now"
	} else if duration.Minutes() < 60 {
		n := int(duration.Minutes())
		if n == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", n)
	} else if duration.Hours() < 24 {
		n := int(duration.Hours())
		if n == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", n)
	} else if duration.Hours() < 24*30 {
		n := int(duration.Hours() / 24)
		if n == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", n)
	} else if duration.Hours() < 24*365 {
		n := int(duration.Hours() / (24 * 30))
		if n == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", n)
	}
	
	n := int(duration.Hours() / (24 * 365))
	if n == 1 {
		return "1 year ago"
	}
	return fmt.Sprintf("%d years ago", n)
}

func formatDate(t time.Time, format string) string {
	if format == "" {
		format = "Jan 2, 2006"
	}
	return t.Format(format)
}

func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func defaultValue(value, defaultValue interface{}) interface{} {
	if value == nil || value == "" {
		return defaultValue
	}
	return value
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func add(a, b int) int {
	return a + b
}

func sub(a, b int) int {
	return a - b
}

func mul(a, b int) int {
	return a * b
}

func div(a, b int) int {
	if b == 0 {
		return 0
	}
	return a / b
}

func mod(a, b int) int {
	if b == 0 {
		return 0
	}
	return a % b
}

func dict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("dict requires even number of arguments")
	}
	dict := make(map[string]interface{})
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict keys must be strings")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}

func list(values ...interface{}) []interface{} {
	return values
}

func safe(s string) template.HTML {
	return template.HTML(s)
}

func jsonEncode(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, singular)
	}
	if plural == "" {
		plural = singular + "s"
	}
	return fmt.Sprintf("%d %s", count, plural)
}

func substr(s string, start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(s) {
		end = len(s)
	}
	if start > end {
		return ""
	}
	return s[start:end]
}

func now() time.Time {
	return time.Now()
}