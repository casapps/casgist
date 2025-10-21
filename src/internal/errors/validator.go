package errors

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Validator provides comprehensive input validation
type Validator struct {
	errors []ValidationError
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
	Code    string      `json:"code"`
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		errors: make([]ValidationError, 0),
	}
}

// HasErrors returns true if there are validation errors
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// GetErrors returns all validation errors
func (v *Validator) GetErrors() []ValidationError {
	return v.errors
}

// GetErrorsMap returns validation errors as a map
func (v *Validator) GetErrorsMap() map[string]interface{} {
	errorMap := make(map[string]interface{})
	
	for _, err := range v.errors {
		if existing, exists := errorMap[err.Field]; exists {
			// Multiple errors for same field
			if existingSlice, ok := existing.([]string); ok {
				errorMap[err.Field] = append(existingSlice, err.Message)
			} else {
				errorMap[err.Field] = []string{existing.(string), err.Message}
			}
		} else {
			errorMap[err.Field] = err.Message
		}
	}
	
	return errorMap
}

// Clear clears all validation errors
func (v *Validator) Clear() {
	v.errors = make([]ValidationError, 0)
}

// AddError adds a validation error
func (v *Validator) AddError(field, message, code string, value interface{}) *Validator {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
		Code:    code,
	})
	return v
}

// Required validates that a field is not empty
func (v *Validator) Required(field string, value interface{}) *Validator {
	if isEmpty(value) {
		v.AddError(field, fmt.Sprintf("%s is required", field), "REQUIRED", value)
	}
	return v
}

// Email validates email format
func (v *Validator) Email(field, value string) *Validator {
	if value == "" {
		return v // Skip validation for empty values
	}
	
	if _, err := mail.ParseAddress(value); err != nil {
		v.AddError(field, fmt.Sprintf("%s must be a valid email address", field), "INVALID_EMAIL", value)
	}
	return v
}

// MinLength validates minimum string length
func (v *Validator) MinLength(field, value string, min int) *Validator {
	if value == "" {
		return v // Skip validation for empty values
	}
	
	if len(value) < min {
		v.AddError(field, fmt.Sprintf("%s must be at least %d characters long", field, min), "MIN_LENGTH", value)
	}
	return v
}

// MaxLength validates maximum string length
func (v *Validator) MaxLength(field, value string, max int) *Validator {
	if value == "" {
		return v // Skip validation for empty values
	}
	
	if len(value) > max {
		v.AddError(field, fmt.Sprintf("%s must not exceed %d characters", field, max), "MAX_LENGTH", value)
	}
	return v
}

// Pattern validates string against regex pattern
func (v *Validator) Pattern(field, value, pattern, message string) *Validator {
	if value == "" {
		return v // Skip validation for empty values
	}
	
	matched, err := regexp.MatchString(pattern, value)
	if err != nil {
		v.AddError(field, fmt.Sprintf("Invalid pattern for %s", field), "INVALID_PATTERN", value)
		return v
	}
	
	if !matched {
		if message == "" {
			message = fmt.Sprintf("%s format is invalid", field)
		}
		v.AddError(field, message, "PATTERN_MISMATCH", value)
	}
	return v
}

// Username validates username format
func (v *Validator) Username(field, value string) *Validator {
	if value == "" {
		return v // Skip validation for empty values
	}
	
	// Username rules: 3-30 characters, alphanumeric + underscore/hyphen, no consecutive special chars
	if len(value) < 3 || len(value) > 30 {
		v.AddError(field, "Username must be between 3 and 30 characters", "INVALID_LENGTH", value)
		return v
	}
	
	// Must start and end with alphanumeric
	if !isAlphanumeric(rune(value[0])) || !isAlphanumeric(rune(value[len(value)-1])) {
		v.AddError(field, "Username must start and end with a letter or number", "INVALID_FORMAT", value)
		return v
	}
	
	// Check allowed characters
	prev := rune(0)
	for i, r := range value {
		if !isAlphanumeric(r) && r != '-' && r != '_' {
			v.AddError(field, "Username can only contain letters, numbers, hyphens, and underscores", "INVALID_CHARACTER", value)
			return v
		}
		
		// No consecutive special characters
		if i > 0 && isSpecialChar(r) && isSpecialChar(prev) {
			v.AddError(field, "Username cannot have consecutive hyphens or underscores", "INVALID_FORMAT", value)
			return v
		}
		prev = r
	}
	
	return v
}

