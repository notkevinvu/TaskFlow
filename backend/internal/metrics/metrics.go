// Package metrics provides Prometheus metrics for TaskFlow
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics

	// HTTPRequestsTotal counts total HTTP requests
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "taskflow_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestDuration measures HTTP request duration
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "taskflow_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// HTTPRequestsInFlight tracks current in-flight requests
	HTTPRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "taskflow_http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
	)

	// Business metrics

	// TasksCreatedTotal counts total tasks created
	TasksCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "taskflow_tasks_created_total",
			Help: "Total number of tasks created",
		},
		[]string{"category", "effort"},
	)

	// TasksCompletedTotal counts total tasks completed
	TasksCompletedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "taskflow_tasks_completed_total",
			Help: "Total number of tasks completed",
		},
		[]string{"category", "effort"},
	)

	// TasksBumpedTotal counts total task bumps
	TasksBumpedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "taskflow_tasks_bumped_total",
			Help: "Total number of task bumps",
		},
		[]string{"category"},
	)

	// TasksAtRisk tracks current at-risk task count
	TasksAtRisk = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "taskflow_tasks_at_risk",
			Help: "Current number of tasks at risk (bump_count >= 3)",
		},
	)

	// TasksDeletedTotal counts total tasks deleted
	TasksDeletedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "taskflow_tasks_deleted_total",
			Help: "Total number of tasks deleted",
		},
		[]string{"category"},
	)
)

// RecordTaskCreated increments the task created counter
func RecordTaskCreated(category, effort string) {
	if category == "" {
		category = "uncategorized"
	}
	if effort == "" {
		effort = "medium"
	}
	TasksCreatedTotal.WithLabelValues(category, effort).Inc()
}

// RecordTaskCompleted increments the task completed counter
func RecordTaskCompleted(category, effort string) {
	if category == "" {
		category = "uncategorized"
	}
	if effort == "" {
		effort = "medium"
	}
	TasksCompletedTotal.WithLabelValues(category, effort).Inc()
}

// RecordTaskBumped increments the task bumped counter
func RecordTaskBumped(category string) {
	if category == "" {
		category = "uncategorized"
	}
	TasksBumpedTotal.WithLabelValues(category).Inc()
}

// RecordTaskDeleted increments the task deleted counter
func RecordTaskDeleted(category string) {
	if category == "" {
		category = "uncategorized"
	}
	TasksDeletedTotal.WithLabelValues(category).Inc()
}

// SetTasksAtRisk sets the current at-risk task count
func SetTasksAtRisk(count float64) {
	TasksAtRisk.Set(count)
}
