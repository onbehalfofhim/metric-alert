package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/onbehalfofhim/metric-alert/internal/models"
)

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

			s := NewSender(server.URL, "")

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
