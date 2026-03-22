package service_test

import (
	"testing"

	"github.com/onbehalfofhim/metric-alert/internal/repository/memStorage"
	"github.com/onbehalfofhim/metric-alert/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestMetricsService_UpdateMetric(t *testing.T) {
	tests := []struct {
		name    string
		mType   string
		mName   string
		value   string
		wantErr bool
		textErr string
	}{
		{"update gauge", "gauge", "alloc", "7.6", false, ""},
		{"update counter", "counter", "test", "10", false, ""},
		{"wrong type", "type", "test", "0", true, "unknown metric type"},
		{"wrong gauge value", "gauge", "test", "bbbfbf", true, "unknown metric value"},
		{"wrong counter value", "counter", "test", "jfjfkf", true, "unknown metric value"},
	}

	storage := memStorage.NewMemStorage()
	s := service.NewMetricService(storage, "filename.txt")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := s.UpdateMetric(tt.mType, tt.name, tt.value)
			assert.Equal(t, tt.wantErr, gotErr != nil)
			if tt.wantErr {
				assert.EqualError(t, gotErr, tt.textErr)
			}
		})
	}
}

func TestMetricsService_GetMetric(t *testing.T) {
	tests := []struct {
		name    string
		mType   string
		mName   string
		want    string
		wantErr bool
		textErr string
	}{
		{"get gauge", "gauge", "alloc", "7.6", false, ""},
		{"get counter", "counter", "test", "10", false, ""},
		{"wrong type", "type", "test", "", true, "unknown metric type"},
		{"wrong counter metric name", "counter", "test2", "", true, "metric not found"},
		{"wrong gauge metric name", "gauge", "heap", "", true, "metric not found"},
	}

	storage := memStorage.NewMemStorage()
	s := service.NewMetricService(storage, "filename.txt")

	s.UpdateMetric("counter", "test", "10")
	s.UpdateMetric("gauge", "alloc", "7.6")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := s.GetMetric(tt.mType, tt.mName)
			if tt.wantErr {
				assert.EqualError(t, gotErr, tt.textErr)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
