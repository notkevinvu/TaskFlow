package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecordTaskCreated_NormalizesEmptyValues(t *testing.T) {
	tests := []struct {
		name         string
		category     string
		effort       string
		wantCategory string
		wantEffort   string
	}{
		{
			name:         "empty category normalized to uncategorized",
			category:     "",
			effort:       "small",
			wantCategory: "uncategorized",
			wantEffort:   "small",
		},
		{
			name:         "empty effort normalized to medium",
			category:     "Work",
			effort:       "",
			wantCategory: "Work",
			wantEffort:   "medium",
		},
		{
			name:         "both empty normalized",
			category:     "",
			effort:       "",
			wantCategory: "uncategorized",
			wantEffort:   "medium",
		},
		{
			name:         "valid values unchanged",
			category:     "Personal",
			effort:       "large",
			wantCategory: "Personal",
			wantEffort:   "large",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get initial value
			initialValue := getCounterVecValue(t, TasksCreatedTotal, tt.wantCategory, tt.wantEffort)

			// Record metric
			RecordTaskCreated(tt.category, tt.effort)

			// Verify counter incremented with normalized values
			newValue := getCounterVecValue(t, TasksCreatedTotal, tt.wantCategory, tt.wantEffort)
			assert.Equal(t, initialValue+1, newValue, "counter should increment by 1")
		})
	}
}

func TestRecordTaskCompleted_NormalizesEmptyValues(t *testing.T) {
	// Get initial value for normalized labels
	initialValue := getCounterVecValue(t, TasksCompletedTotal, "uncategorized", "medium")

	// Record with empty values
	RecordTaskCompleted("", "")

	// Verify incremented
	newValue := getCounterVecValue(t, TasksCompletedTotal, "uncategorized", "medium")
	assert.Equal(t, initialValue+1, newValue)
}

func TestRecordTaskBumped_NormalizesEmptyCategory(t *testing.T) {
	initialValue := getCounterVecValue(t, TasksBumpedTotal, "uncategorized")

	RecordTaskBumped("")

	newValue := getCounterVecValue(t, TasksBumpedTotal, "uncategorized")
	assert.Equal(t, initialValue+1, newValue)
}

func TestRecordTaskDeleted_NormalizesEmptyCategory(t *testing.T) {
	initialValue := getCounterVecValue(t, TasksDeletedTotal, "uncategorized")

	RecordTaskDeleted("")

	newValue := getCounterVecValue(t, TasksDeletedTotal, "uncategorized")
	assert.Equal(t, initialValue+1, newValue)
}

func TestSetTasksAtRisk(t *testing.T) {
	SetTasksAtRisk(5.0)

	var m io_prometheus_client.Metric
	err := TasksAtRisk.Write(&m)
	require.NoError(t, err)
	assert.Equal(t, 5.0, m.GetGauge().GetValue())
}

// Helper to get counter value from CounterVec
func getCounterVecValue(t *testing.T, counter *prometheus.CounterVec, labels ...string) float64 {
	c, err := counter.GetMetricWithLabelValues(labels...)
	if err != nil {
		return 0
	}
	var m io_prometheus_client.Metric
	err = c.Write(&m)
	require.NoError(t, err)
	return m.GetCounter().GetValue()
}
