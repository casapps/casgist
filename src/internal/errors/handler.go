package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// ErrorHandler provides comprehensive error handling for CasGists
type ErrorHandler struct {
	config     *viper.Viper
	db         *gorm.DB
	production bool
	logger     ErrorLogger
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(config *viper.Viper, db *gorm.DB) *ErrorHandler {
	return &ErrorHandler{
		config:     config,
		db:         db,
		production: config.GetString("environment") == "production",
		logger:     NewErrorLogger(config, db),
	}
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error      string                 `json:"error"`
	Message    string                 `json:"message"`
	Code       string                 `json:"code,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	RequestID  string                 `json:"request_id,omitempty"`
	Path       string                 `json:"path,omitempty"`
	Method     string                 `json:"method,omitempty"`
	StatusCode int                    `json:"status_code"`
}

// CustomError represents a custom application error
type CustomError struct {
	Type       ErrorType              `json:"type"`
	Message    string                 `json:"message"`
	Code       string                 `json:"code"`
	StatusCode int                    `json:"status_code"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Cause      error                  `json:"-"`
}

// ErrorType represents different types of errors
type ErrorType string

const (
	ErrorTypeValidation    ErrorType = "validation_error"
	ErrorTypeDatabase      ErrorType = "database_error"
	ErrorTypeAuthentication ErrorType = "authentication_error"
	ErrorTypeAuthorization ErrorType = "authorization_error"
	ErrorTypeNotFound      ErrorType = "not_found_error"
	ErrorTypeConflict      ErrorType = "conflict_error"
	ErrorTypeRateLimit     ErrorType = "rate_limit_error"
	ErrorTypeServer        ErrorType = "server_error"
	ErrorTypeNetwork       ErrorType = "network_error"
	ErrorTypeExternal      ErrorType = "external_service_error"
	ErrorTypeTimeout       ErrorType = "timeout_error"
	ErrorTypeStorage       ErrorType = "storage_error"
	ErrorTypeConfig        ErrorType = "configuration_error"
)

// Error implements the error interface
func (e *CustomError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// NewCustomError creates a new custom error
func NewCustomError(errorType ErrorType, message, code string, statusCode int) *CustomError {
	return &CustomError{
		Type:       errorType,
		Message:    message,
		Code:       code,
		StatusCode: statusCode,
		Details:    make(map[string]interface{}),
	}
}

// WithCause adds a cause to the error
func (e *CustomError) WithCause(cause error) *CustomError {
	e.Cause = cause
	return e
}

// WithDetail adds a detail to the error
func (e *CustomError) WithDetail(key string, value interface{}) *CustomError {
	e.Details[key] = value
	return e
}

// Common error constructors
func NewValidationError(message, field string) *CustomError {
	return NewCustomError(ErrorTypeValidation, message, "VALIDATION_FAILED", http.StatusBadRequest).
		WithDetail("field", field)
}

func DatabaseError(message string, cause error) *CustomError {
	return NewCustomError(ErrorTypeDatabase, message, "DATABASE_ERROR", http.StatusInternalServerError).
		WithCause(cause)
}

func NotFoundError(resource, id string) *CustomError {
	return NewCustomError(ErrorTypeNotFound, fmt.Sprintf("%s not found", resource), "NOT_FOUND", http.StatusNotFound).
		WithDetail("resource", resource).
		WithDetail("id", id)
}

func UnauthorizedError(message string) *CustomError {
	return NewCustomError(ErrorTypeAuthentication, message, "UNAUTHORIZED", http.StatusUnauthorized)
}

func ForbiddenError(message string) *CustomError {
	return NewCustomError(ErrorTypeAuthorization, message, "FORBIDDEN", http.StatusForbidden)
}

func ConflictError(message, resource string) *CustomError {
	return NewCustomError(ErrorTypeConflict, message, "CONFLICT", http.StatusConflict).
		WithDetail("resource", resource)
}

func RateLimitError(limit int, window string) *CustomError {
	return NewCustomError(ErrorTypeRateLimit, "Rate limit exceeded", "RATE_LIMITED", http.StatusTooManyRequests).
		WithDetail("limit", limit).
		WithDetail("window", window)
}

func TimeoutError(operation string, timeout time.Duration) *CustomError {
	return NewCustomError(ErrorTypeTimeout, "Operation timed out", "TIMEOUT", http.StatusRequestTimeout).
		WithDetail("operation", operation).
		WithDetail("timeout", timeout.String())
}

func ExternalServiceError(service, message string, cause error) *CustomError {
	return NewCustomError(ErrorTypeExternal, fmt.Sprintf("%s service error: %s", service, message), "EXTERNAL_SERVICE_ERROR", http.StatusBadGateway).
		WithDetail("service", service).
		WithCause(cause)
}

func StorageError(operation, message string, cause error) *CustomError {
	return NewCustomError(ErrorTypeStorage, message, "STORAGE_ERROR", http.StatusInternalServerError).
		WithDetail("operation", operation).
		WithCause(cause)
}

func ConfigError(key, message string) *CustomError {
	return NewCustomError(ErrorTypeConfig, message, "CONFIG_ERROR", http.StatusInternalServerError).
		WithDetail("config_key", key)
}

// HTTPErrorHandler handles HTTP errors for Echo
func (h *ErrorHandler) HTTPErrorHandler(err error, c echo.Context) {
	var (
		code    = http.StatusInternalServerError
		message = "Internal server error"
		details = make(map[string]interface{})
		errCode = "INTERNAL_ERROR"
	)

	// Extract request information
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)
	if requestID == "" {
		requestID = c.Request().Header.Get("X-Request-ID")
	}

	path := c.Request().URL.Path
	method := c.Request().Method

	// Handle different error types
	switch e := err.(type) {
	case *CustomError:
		code = e.StatusCode
		message = e.Message
		errCode = e.Code
		details = e.Details
		
		// Log the error with context
		h.logger.LogError(c.Request().Context(), e, map[string]interface{}{
			"request_id": requestID,
			"path":       path,
			"method":     method,
			"user_agent": c.Request().UserAgent(),
			"ip":         c.RealIP(),
		})

	case *echo.HTTPError:
		code = e.Code
		message = fmt.Sprintf("%v", e.Message)
		
		// Map common HTTP errors to our error codes
		switch code {
		case http.StatusNotFound:
			errCode = "NOT_FOUND"
			message = "Resource not found"
		case http.StatusMethodNotAllowed:
			errCode = "METHOD_NOT_ALLOWED"
			message = "Method not allowed"
		case http.StatusBadRequest:
			errCode = "BAD_REQUEST"
		case http.StatusUnauthorized:
			errCode = "UNAUTHORIZED"
			message = "Authentication required"
		case http.StatusForbidden:
			errCode = "FORBIDDEN"
			message = "Access denied"
		}

	case *json.SyntaxError:
		code = http.StatusBadRequest
		message = "Invalid JSON format"
		errCode = "INVALID_JSON"
		details["offset"] = e.Offset

	default:
		// Log unexpected errors with stack trace
		h.logger.LogError(c.Request().Context(), err, map[string]interface{}{
			"request_id": requestID,
			"path":       path,
			"method":     method,
			"stack":      string(debug.Stack()),
		})

		// Handle specific error types
		if strings.Contains(err.Error(), "connection refused") {
			code = http.StatusBadGateway
			message = "Service temporarily unavailable"
			errCode = "SERVICE_UNAVAILABLE"
		} else if strings.Contains(err.Error(), "timeout") {
			code = http.StatusRequestTimeout
			message = "Request timeout"
			errCode = "TIMEOUT"
		} else if strings.Contains(err.Error(), "database") {
			code = http.StatusInternalServerError
			message = "Database error"
			errCode = "DATABASE_ERROR"
		}
	}

	// Don't expose internal errors in production
	if h.production && code == http.StatusInternalServerError {
		message = "Internal server error"
		details = map[string]interface{}{
			"error_id": requestID,
		}
	}

	// Create error response
	errorResponse := ErrorResponse{
		Error:      message,
		Message:    message,
		Code:       errCode,
		Details:    details,
		Timestamp:  time.Now(),
		RequestID:  requestID,
		Path:       path,
		Method:     method,
		StatusCode: code,
	}

	// Send error response
	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead {
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, errorResponse)
		}
		if err != nil {
			h.logger.LogError(c.Request().Context(), fmt.Errorf("failed to send error response: %w", err), nil)
		}
	}
}