// Password validates password strength
func (v *Validator) Password(field, value string) *Validator {
	if value == "" {
		return v // Skip validation for empty values
	}
	
	// Minimum length
	if len(value) < 8 {
		v.AddError(field, "Password must be at least 8 characters long", "MIN_LENGTH", nil)
	}
	
	// Maximum length (for bcrypt compatibility)
	if len(value) > 72 {
		v.AddError(field, "Password must not exceed 72 characters", "MAX_LENGTH", nil)
	}
	
	// Check for required character types
	var hasLower, hasUpper, hasDigit, hasSpecial bool
	
	for _, r := range value {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}
	
	missing := []string{}
	if !hasLower {
		missing = append(missing, "lowercase letter")
	}
	if !hasUpper {
		missing = append(missing, "uppercase letter")
	}
	if !hasDigit {
		missing = append(missing, "number")
	}
	if !hasSpecial {
		missing = append(missing, "special character")
	}
	
	if len(missing) > 0 {
		v.AddError(field, fmt.Sprintf("Password must contain at least one %s", strings.Join(missing, ", ")), "WEAK_PASSWORD", nil)
	}
	
	// Check for common weak patterns
	lower := strings.ToLower(value)
	weakPatterns := []string{
		"password", "123456", "qwerty", "abc123", "admin", "user", "test",
		"welcome", "login", "guest", "root", "master", "default",
	}
	
	for _, pattern := range weakPatterns {
		if strings.Contains(lower, pattern) {
			v.AddError(field, "Password contains common weak patterns", "WEAK_PASSWORD", nil)
			break
		}
	}
	
	return v
}

// URL validates URL format
func (v *Validator) URL(field, value string) *Validator {
	if value == "" {
		return v // Skip validation for empty values
	}
	
	urlPattern := `^https?://[^\s/$.?#].[^\s]*$`
	return v.Pattern(field, value, urlPattern, "Invalid URL format")
}

// Integer validates integer value and range
func (v *Validator) Integer(field string, value interface{}, min, max *int) *Validator {
	var intValue int
	var err error
	
	switch val := value.(type) {
	case int:
		intValue = val
	case string:
		if val == "" {
			return v // Skip validation for empty values
		}
		intValue, err = strconv.Atoi(val)
		if err != nil {
			v.AddError(field, fmt.Sprintf("%s must be a valid integer", field), "INVALID_INTEGER", value)
			return v
		}
	default:
		v.AddError(field, fmt.Sprintf("%s must be an integer", field), "INVALID_TYPE", value)
		return v
	}
	
	if min != nil && intValue < *min {
		v.AddError(field, fmt.Sprintf("%s must be at least %d", field, *min), "MIN_VALUE", value)
	}
	
	if max != nil && intValue > *max {
		v.AddError(field, fmt.Sprintf("%s must not exceed %d", field, *max), "MAX_VALUE", value)
	}
	
	return v
}

// Float validates float value and range
func (v *Validator) Float(field string, value interface{}, min, max *float64) *Validator {
	var floatValue float64
	var err error
	
	switch val := value.(type) {
	case float64:
		floatValue = val
	case float32:
		floatValue = float64(val)
	case int:
		floatValue = float64(val)
	case string:
		if val == "" {
			return v // Skip validation for empty values
		}
		floatValue, err = strconv.ParseFloat(val, 64)
		if err != nil {
			v.AddError(field, fmt.Sprintf("%s must be a valid number", field), "INVALID_NUMBER", value)
			return v
		}
	default:
		v.AddError(field, fmt.Sprintf("%s must be a number", field), "INVALID_TYPE", value)
		return v
	}
	
	if min != nil && floatValue < *min {
		v.AddError(field, fmt.Sprintf("%s must be at least %v", field, *min), "MIN_VALUE", value)
	}
	
	if max != nil && floatValue > *max {
		v.AddError(field, fmt.Sprintf("%s must not exceed %v", field, *max), "MAX_VALUE", value)
	}
	
	return v
}

