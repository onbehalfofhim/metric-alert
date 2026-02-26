package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/onbehalfofhim/metric-alert/internal/agent"
	"github.com/onbehalfofhim/metric-alert/internal/models"
)

func TestMetric_NewGauge(t *testing.T) {
	type addMetric struct {
		name  string
		value float64
	}
	tests := []struct {
		name  string
		value addMetric
		want  string
	}{
		{
			name:  "add metric",
			value: addMetric{name: "metric1", value: 0.564},
			want:  `{"id": "metric1", "type": "gauge", "value": 0.564}`,
		},
		{
			name:  "add negative metric",
			value: addMetric{name: "metric2", value: -0.98},
			want:  `{"id": "metric2", "type": "gauge", "value": -0.98}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metric := models.NewGauge(test.value.name, test.value.value)
			metricJson, err := json.Marshal(metric)
			if assert.NoError(t, err) {
				assert.JSONEq(t, test.want, string(metricJson))
			}
		})
	}
}

func TestMetric_NewCounter(t *testing.T) {
	type addMetric struct {
		name  string
		value int64
	}
	tests := []struct {
		name  string
		value addMetric
		want  string
	}{
		{
			name:  "add metric",
			value: addMetric{name: "metric1", value: 564},
			want:  `{"id": "metric1", "type": "counter", "delta": 564}`,
		},
		{
			name:  "add negative metric",
			value: addMetric{name: "metric2", value: 98},
			want:  `{"id": "metric2", "type": "counter", "delta": 98}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metric := models.NewCounter(test.value.name, test.value.value)
			metricJson, err := json.Marshal(metric)
			if assert.NoError(t, err) {
				assert.JSONEq(t, test.want, string(metricJson))
			}
		})
	}
}

func TestCollector_CollectMetrics(t *testing.T) {
	c := models.NewCollector()

	c.CollectMetrics()
	c.CollectMetrics()

	metrics := c.GetMetrics()

	if assert.NotEmpty(t, metrics) {
		for _, m := range metrics {
			if m.ID == "PollCount" {
				if m.Delta == nil || *m.Delta != 2 {
					assert.Equal(t, 2, *m.Delta)
				}
			}
		}
	}
}

func TestCollector_GetMetrics(t *testing.T) {
	c := models.NewCollector()

	c.CollectMetrics()

	metrics := c.GetMetrics()
	checkValue := 9999

	if assert.NotEmpty(t, metrics) {
		for i := range metrics {
			if metrics[i].ID == "PollCount" && metrics[i].Delta != nil {
				*metrics[i].Delta = int64(checkValue)
			}
		}
	}

	newMetrics := c.GetMetrics()
	for _, m := range newMetrics {
		if m.ID == "PollCount" && m.Delta != nil {
			assert.Equal(t, int64(checkValue), *m.Delta)
		}
	}

}

func TestSender_Send(t *testing.T) {
	value := float64(45)
	delta := int64(67)

	tests := []struct {
		name    string
		url     string
		metrics []models.Metric
		wantErr bool
	}{
		{
			name: "send gauge metric",
			url:  "http://localhost:8080",
			metrics: []models.Metric{
				{
					ID:    "Metric1",
					MType: "gauge",
					Value: &value,
				},
			},
			wantErr: false,
		},
		{
			name: "send counter metric",
			url:  "http://localhost:8080",
			metrics: []models.Metric{
				{
					ID:    "Metric1",
					MType: "counter",
					Delta: &delta,
				},
			},
			wantErr: false,
		},
		{
			name: "send empty value",
			url:  "http://localhost:8080",
			metrics: []models.Metric{
				{
					ID:    "Metric1",
					MType: "gauge",
					Value: nil,
				},
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := agent.NewSender(test.url)
			gotErr := s.Send(test.metrics)

			if gotErr != nil {
				if !test.wantErr {
					t.Errorf("Send() failed: %v", gotErr)
				}
				return
			}
			if test.wantErr {
				t.Fatal("Send() succeeded unexpectedly")
			}
		})
	}
}