// RecoverMiddleware provides panic recovery
func (h *ErrorHandler) RecoverMiddleware() echo.MiddlewareFunc {
	return middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			h.logger.LogError(c.Request().Context(), err, map[string]interface{}{
				"panic_stack": string(stack),
				"request_id":  c.Response().Header().Get(echo.HeaderXRequestID),
				"path":        c.Request().URL.Path,
				"method":      c.Request().Method,
			})
			return err
		},
	})
}

// ValidationMiddleware provides request validation
func (h *ErrorHandler) ValidationMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Set up validation context
			c.Set("validator", NewValidator())
			return next(c)
		}
	}
}

// TimeoutMiddleware provides request timeout handling
func (h *ErrorHandler) TimeoutMiddleware(timeout time.Duration) echo.MiddlewareFunc {
	return middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: timeout,
		ErrorMessage: "Request timeout",
	})
}

// DatabaseErrorWrapper wraps database operations with error handling
func (h *ErrorHandler) DatabaseErrorWrapper(operation string, fn func() error) error {
	err := fn()
	if err != nil {
		// Handle specific database errors
		if err == gorm.ErrRecordNotFound {
			return NotFoundError("Resource", "")
		}
		
		// Handle database constraint violations
		if strings.Contains(err.Error(), "UNIQUE constraint failed") ||
		   strings.Contains(err.Error(), "duplicate key") {
			return ConflictError("Resource already exists", "")
		}
		
		// Handle foreign key violations
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return NewValidationError("Referenced resource does not exist", "")
		}
		
		// Generic database error
		return DatabaseError(fmt.Sprintf("Database operation failed: %s", operation), err)
	}
	return nil
}

