//go:build compliance
// +build compliance

package compliance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// AuditMiddleware provides comprehensive request/response auditing
type AuditMiddleware struct {
	auditService     *AuditService
	enabledPaths     []string // Paths to audit (empty = audit all)
	sensitiveFields  []string // Fields to redact from audit logs
	auditRequestBody bool     // Whether to audit request bodies
	auditResponseBody bool    // Whether to audit response bodies
	maxBodySize      int64    // Maximum body size to audit (bytes)
}

// NewAuditMiddleware creates a new audit middleware
func NewAuditMiddleware(auditService *AuditService) *AuditMiddleware {
	return &AuditMiddleware{
		auditService: auditService,
		sensitiveFields: []string{
			"password", "token", "secret", "key", "auth",
			"credit_card", "ssn", "phone", "email",
		},
		auditRequestBody:  true,
		auditResponseBody: false, // Usually too verbose
		maxBodySize:       10240, // 10KB default
	}
}

// Middleware returns the Echo middleware function
func (am *AuditMiddleware) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip if path is not in enabled paths (if specified)
			if len(am.enabledPaths) > 0 {
				path := c.Request().URL.Path
				enabled := false
				for _, enabledPath := range am.enabledPaths {
					if strings.HasPrefix(path, enabledPath) {
						enabled = true
						break
					}
				}
				if !enabled {
					return next(c)
				}
			}
			
			// Capture request details
			req := c.Request()
			startTime := time.Now()
			
			// Read and restore request body if needed
			var requestBody string
			if am.auditRequestBody && req.Body != nil && req.ContentLength > 0 && req.ContentLength <= am.maxBodySize {
				bodyBytes, err := io.ReadAll(req.Body)
				if err == nil {
					requestBody = am.sanitizeJSON(string(bodyBytes))
					// Restore body for handler
					req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
			}
			
			// Capture response if needed
			var responseBody string
			var responseWriter *auditResponseWriter
			if am.auditResponseBody {
				responseWriter = &auditResponseWriter{
					ResponseWriter: c.Response().Writer,
					body:          &bytes.Buffer{},
					maxSize:       am.maxBodySize,
				}
				c.Response().Writer = responseWriter
			}
			
			// Process request
			err := next(c)
			
			// Calculate duration
			duration := time.Since(startTime)
			
			// Get user ID from context if available
			var userID *uuid.UUID
			if user := c.Get("user"); user != nil {
				if userUUID, ok := user.(uuid.UUID); ok {
					userID = &userUUID
				}
			}
			
			// Capture response body if enabled
			if responseWriter != nil {
				responseBody = am.sanitizeJSON(responseWriter.body.String())
			}
			
			// Create audit log
			auditDetails := map[string]interface{}{
				"method":        req.Method,
				"path":          req.URL.Path,
				"query_params":  req.URL.RawQuery,
				"user_agent":    req.UserAgent(),
				"referer":       req.Referer(),
				"content_type":  req.Header.Get("Content-Type"),
				"duration_ms":   duration.Milliseconds(),
				"status_code":   c.Response().Status,
				"response_size": c.Response().Size,
			}
			
			if requestBody != "" {
				auditDetails["request_body"] = requestBody
			}
			
			if responseBody != "" {
				auditDetails["response_body"] = responseBody
			}
			
			// Log the request
			success := err == nil && c.Response().Status < 400
			action := fmt.Sprintf("%s %s", req.Method, req.URL.Path)
			
			if success {
				am.auditService.LogAction(
					userID,
					action,
					"http_request",
					"",
					auditDetails,
					req,
				)
			} else {
				errorMsg := ""
				if err != nil {
					errorMsg = err.Error()
				}
				am.auditService.LogFailedAction(
					userID,
					action,
					"http_request",
					"",
					errorMsg,
					req,
				)
			}
			
			// Log security events for sensitive operations
			am.checkSecurityEvents(c, userID, auditDetails)
			
			return err
		}
	}
}

// auditResponseWriter captures response data for auditing
type auditResponseWriter struct {
	http.ResponseWriter
	body    *bytes.Buffer
	maxSize int64
}

func (w *auditResponseWriter) Write(data []byte) (int, error) {
	// Capture response body up to max size
	if int64(w.body.Len()) < w.maxSize {
		remainingSpace := w.maxSize - int64(w.body.Len())
		if int64(len(data)) <= remainingSpace {
			w.body.Write(data)
		} else {
			w.body.Write(data[:remainingSpace])
		}
	}
	
	return w.ResponseWriter.Write(data)
}

// sanitizeJSON removes sensitive fields from JSON strings
func (am *AuditMiddleware) sanitizeJSON(jsonStr string) string {
	if jsonStr == "" {
		return ""
	}
	
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		// Not valid JSON, return as-is but truncated if too long
		if len(jsonStr) > 1000 {
			return jsonStr[:1000] + "..."
		}
		return jsonStr
	}
	
	// Sanitize the data
	sanitized := am.sanitizeData(data)
	
	// Marshal back to JSON
	sanitizedBytes, err := json.Marshal(sanitized)
	if err != nil {
		return "[sanitization error]"
	}
	
	return string(sanitizedBytes)
}

// sanitizeData recursively removes sensitive fields from data
func (am *AuditMiddleware) sanitizeData(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		sanitized := make(map[string]interface{})
		for key, value := range v {
			// Check if field is sensitive
			if am.isSensitiveField(key) {
				sanitized[key] = "[REDACTED]"
			} else {
				sanitized[key] = am.sanitizeData(value)
			}
		}
		return sanitized
	case []interface{}:
		sanitized := make([]interface{}, len(v))
		for i, item := range v {
			sanitized[i] = am.sanitizeData(item)
		}
		return sanitized
	default:
		return v
	}
}

