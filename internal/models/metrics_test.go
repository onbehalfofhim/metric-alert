package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
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
			metric := NewGauge(test.value.name, test.value.value)
			metricJSON, err := json.Marshal(metric)
			if assert.NoError(t, err) {
				assert.JSONEq(t, test.want, string(metricJSON))
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
			metric := NewCounter(test.value.name, test.value.value)
			metricJSON, err := json.Marshal(metric)
			if assert.NoError(t, err) {
				assert.JSONEq(t, test.want, string(metricJSON))
			}
		})
	}
}
