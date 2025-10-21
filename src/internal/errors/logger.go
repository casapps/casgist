package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// ErrorLogger handles comprehensive error logging
type ErrorLogger struct {
	config     *viper.Viper
	db         *gorm.DB
	fileLogger *log.Logger
	logFile    *os.File
	mu         sync.RWMutex
	stats      *ErrorStats
}

// ErrorStats tracks error statistics
type ErrorStats struct {
	mu                sync.RWMutex
	TotalErrors       int64                    `json:"total_errors"`
	ErrorsByType      map[ErrorType]int64      `json:"errors_by_type"`
	ErrorsByStatus    map[int]int64            `json:"errors_by_status"`
	ErrorsByPath      map[string]int64         `json:"errors_by_path"`
	RecentErrors      []ErrorLogEntry          `json:"recent_errors"`
	StartTime         time.Time                `json:"start_time"`
	LastError         *ErrorLogEntry           `json:"last_error,omitempty"`
}

// ErrorLogEntry represents a logged error
type ErrorLogEntry struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	Timestamp   time.Time              `json:"timestamp" gorm:"index"`
	ErrorType   string                 `json:"error_type" gorm:"index"`
	Message     string                 `json:"message"`
	Code        string                 `json:"code"`
	StatusCode  int                    `json:"status_code" gorm:"index"`
	Path        string                 `json:"path" gorm:"index"`
	Method      string                 `json:"method"`
	UserID      *string                `json:"user_id,omitempty" gorm:"index"`
	RequestID   string                 `json:"request_id" gorm:"index"`
	UserAgent   string                 `json:"user_agent"`
	IP          string                 `json:"ip" gorm:"index"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
	Context     string                 `json:"context,omitempty"` // JSON encoded context
	Resolved    bool                   `json:"resolved" gorm:"default:false;index"`
	Resolution  string                 `json:"resolution,omitempty"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	ResolvedBy  *string                `json:"resolved_by,omitempty"`
	Severity    string                 `json:"severity" gorm:"default:'medium';index"`
	Tags        string                 `json:"tags,omitempty"` // Comma-separated tags
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// NewErrorLogger creates a new error logger
func NewErrorLogger(config *viper.Viper, db *gorm.DB) ErrorLogger {
	logger := ErrorLogger{
		config: config,
		db:     db,
		stats: &ErrorStats{
			ErrorsByType:   make(map[ErrorType]int64),
			ErrorsByStatus: make(map[int]int64),
			ErrorsByPath:   make(map[string]int64),
			RecentErrors:   make([]ErrorLogEntry, 0, 100),
			StartTime:      time.Now(),
		},
	}
	
	logger.initializeFileLogger()
	logger.initializeDatabase()
	
	return logger
}

// initializeFileLogger sets up file-based logging
func (l *ErrorLogger) initializeFileLogger() {
	logDir := l.config.GetString("logging.directory")
	if logDir == "" {
		logDir = "logs"
	}
	
	// Create log directory
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Failed to create log directory: %v", err)
		return
	}
	
	// Open log file
	logPath := filepath.Join(logDir, "errors.log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Failed to open error log file: %v", err)
		return
	}
	
	l.logFile = file
	l.fileLogger = log.New(file, "", log.LstdFlags|log.Lmicroseconds)
}

// initializeDatabase sets up database logging
func (l *ErrorLogger) initializeDatabase() {
	if l.db != nil {
		// Auto-migrate error log table
		if err := l.db.AutoMigrate(&ErrorLogEntry{}); err != nil {
			log.Printf("Failed to migrate error log table: %v", err)
		}
	}
}

// LogError logs an error with context
func (l *ErrorLogger) LogError(ctx context.Context, err error, context map[string]interface{}) {
	entry := l.createErrorLogEntry(err, context)
	
	// Update statistics
	l.updateStats(entry)
	
	// Log to file
	l.logToFile(entry)
	
	// Log to database
	l.logToDatabase(entry)
	
	// Log to console in development
	if l.config.GetString("environment") != "production" {
		l.logToConsole(entry)
	}
}

