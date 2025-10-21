package webhooks

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// WebhookMetric contains metrics for a webhook
type WebhookMetric struct {
	WebhookID              uuid.UUID     `json:"webhook_id"`
	Deliveries             int64         `json:"deliveries"`
	Errors                 int64         `json:"errors"`
	Filtered               int64         `json:"filtered"`
	RateLimited            int64         `json:"rate_limited"`
	CircuitBreakerOpen     int64         `json:"circuit_breaker_open"`
	TotalLatency           time.Duration `json:"total_latency"`
	AverageLatency         time.Duration `json:"average_latency"`
	MinLatency             time.Duration `json:"min_latency"`
	MaxLatency             time.Duration `json:"max_latency"`
	LastActivity           time.Time     `json:"last_activity"`
	SuccessRate            float64       `json:"success_rate"`
	DeliveryRate           float64       `json:"delivery_rate"` // Deliveries / (Deliveries + Filtered)
	ErrorRate              float64       `json:"error_rate"`
	RecentDeliveries       []time.Time   `json:"recent_deliveries"` // Last 100 deliveries for rate calculation
	HourlyStats            [24]int64     `json:"hourly_stats"`      // Deliveries by hour of day
	DailyStats             [7]int64      `json:"daily_stats"`       // Deliveries by day of week
	StatusCodeCounts       map[int]int64 `json:"status_code_counts"`
}

// WebhookMetrics manages metrics for all webhooks
type WebhookMetrics struct {
	metrics map[uuid.UUID]*WebhookMetric
	mu      sync.RWMutex
}

// NewWebhookMetrics creates a new metrics manager
func NewWebhookMetrics() *WebhookMetrics {
	return &WebhookMetrics{
		metrics: make(map[uuid.UUID]*WebhookMetric),
	}
}

// GetMetrics returns metrics for a webhook
func (wm *WebhookMetrics) GetMetrics(webhookID uuid.UUID) *WebhookMetric {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	
	metric, exists := wm.metrics[webhookID]
	if !exists {
		return nil
	}
	
	// Create a copy to avoid race conditions
	copy := *metric
	copy.StatusCodeCounts = make(map[int]int64)
	for k, v := range metric.StatusCodeCounts {
		copy.StatusCodeCounts[k] = v
	}
	
	// Update calculated fields
	wm.updateCalculatedFields(&copy)
	
	return &copy
}

// GetAllMetrics returns metrics for all webhooks
func (wm *WebhookMetrics) GetAllMetrics() map[uuid.UUID]*WebhookMetric {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	
	result := make(map[uuid.UUID]*WebhookMetric)
	for id, metric := range wm.metrics {
		copy := *metric
		copy.StatusCodeCounts = make(map[int]int64)
		for k, v := range metric.StatusCodeCounts {
			copy.StatusCodeCounts[k] = v
		}
		wm.updateCalculatedFields(&copy)
		result[id] = &copy
	}
	
	return result
}

// IncrementDeliveries increments delivery count
func (wm *WebhookMetrics) IncrementDeliveries(webhookID uuid.UUID) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	
	metric := wm.getOrCreateMetric(webhookID)
	metric.Deliveries++
	metric.LastActivity = time.Now()
	
	// Update recent deliveries (keep last 100)
	metric.RecentDeliveries = append(metric.RecentDeliveries, time.Now())
	if len(metric.RecentDeliveries) > 100 {
		metric.RecentDeliveries = metric.RecentDeliveries[1:]
	}
	
	// Update hourly and daily stats
	now := time.Now()
	metric.HourlyStats[now.Hour()]++
	metric.DailyStats[now.Weekday()]++
}

// IncrementErrors increments error count
func (wm *WebhookMetrics) IncrementErrors(webhookID uuid.UUID) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	
	metric := wm.getOrCreateMetric(webhookID)
	metric.Errors++
	metric.LastActivity = time.Now()
}

// IncrementFiltered increments filtered count
func (wm *WebhookMetrics) IncrementFiltered(webhookID uuid.UUID) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	
	metric := wm.getOrCreateMetric(webhookID)
	metric.Filtered++
	metric.LastActivity = time.Now()
}

// IncrementRateLimited increments rate limited count
func (wm *WebhookMetrics) IncrementRateLimited(webhookID uuid.UUID) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	
	metric := wm.getOrCreateMetric(webhookID)
	metric.RateLimited++
	metric.LastActivity = time.Now()
}

// IncrementCircuitBreakerOpen increments circuit breaker open count
func (wm *WebhookMetrics) IncrementCircuitBreakerOpen(webhookID uuid.UUID) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	
	metric := wm.getOrCreateMetric(webhookID)
	metric.CircuitBreakerOpen++
	metric.LastActivity = time.Now()
}

// RecordLatency records delivery latency
func (wm *WebhookMetrics) RecordLatency(webhookID uuid.UUID, latency time.Duration) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	
	metric := wm.getOrCreateMetric(webhookID)
	metric.TotalLatency += latency
	
	if metric.MinLatency == 0 || latency < metric.MinLatency {
		metric.MinLatency = latency
	}
	
	if latency > metric.MaxLatency {
		metric.MaxLatency = latency
	}
}

