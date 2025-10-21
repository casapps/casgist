package webhooks

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// FilterRule represents a single webhook filter rule
type FilterRule struct {
	Field    string      `json:"field"`    // Field to match against (e.g., "gist.visibility", "user.username")
	Operator string      `json:"operator"` // Comparison operator (eq, ne, contains, regex, etc.)
	Value    interface{} `json:"value"`    // Value to compare against
	Negate   bool        `json:"negate"`   // Whether to negate the result
}

// FilterGroup represents a group of filter rules with logical operations
type FilterGroup struct {
	Rules    []FilterRule  `json:"rules"`    // List of rules in this group
	Groups   []FilterGroup `json:"groups"`   // Nested groups for complex logic
	Logic    string        `json:"logic"`    // "and" or "or" for combining rules/groups
	Disabled bool          `json:"disabled"` // Whether this group is disabled
}

// WebhookFilter contains filtering configuration for webhooks
type WebhookFilter struct {
	ID          uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	WebhookID   uuid.UUID   `json:"webhook_id" gorm:"type:uuid;not null;index"`
	Name        string      `json:"name" gorm:"size:255;not null"`
	Description string      `json:"description" gorm:"type:text"`
	FilterGroup FilterGroup `json:"filter_group" gorm:"type:json"`
	IsActive    bool        `json:"is_active" gorm:"default:true"`
	Priority    int         `json:"priority" gorm:"default:0"` // Higher priority filters are applied first
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// EventFilter evaluates webhook events against filter rules
type EventFilter struct {
	filters map[uuid.UUID][]WebhookFilter
}

// NewEventFilter creates a new event filter
func NewEventFilter() *EventFilter {
	return &EventFilter{
		filters: make(map[uuid.UUID][]WebhookFilter),
	}
}

// AddFilter adds a filter for a webhook
func (ef *EventFilter) AddFilter(webhookID uuid.UUID, filter WebhookFilter) {
	if ef.filters[webhookID] == nil {
		ef.filters[webhookID] = []WebhookFilter{}
	}
	ef.filters[webhookID] = append(ef.filters[webhookID], filter)
}

// RemoveFilter removes a filter for a webhook
func (ef *EventFilter) RemoveFilter(webhookID uuid.UUID, filterID uuid.UUID) {
	filters := ef.filters[webhookID]
	for i, filter := range filters {
		if filter.ID == filterID {
			ef.filters[webhookID] = append(filters[:i], filters[i+1:]...)
			break
		}
	}
}

// ShouldDeliver determines if an event should be delivered to a webhook
func (ef *EventFilter) ShouldDeliver(webhookID uuid.UUID, event *Event) bool {
	filters, exists := ef.filters[webhookID]
	if !exists || len(filters) == 0 {
		return true // No filters means deliver all events
	}
	
	// Apply each active filter
	for _, filter := range filters {
		if !filter.IsActive {
			continue
		}
		
		// If any filter passes, deliver the event
		if ef.evaluateFilterGroup(filter.FilterGroup, event) {
			return true
		}
	}
	
	// No filters passed, don't deliver
	return false
}

// evaluateFilterGroup evaluates a filter group against an event
func (ef *EventFilter) evaluateFilterGroup(group FilterGroup, event *Event) bool {
	if group.Disabled {
		return false
	}
	
	results := []bool{}
	
	// Evaluate rules
	for _, rule := range group.Rules {
		result := ef.evaluateRule(rule, event)
		if rule.Negate {
			result = !result
		}
		results = append(results, result)
	}
	
	// Evaluate nested groups
	for _, nestedGroup := range group.Groups {
		result := ef.evaluateFilterGroup(nestedGroup, event)
		results = append(results, result)
	}
	
	// Apply logic operator
	if len(results) == 0 {
		return true // Empty group passes
	}
	
	switch strings.ToLower(group.Logic) {
	case "or":
		// OR: at least one result must be true
		for _, result := range results {
			if result {
				return true
			}
		}
		return false
	case "and", "":
		// AND: all results must be true (default)
		for _, result := range results {
			if !result {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// evaluateRule evaluates a single filter rule against an event
func (ef *EventFilter) evaluateRule(rule FilterRule, event *Event) bool {
	// Extract field value from event
	fieldValue, err := ef.extractFieldValue(rule.Field, event)
	if err != nil {
		return false // Field not found or invalid
	}
	
	// Apply operator
	return ef.applyOperator(rule.Operator, fieldValue, rule.Value)
}

// extractFieldValue extracts a field value from an event using dot notation
func (ef *EventFilter) extractFieldValue(fieldPath string, event *Event) (interface{}, error) {
	// Convert event to JSON for easy field extraction
	eventData, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	
	var eventMap map[string]interface{}
	if err := json.Unmarshal(eventData, &eventMap); err != nil {
		return nil, err
	}
	
	// Navigate through the field path
	parts := strings.Split(fieldPath, ".")
	current := eventMap
	
	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - return the value
			return current[part], nil
		}
		
		// Navigate deeper
		next, exists := current[part]
		if !exists {
			return nil, fmt.Errorf("field not found: %s", part)
		}
		
		nextMap, ok := next.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot navigate further into field: %s", part)
		}
		
		current = nextMap
	}
	
	return nil, fmt.Errorf("invalid field path: %s", fieldPath)
}

// applyOperator applies a comparison operator
func (ef *EventFilter) applyOperator(operator string, fieldValue, ruleValue interface{}) bool {
	switch strings.ToLower(operator) {
	case "eq", "equals", "=":
		return ef.equals(fieldValue, ruleValue)
	case "ne", "not_equals", "!=":
		return !ef.equals(fieldValue, ruleValue)
	case "contains":
		return ef.contains(fieldValue, ruleValue)
	case "not_contains":
		return !ef.contains(fieldValue, ruleValue)
	case "starts_with":
		return ef.startsWith(fieldValue, ruleValue)
	case "ends_with":
		return ef.endsWith(fieldValue, ruleValue)
	case "regex":
		return ef.matchesRegex(fieldValue, ruleValue)
	case "gt", ">":
		return ef.greaterThan(fieldValue, ruleValue)
	case "gte", ">=":
		return ef.greaterThanOrEqual(fieldValue, ruleValue)
	case "lt", "<":
		return ef.lessThan(fieldValue, ruleValue)
	case "lte", "<=":
		return ef.lessThanOrEqual(fieldValue, ruleValue)
	case "in":
		return ef.in(fieldValue, ruleValue)
	case "not_in":
		return !ef.in(fieldValue, ruleValue)
	case "exists":
		return fieldValue != nil
	case "not_exists":
		return fieldValue == nil
	default:
		return false
	}
}

// Comparison helper methods

func (ef *EventFilter) equals(a, b interface{}) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func (ef *EventFilter) contains(field, value interface{}) bool {
	fieldStr := fmt.Sprintf("%v", field)
	valueStr := fmt.Sprintf("%v", value)
	return strings.Contains(fieldStr, valueStr)
}

func (ef *EventFilter) startsWith(field, value interface{}) bool {
	fieldStr := fmt.Sprintf("%v", field)
	valueStr := fmt.Sprintf("%v", value)
	return strings.HasPrefix(fieldStr, valueStr)
}

func (ef *EventFilter) endsWith(field, value interface{}) bool {
	fieldStr := fmt.Sprintf("%v", field)
	valueStr := fmt.Sprintf("%v", value)
	return strings.HasSuffix(fieldStr, valueStr)
}

func (ef *EventFilter) matchesRegex(field, pattern interface{}) bool {
	fieldStr := fmt.Sprintf("%v", field)
	patternStr := fmt.Sprintf("%v", pattern)
	
	regex, err := regexp.Compile(patternStr)
	if err != nil {
		return false
	}
	
	return regex.MatchString(fieldStr)
}

func (ef *EventFilter) greaterThan(field, value interface{}) bool {
	return ef.compareNumbers(field, value) > 0
}

func (ef *EventFilter) greaterThanOrEqual(field, value interface{}) bool {
	return ef.compareNumbers(field, value) >= 0
}

func (ef *EventFilter) lessThan(field, value interface{}) bool {
	return ef.compareNumbers(field, value) < 0
}

func (ef *EventFilter) lessThanOrEqual(field, value interface{}) bool {
	return ef.compareNumbers(field, value) <= 0
}

func (ef *EventFilter) compareNumbers(a, b interface{}) int {
	// Simple numeric comparison - convert to float64
	aFloat, aOk := ef.toFloat64(a)
	bFloat, bOk := ef.toFloat64(b)
	
	if !aOk || !bOk {
		// Fall back to string comparison
		aStr := fmt.Sprintf("%v", a)
		bStr := fmt.Sprintf("%v", b)
		return strings.Compare(aStr, bStr)
	}
	
	if aFloat < bFloat {
		return -1
	} else if aFloat > bFloat {
		return 1
	}
	return 0
}

func (ef *EventFilter) toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case float32:
		return float64(val), true
	case float64:
		return val, true
	default:
		return 0, false
	}
}

func (ef *EventFilter) in(field, values interface{}) bool {
	// Check if field value is in the list of values
	fieldStr := fmt.Sprintf("%v", field)
	
	// Try to parse values as array
	valuesBytes, err := json.Marshal(values)
	if err != nil {
		return false
	}
	
	var valuesArray []interface{}
	if err := json.Unmarshal(valuesBytes, &valuesArray); err != nil {
		return false
	}
	
	for _, value := range valuesArray {
		if fieldStr == fmt.Sprintf("%v", value) {
			return true
		}
	}
	
	return false
}