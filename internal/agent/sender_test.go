package agent

import (
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

			s := NewSender(server.URL, "", nil)

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

func TestDoRequest(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		key        string
		expectHash bool
		wantErr    bool
	}{
		{
			name:       "success without hash",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "success with hash",
			statusCode: http.StatusOK,
			key:        "secret",
			expectHash: true,
			wantErr:    false,
		},
		{
			name:       "bad request",
			statusCode: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			wantErr:    true,
		},
		{
			name:       "internal server error",
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))
				assert.Equal(t, "gzip", r.Header.Get("Accept-Encoding"))

				hash := r.Header.Get("HashSHA256")
				if tt.expectHash {
					assert.NotEmpty(t, hash)
				} else {
					assert.Empty(t, hash)
				}

				w.WriteHeader(tt.statusCode)
			}))
			defer ts.Close()

			s := &Sender{
				URL:    ts.URL,
				Client: ts.Client(),
				key:    tt.key,
			}

			err := s.doRequest(
				map[string]string{"test": "value"},
				"/update/",
			)

			if !tt.wantErr {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
		})
	}
}

func TestDoRequest_GzipBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gz, err := gzip.NewReader(r.Body)
		require.NoError(t, err)

		err = gz.Close()
		require.NoError(t, err)

		var body map[string]string
		err = json.NewDecoder(gz).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "value", body["test"])

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := &Sender{
		URL:    ts.URL,
		Client: ts.Client(),
	}
	err := s.doRequest(map[string]string{"test": "value"}, "/update/")
	require.NoError(t, err)
}

func TestSendJSON(t *testing.T) {
	var requests int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := &Sender{
		URL:    ts.URL,
		Client: ts.Client(),
	}

	metrics := []models.Metric{
		{ID: "m1"},
		{ID: "m2"},
		{ID: "m3"},
	}

	err := s.SendJSON(metrics)
	require.NoError(t, err)
	assert.Equal(t, int32(3), atomic.LoadInt32(&requests))
}

func TestSendBatch(t *testing.T) {
	var requests int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/updates/", r.URL.Path)
		atomic.AddInt32(&requests, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := &Sender{
		URL:    ts.URL,
		Client: ts.Client(),
	}

	metrics := []models.Metric{
		{ID: "m1"},
		{ID: "m2"},
		{ID: "m3"},
	}

	err := s.SendBatch(metrics)
	require.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&requests))
}