// RecordStatusCode records HTTP status code
func (wm *WebhookMetrics) RecordStatusCode(webhookID uuid.UUID, statusCode int) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	
	metric := wm.getOrCreateMetric(webhookID)
	if metric.StatusCodeCounts == nil {
		metric.StatusCodeCounts = make(map[int]int64)
	}
	metric.StatusCodeCounts[statusCode]++
}

// ResetMetrics resets metrics for a webhook
func (wm *WebhookMetrics) ResetMetrics(webhookID uuid.UUID) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	
	delete(wm.metrics, webhookID)
}

// getOrCreateMetric gets or creates a metric for a webhook
func (wm *WebhookMetrics) getOrCreateMetric(webhookID uuid.UUID) *WebhookMetric {
	metric, exists := wm.metrics[webhookID]
	if !exists {
		metric = &WebhookMetric{
			WebhookID:        webhookID,
			StatusCodeCounts: make(map[int]int64),
			RecentDeliveries: make([]time.Time, 0, 100),
		}
		wm.metrics[webhookID] = metric
	}
	return metric
}

// updateCalculatedFields updates calculated fields in metrics
func (wm *WebhookMetrics) updateCalculatedFields(metric *WebhookMetric) {
	total := metric.Deliveries + metric.Errors
	if total > 0 {
		metric.SuccessRate = float64(metric.Deliveries) / float64(total) * 100
		metric.ErrorRate = float64(metric.Errors) / float64(total) * 100
	}
	
	totalEvents := metric.Deliveries + metric.Filtered
	if totalEvents > 0 {
		metric.DeliveryRate = float64(metric.Deliveries) / float64(totalEvents) * 100
	}
	
	if metric.Deliveries > 0 {
		metric.AverageLatency = metric.TotalLatency / time.Duration(metric.Deliveries)
	}
}

// GetDeliveryRate calculates recent delivery rate (deliveries per minute)
func (wm *WebhookMetrics) GetDeliveryRate(webhookID uuid.UUID, windowMinutes int) float64 {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	
	metric, exists := wm.metrics[webhookID]
	if !exists {
		return 0
	}
	
	cutoff := time.Now().Add(-time.Duration(windowMinutes) * time.Minute)
	count := 0
	
	for _, delivery := range metric.RecentDeliveries {
		if delivery.After(cutoff) {
			count++
		}
	}
	
	return float64(count) / float64(windowMinutes)
}

// GetMetricsSummary returns a summary of all webhook metrics
func (wm *WebhookMetrics) GetMetricsSummary() *MetricsSummary {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	
	summary := &MetricsSummary{
		TotalWebhooks: len(wm.metrics),
	}
	
	for _, metric := range wm.metrics {
		summary.TotalDeliveries += metric.Deliveries
		summary.TotalErrors += metric.Errors
		summary.TotalFiltered += metric.Filtered
		summary.TotalRateLimited += metric.RateLimited
		summary.TotalCircuitBreakerOpen += metric.CircuitBreakerOpen
		
		if metric.LastActivity.After(summary.LastActivity) {
			summary.LastActivity = metric.LastActivity
		}
	}
	
	total := summary.TotalDeliveries + summary.TotalErrors
	if total > 0 {
		summary.OverallSuccessRate = float64(summary.TotalDeliveries) / float64(total) * 100
	}
	
	return summary
}

// MetricsSummary contains summary metrics for all webhooks
type MetricsSummary struct {
	TotalWebhooks            int       `json:"total_webhooks"`
	TotalDeliveries          int64     `json:"total_deliveries"`
	TotalErrors              int64     `json:"total_errors"`
	TotalFiltered            int64     `json:"total_filtered"`
	TotalRateLimited         int64     `json:"total_rate_limited"`
	TotalCircuitBreakerOpen  int64     `json:"total_circuit_breaker_open"`
	OverallSuccessRate       float64   `json:"overall_success_rate"`
	LastActivity             time.Time `json:"last_activity"`
}

// ExportMetrics exports metrics in a format suitable for monitoring systems
func (wm *WebhookMetrics) ExportMetrics() map[string]interface{} {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	
	exported := make(map[string]interface{})
	
	for webhookID, metric := range wm.metrics {
		prefix := fmt.Sprintf("webhook_%s", webhookID.String())
		
		exported[prefix+"_deliveries"] = metric.Deliveries
		exported[prefix+"_errors"] = metric.Errors
		exported[prefix+"_filtered"] = metric.Filtered
		exported[prefix+"_rate_limited"] = metric.RateLimited
		exported[prefix+"_circuit_breaker_open"] = metric.CircuitBreakerOpen
		exported[prefix+"_success_rate"] = metric.SuccessRate
		exported[prefix+"_error_rate"] = metric.ErrorRate
		exported[prefix+"_delivery_rate"] = metric.DeliveryRate
		exported[prefix+"_avg_latency_ms"] = metric.AverageLatency.Milliseconds()
		exported[prefix+"_min_latency_ms"] = metric.MinLatency.Milliseconds()
		exported[prefix+"_max_latency_ms"] = metric.MaxLatency.Milliseconds()
		
		// Export status code counts
		for statusCode, count := range metric.StatusCodeCounts {
			exported[fmt.Sprintf("%s_status_%d", prefix, statusCode)] = count
		}
	}
	
	return exported
}