// OneOf validates that value is one of allowed values
func (v *Validator) OneOf(field string, value interface{}, allowed []interface{}) *Validator {
	if isEmpty(value) {
		return v // Skip validation for empty values
	}
	
	for _, allowedValue := range allowed {
		if reflect.DeepEqual(value, allowedValue) {
			return v
		}
	}
	
	v.AddError(field, fmt.Sprintf("%s must be one of: %v", field, allowed), "INVALID_CHOICE", value)
	return v
}

// Date validates date format and range
func (v *Validator) Date(field, value, layout string, after, before *time.Time) *Validator {
	if value == "" {
		return v // Skip validation for empty values
	}
	
	parsedTime, err := time.Parse(layout, value)
	if err != nil {
		v.AddError(field, fmt.Sprintf("%s must be a valid date in format %s", field, layout), "INVALID_DATE", value)
		return v
	}
	
	if after != nil && parsedTime.Before(*after) {
		v.AddError(field, fmt.Sprintf("%s must be after %s", field, after.Format(layout)), "DATE_TOO_EARLY", value)
	}
	
	if before != nil && parsedTime.After(*before) {
		v.AddError(field, fmt.Sprintf("%s must be before %s", field, before.Format(layout)), "DATE_TOO_LATE", value)
	}
	
	return v
}

// JSON validates JSON format
func (v *Validator) JSON(field, value string) *Validator {
	if value == "" {
		return v // Skip validation for empty values
	}
	
	var js interface{}
	if err := json.Unmarshal([]byte(value), &js); err != nil {
		v.AddError(field, fmt.Sprintf("%s must be valid JSON", field), "INVALID_JSON", value)
	}
	
	return v
}

// GistVisibility validates gist visibility
func (v *Validator) GistVisibility(field, value string) *Validator {
	allowed := []interface{}{"public", "unlisted", "private"}
	return v.OneOf(field, value, allowed)
}

// GistTitle validates gist title
func (v *Validator) GistTitle(field, value string) *Validator {
	return v.Required(field, value).
		MinLength(field, value, 1).
		MaxLength(field, value, 200)
}

// GistDescription validates gist description
func (v *Validator) GistDescription(field, value string) *Validator {
	return v.MaxLength(field, value, 1000)
}

// FileName validates file name
func (v *Validator) FileName(field, value string) *Validator {
	if value == "" {
		return v // Skip validation for empty values
	}
	
	// Check length
	if len(value) > 255 {
		v.AddError(field, "Filename must not exceed 255 characters", "MAX_LENGTH", value)
	}
	
	// Check for invalid characters (Windows + Unix)
	invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*", "\x00"}
	for _, char := range invalidChars {
		if strings.Contains(value, char) {
			v.AddError(field, "Filename contains invalid characters", "INVALID_CHARACTER", value)
			break
		}
	}
	
	// Check for reserved names (Windows)
	reserved := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}
	
	upperName := strings.ToUpper(strings.TrimSuffix(value, filepath.Ext(value)))
	for _, res := range reserved {
		if upperName == res {
			v.AddError(field, "Filename uses reserved name", "RESERVED_NAME", value)
			break
		}
	}
	
	// Cannot start or end with dot or space
	if strings.HasPrefix(value, ".") || strings.HasPrefix(value, " ") ||
	   strings.HasSuffix(value, ".") || strings.HasSuffix(value, " ") {
		v.AddError(field, "Filename cannot start or end with dot or space", "INVALID_FORMAT", value)
	}
	
	return v
}

// CreateValidationError creates a validation error response
func (v *Validator) CreateValidationError() *CustomError {
	if !v.HasErrors() {
		return nil
	}
	
	return NewValidationError("Validation failed", "").
		WithDetail("validation_errors", v.GetErrors()).
		WithDetail("error_count", len(v.errors))
}

// Helper functions

func isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}
	
	switch val := value.(type) {
	case string:
		return strings.TrimSpace(val) == ""
	case *string:
		return val == nil || strings.TrimSpace(*val) == ""
	case []interface{}:
		return len(val) == 0
	case map[string]interface{}:
		return len(val) == 0
	}
	
	// Use reflection for other types
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String, reflect.Array, reflect.Map, reflect.Slice:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	
	return false
}

func isAlphanumeric(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isSpecialChar(r rune) bool {
	return r == '-' || r == '_'
}