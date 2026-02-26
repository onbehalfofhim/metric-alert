package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
			metric := models.NewCounter(test.value.name, test.value.value)
			metricJSON, err := json.Marshal(metric)
			if assert.NoError(t, err) {
				assert.JSONEq(t, test.want, string(metricJSON))
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
	tests := []struct {
		name         string
		metric       models.Metric
		statusCode   int
		expectError  bool
		expectedPath string
	}{
		{
			name: "gauge success",
			metric: func() models.Metric {
				v := 123.45
				return models.Metric{
					ID:    "Metric1",
					MType: "gauge",
					Value: &v,
				}
			}(),
			statusCode:   http.StatusOK,
			expectError:  false,
			expectedPath: "/update/gauge/Metric1/123.45",
		},
		{
			name: "counter success",
			metric: func() models.Metric {
				d := int64(10)
				return models.Metric{
					ID:    "metric",
					MType: "counter",
					Delta: &d,
				}
			}(),
			statusCode:   http.StatusOK,
			expectError:  false,
			expectedPath: "/update/counter/metric/10",
		},
		{
			name: "404 error",
			metric: func() models.Metric {
				v := 1.0
				return models.Metric{
					ID:    "metric",
					MType: "gauge",
					Value: &v,
				}
			}(),
			statusCode:   http.StatusNotFound,
			expectError:  true,
			expectedPath: "/update/gauge/metric/1",
		},
		{
			name: "nil value gauge",
			metric: models.Metric{
				ID:    "metric",
				MType: "gauge",
				Value: nil,
			},
			statusCode:   http.StatusOK,
			expectError:  false, // запрос уйдёт на base URL
			expectedPath: "/",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			var receivedPath string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedPath = r.URL.Path
				w.WriteHeader(test.statusCode)
			}))
			defer server.Close()

			s := agent.NewSender(server.URL)

			err := s.Send([]models.Metric{test.metric})

			if test.expectError && err == nil {
				assert.Error(t, err)
			}

			if !test.expectError {
				assert.NoError(t, err)
			}

			assert.Equal(t, test.expectedPath, receivedPath)
		})
	}
}