// createErrorLogEntry creates an error log entry
func (l *ErrorLogger) createErrorLogEntry(err error, context map[string]interface{}) ErrorLogEntry {
	entry := ErrorLogEntry{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Message:   err.Error(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// Extract information from custom errors
	if customErr, ok := err.(*CustomError); ok {
		entry.ErrorType = string(customErr.Type)
		entry.Code = customErr.Code
		entry.StatusCode = customErr.StatusCode
		entry.Severity = l.determineSeverity(customErr)
	} else {
		entry.ErrorType = string(ErrorTypeServer)
		entry.Code = "UNKNOWN_ERROR"
		entry.StatusCode = 500
		entry.Severity = "high"
	}
	
	// Add context information
	if context != nil {
		if path, ok := context["path"].(string); ok {
			entry.Path = path
		}
		if method, ok := context["method"].(string); ok {
			entry.Method = method
		}
		if requestID, ok := context["request_id"].(string); ok {
			entry.RequestID = requestID
		}
		if userAgent, ok := context["user_agent"].(string); ok {
			entry.UserAgent = userAgent
		}
		if ip, ok := context["ip"].(string); ok {
			entry.IP = ip
		}
		if userID, ok := context["user_id"].(string); ok {
			entry.UserID = &userID
		}
		if stack, ok := context["stack"].(string); ok {
			entry.StackTrace = stack
		}
		if panicStack, ok := context["panic_stack"].(string); ok {
			entry.StackTrace = panicStack
			entry.Severity = "critical"
		}
		
		// Encode remaining context as JSON
		contextData := make(map[string]interface{})
		for k, v := range context {
			if !l.isStandardField(k) {
				contextData[k] = v
			}
		}
		
		if len(contextData) > 0 {
			if contextJSON, err := json.Marshal(contextData); err == nil {
				entry.Context = string(contextJSON)
			}
		}
	}
	
	// Add stack trace if not already present
	if entry.StackTrace == "" && entry.Severity == "high" {
		entry.StackTrace = l.getStackTrace(3) // Skip LogError, createErrorLogEntry, and getStackTrace
	}
	
	return entry
}

// isStandardField checks if a context field is already handled
func (l *ErrorLogger) isStandardField(field string) bool {
	standardFields := []string{
		"path", "method", "request_id", "user_agent", "ip", "user_id", "stack", "panic_stack",
	}
	
	for _, std := range standardFields {
		if field == std {
			return true
		}
	}
	return false
}

// determineSeverity determines error severity
func (l *ErrorLogger) determineSeverity(err *CustomError) string {
	switch err.Type {
	case ErrorTypeValidation, ErrorTypeNotFound:
		return "low"
	case ErrorTypeAuthentication, ErrorTypeAuthorization, ErrorTypeRateLimit:
		return "medium"
	case ErrorTypeDatabase, ErrorTypeServer, ErrorTypeNetwork, ErrorTypeExternal:
		return "high"
	case ErrorTypeTimeout, ErrorTypeStorage, ErrorTypeConfig:
		return "high"
	default:
		return "medium"
	}
}

// getStackTrace gets current stack trace
func (l *ErrorLogger) getStackTrace(skip int) string {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return string(buf[:n])
		}
		buf = make([]byte, 2*len(buf))
	}
}

// updateStats updates error statistics
func (l *ErrorLogger) updateStats(entry ErrorLogEntry) {
	l.stats.mu.Lock()
	defer l.stats.mu.Unlock()
	
	l.stats.TotalErrors++
	
	// Update error type stats
	errorType := ErrorType(entry.ErrorType)
	l.stats.ErrorsByType[errorType]++
	
	// Update status code stats
	l.stats.ErrorsByStatus[entry.StatusCode]++
	
	// Update path stats
	if entry.Path != "" {
		l.stats.ErrorsByPath[entry.Path]++
	}
	
	// Update recent errors (keep last 100)
	l.stats.RecentErrors = append(l.stats.RecentErrors, entry)
	if len(l.stats.RecentErrors) > 100 {
		l.stats.RecentErrors = l.stats.RecentErrors[1:]
	}
	
	// Update last error
	entryCopy := entry
	l.stats.LastError = &entryCopy
}

// logToFile logs to file
func (l *ErrorLogger) logToFile(entry ErrorLogEntry) {
	if l.fileLogger == nil {
		return
	}
	
	logLine := fmt.Sprintf("[%s] %s %s %d %s %s - %s",
		entry.Severity,
		entry.Method,
		entry.Path,
		entry.StatusCode,
		entry.Code,
		entry.RequestID,
		entry.Message,
	)
	
	if entry.StackTrace != "" {
		logLine += fmt.Sprintf("\nStack: %s", entry.StackTrace)
	}
	
	l.fileLogger.Println(logLine)
}

// logToDatabase logs to database
func (l *ErrorLogger) logToDatabase(entry ErrorLogEntry) {
	if l.db == nil {
		return
	}
	
	// Log to database in background to avoid blocking
	go func() {
		if err := l.db.Create(&entry).Error; err != nil {
			log.Printf("Failed to log error to database: %v", err)
		}
	}()
}

// logToConsole logs to console
func (l *ErrorLogger) logToConsole(entry ErrorLogEntry) {
	color := l.getSeverityColor(entry.Severity)
	reset := "\033[0m"
	
	fmt.Printf("%s[%s] %s %s %d %s%s\n",
		color,
		entry.Severity,
		entry.Method,
		entry.Path,
		entry.StatusCode,
		entry.Message,
		reset,
	)
	
	if entry.StackTrace != "" && entry.Severity == "high" {
		fmt.Printf("Stack trace:\n%s\n", entry.StackTrace)
	}
}