// isSensitiveField checks if a field name is considered sensitive
func (am *AuditMiddleware) isSensitiveField(fieldName string) bool {
	fieldLower := strings.ToLower(fieldName)
	for _, sensitive := range am.sensitiveFields {
		if strings.Contains(fieldLower, sensitive) {
			return true
		}
	}
	return false
}

// checkSecurityEvents checks for security-relevant events to log
func (am *AuditMiddleware) checkSecurityEvents(c echo.Context, userID *uuid.UUID, details map[string]interface{}) {
	req := c.Request()
	statusCode := c.Response().Status
	
	// Failed authentication attempts
	if strings.Contains(req.URL.Path, "/auth/") && statusCode == 401 {
		am.auditService.LogSecurityEvent(
			userID,
			"failed_authentication",
			"medium",
			"Failed authentication attempt",
			details,
			req,
		)
	}
	
	// Multiple failed attempts (would need session tracking)
	if statusCode == 429 {
		am.auditService.LogSecurityEvent(
			userID,
			"rate_limit_exceeded",
			"medium",
			"Rate limit exceeded",
			details,
			req,
		)
	}
	
	// Admin operations
	if strings.Contains(req.URL.Path, "/admin/") && statusCode < 400 {
		am.auditService.LogSecurityEvent(
			userID,
			"admin_operation",
			"low",
			"Administrative operation performed",
			details,
			req,
		)
	}
	
	// Data export requests (GDPR)
	if strings.Contains(req.URL.Path, "/gdpr/export") && req.Method == "POST" {
		am.auditService.LogSecurityEvent(
			userID,
			"data_export_request",
			"low",
			"Data export requested",
			details,
			req,
		)
	}
	
	// Data deletion requests (GDPR)
	if strings.Contains(req.URL.Path, "/gdpr/delete") && req.Method == "POST" {
		am.auditService.LogSecurityEvent(
			userID,
			"data_deletion_request",
			"high",
			"Data deletion requested",
			details,
			req,
		)
	}
	
	// Suspicious activity patterns
	userAgent := req.UserAgent()
	if strings.Contains(strings.ToLower(userAgent), "bot") ||
		strings.Contains(strings.ToLower(userAgent), "crawler") ||
		strings.Contains(strings.ToLower(userAgent), "spider") {
		am.auditService.LogSecurityEvent(
			userID,
			"bot_activity",
			"low",
			"Bot-like user agent detected",
			details,
			req,
		)
	}
	
	// Unusual HTTP methods
	if req.Method != "GET" && req.Method != "POST" && req.Method != "PUT" && req.Method != "DELETE" {
		am.auditService.LogSecurityEvent(
			userID,
			"unusual_http_method",
			"low",
			fmt.Sprintf("Unusual HTTP method: %s", req.Method),
			details,
			req,
		)
	}
}

// SetEnabledPaths sets which paths to audit (empty = audit all)
func (am *AuditMiddleware) SetEnabledPaths(paths []string) {
	am.enabledPaths = paths
}

// AddSensitiveFields adds fields to the sensitive fields list
func (am *AuditMiddleware) AddSensitiveFields(fields ...string) {
	am.sensitiveFields = append(am.sensitiveFields, fields...)
}

// SetBodyAuditingOptions configures request/response body auditing
func (am *AuditMiddleware) SetBodyAuditingOptions(auditRequest, auditResponse bool, maxSize int64) {
	am.auditRequestBody = auditRequest
	am.auditResponseBody = auditResponse
	am.maxBodySize = maxSize
}

// UserActionMiddleware specifically tracks user actions on resources
func (am *AuditMiddleware) UserActionMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get user ID from context
			var userID *uuid.UUID
			if user := c.Get("user"); user != nil {
				if userUUID, ok := user.(uuid.UUID); ok {
					userID = &userUUID
				}
			}
			
			// Skip if no authenticated user
			if userID == nil {
				return next(c)
			}
			
			req := c.Request()
			path := req.URL.Path
			method := req.Method
			
			// Determine resource type and action
			resourceType, resourceID, action := am.parseResourceAction(path, method)
			
			if resourceType != "" && action != "" {
				// Execute handler first to get status
				err := next(c)
				
				// Log user action
				success := err == nil && c.Response().Status < 400
				details := map[string]interface{}{
					"method":      method,
					"path":        path,
					"status_code": c.Response().Status,
				}
				
				if success {
					am.auditService.LogAction(userID, action, resourceType, resourceID, details, req)
				} else {
					errorMsg := ""
					if err != nil {
						errorMsg = err.Error()
					}
					am.auditService.LogFailedAction(userID, action, resourceType, resourceID, errorMsg, req)
				}
				
				return err
			}
			
			return next(c)
		}
	}
}

// parseResourceAction extracts resource type, ID, and action from path and method
func (am *AuditMiddleware) parseResourceAction(path, method string) (resourceType, resourceID, action string) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	
	// Parse common REST patterns
	switch {
	case len(parts) >= 2 && parts[0] == "api" && parts[1] == "v1":
		if len(parts) >= 3 {
			resourceType = parts[2]
			
			switch method {
			case "GET":
				if len(parts) >= 4 {
					resourceID = parts[3]
					action = "view_" + resourceType
				} else {
					action = "list_" + resourceType
				}
			case "POST":
				action = "create_" + resourceType
			case "PUT", "PATCH":
				if len(parts) >= 4 {
					resourceID = parts[3]
					action = "update_" + resourceType
				}
			case "DELETE":
				if len(parts) >= 4 {
					resourceID = parts[3]
					action = "delete_" + resourceType
				}
			}
		}
	}
	
	return resourceType, resourceID, action
}