// ExternalAPIWrapper wraps external API calls with error handling
func (h *ErrorHandler) ExternalAPIWrapper(service string, fn func() error) error {
	err := fn()
	if err != nil {
		// Handle HTTP client errors
		if strings.Contains(err.Error(), "connection refused") ||
		   strings.Contains(err.Error(), "no such host") {
			return ExternalServiceError(service, "Service unavailable", err)
		}
		
		if strings.Contains(err.Error(), "timeout") {
			return TimeoutError(fmt.Sprintf("%s API call", service), 30*time.Second)
		}
		
		return ExternalServiceError(service, err.Error(), err)
	}
	return nil
}

// StorageOperationWrapper wraps storage operations with error handling
func (h *ErrorHandler) StorageOperationWrapper(operation string, fn func() error) error {
	err := fn()
	if err != nil {
		// Handle storage-specific errors
		if strings.Contains(err.Error(), "no space left") {
			return StorageError(operation, "Storage space exhausted", err)
		}
		
		if strings.Contains(err.Error(), "permission denied") {
			return StorageError(operation, "Storage permission denied", err)
		}
		
		return StorageError(operation, "Storage operation failed", err)
	}
	return nil
}

// GetErrorStats returns error statistics
func (h *ErrorHandler) GetErrorStats() (map[string]interface{}, error) {
	return h.logger.GetErrorStats()
}

// HealthCheck performs error handler health check
func (h *ErrorHandler) HealthCheck() error {
	// Test database connectivity
	if h.db != nil {
		sqlDB, err := h.db.DB()
		if err != nil {
			return ConfigError("database", "Failed to get database connection")
		}
		
		if err := sqlDB.Ping(); err != nil {
			return DatabaseError("Health check failed", err)
		}
	}
	
	return nil
}