// getSeverityColor returns ANSI color code for severity
func (l *ErrorLogger) getSeverityColor(severity string) string {
	switch severity {
	case "low":
		return "\033[32m" // Green
	case "medium":
		return "\033[33m" // Yellow
	case "high":
		return "\033[31m" // Red
	case "critical":
		return "\033[35m" // Magenta
	default:
		return "\033[37m" // White
	}
}

// GetErrorStats returns error statistics
func (l *ErrorLogger) GetErrorStats() (map[string]interface{}, error) {
	l.stats.mu.RLock()
	defer l.stats.mu.RUnlock()
	
	uptime := time.Since(l.stats.StartTime)
	
	stats := map[string]interface{}{
		"total_errors":     l.stats.TotalErrors,
		"errors_by_type":   l.stats.ErrorsByType,
		"errors_by_status": l.stats.ErrorsByStatus,
		"errors_by_path":   l.stats.ErrorsByPath,
		"recent_errors":    len(l.stats.RecentErrors),
		"uptime":          uptime.String(),
		"uptime_seconds":  int64(uptime.Seconds()),
		"start_time":      l.stats.StartTime,
		"last_error":      l.stats.LastError,
		"error_rate":      l.calculateErrorRate(uptime),
	}
	
	return stats, nil
}

// calculateErrorRate calculates errors per hour
func (l *ErrorLogger) calculateErrorRate(uptime time.Duration) float64 {
	hours := uptime.Hours()
	if hours < 1 {
		hours = 1 // Avoid division by zero
	}
	return float64(l.stats.TotalErrors) / hours
}

// GetRecentErrors returns recent errors with pagination
func (l *ErrorLogger) GetRecentErrors(limit, offset int) ([]ErrorLogEntry, int64, error) {
	if l.db == nil {
		l.stats.mu.RLock()
		defer l.stats.mu.RUnlock()
		
		total := int64(len(l.stats.RecentErrors))
		
		// Simple in-memory pagination
		start := offset
		end := offset + limit
		if start > len(l.stats.RecentErrors) {
			return []ErrorLogEntry{}, total, nil
		}
		if end > len(l.stats.RecentErrors) {
			end = len(l.stats.RecentErrors)
		}
		
		return l.stats.RecentErrors[start:end], total, nil
	}
	
	// Database query
	var errors []ErrorLogEntry
	var total int64
	
	if err := l.db.Model(&ErrorLogEntry{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	if err := l.db.Order("timestamp DESC").Limit(limit).Offset(offset).Find(&errors).Error; err != nil {
		return nil, 0, err
	}
	
	return errors, total, nil
}

// MarkErrorResolved marks an error as resolved
func (l *ErrorLogger) MarkErrorResolved(errorID, resolution, resolvedBy string) error {
	if l.db == nil {
		return fmt.Errorf("database not available")
	}
	
	now := time.Now()
	return l.db.Model(&ErrorLogEntry{}).
		Where("id = ?", errorID).
		Updates(map[string]interface{}{
			"resolved":    true,
			"resolution":  resolution,
			"resolved_by": resolvedBy,
			"resolved_at": now,
			"updated_at":  now,
		}).Error
}

// SearchErrors searches error logs
func (l *ErrorLogger) SearchErrors(query string, filters map[string]interface{}) ([]ErrorLogEntry, error) {
	if l.db == nil {
		return nil, fmt.Errorf("database not available")
	}
	
	tx := l.db.Model(&ErrorLogEntry{})
	
	// Apply search query
	if query != "" {
		tx = tx.Where("message LIKE ? OR code LIKE ? OR path LIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%")
	}
	
	// Apply filters
	if errorType, ok := filters["error_type"].(string); ok && errorType != "" {
		tx = tx.Where("error_type = ?", errorType)
	}
	
	if severity, ok := filters["severity"].(string); ok && severity != "" {
		tx = tx.Where("severity = ?", severity)
	}
	
	if resolved, ok := filters["resolved"].(bool); ok {
		tx = tx.Where("resolved = ?", resolved)
	}
	
	if startTime, ok := filters["start_time"].(time.Time); ok {
		tx = tx.Where("timestamp >= ?", startTime)
	}
	
	if endTime, ok := filters["end_time"].(time.Time); ok {
		tx = tx.Where("timestamp <= ?", endTime)
	}
	
	var errors []ErrorLogEntry
	err := tx.Order("timestamp DESC").Limit(1000).Find(&errors).Error
	return errors, err
}

// Close closes the error logger
func (l *ErrorLogger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}