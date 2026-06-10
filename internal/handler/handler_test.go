package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/onbehalfofhim/metric-alert/internal/logger"
	"github.com/onbehalfofhim/metric-alert/internal/models"
	"github.com/onbehalfofhim/metric-alert/internal/repository"
	"github.com/onbehalfofhim/metric-alert/internal/service"
	"github.com/onbehalfofhim/metric-alert/internal/templates"
)

func TestUpdateHandler(t *testing.T) {
	tests := []struct {
		name         string
		params       map[string]string
		serviceErr   error
		expectedCode int
	}{
		{
			name: "success",
			params: map[string]string{
				"type":  "gauge",
				"name":  "Alloc",
				"value": "10.5",
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "empty metric name",
			params: map[string]string{
				"type":  "gauge",
				"name":  "",
				"value": "10.5",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "service error",
			params: map[string]string{
				"type":  "gauge",
				"name":  "Alloc",
				"value": "10.5",
			},
			serviceErr:   errors.New("boom"),
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockMetricService{
				updateMetricFunc: func(_, _, _ string) error {
					return tt.serviceErr
				},
			}

			audit := &mockAudit{}

			h := New(
				svc,
				logger.NewLogger(),
				audit,
			)

			req := httptest.NewRequest(
				http.MethodPost,
				"/update",
				nil,
			)

			req = withURLParams(req, tt.params)

			rr := httptest.NewRecorder()

			h.UpdateHandler().ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
		})
	}
}
func TestUpdateHandlerJSON(t *testing.T) {
	tests := []struct {
		name         string
		contentType  string
		body         string
		serviceErr   error
		expectedCode int
	}{
		{
			name:         "success",
			contentType:  "application/json",
			body:         `{"id":"Alloc","type":"gauge","value":10.5}`,
			expectedCode: http.StatusOK,
		},
		{
			name:         "wrong content type",
			contentType:  "text/plain",
			body:         `{}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid json",
			contentType:  "application/json",
			body:         `{`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "empty id",
			contentType:  "application/json",
			body:         `{"id":""}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "service error",
			contentType:  "application/json",
			body:         `{"id":"Alloc","type":"gauge","value":10.5}`,
			serviceErr:   errors.New("boom"),
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockMetricService{
				updateMetricJSONFunc: func(models.Metric) error {
					return tt.serviceErr
				},
			}

			audit := &mockAudit{}

			h := New(svc, logger.NewLogger(), audit)

			req := httptest.NewRequest(
				http.MethodPost,
				"/update",
				strings.NewReader(tt.body),
			)

			req.Header.Set("Content-Type", tt.contentType)

			rr := httptest.NewRecorder()

			h.UpdateHandlerJSON().ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
		})
	}
}

func TestGetMetricHandler(t *testing.T) {
	tests := []struct {
		name         string
		value        string
		serviceErr   error
		expectedCode int
	}{
		{
			name:         "success",
			value:        "10.5",
			expectedCode: http.StatusOK,
		},
		{
			name:         "not found",
			serviceErr:   repository.ErrMetricNotFound,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "invalid type",
			serviceErr:   service.ErrInvalidType,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "internal error",
			serviceErr:   errors.New("db"),
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockMetricService{
				getMetricFunc: func(_, _ string) (string, error) {
					return tt.value, tt.serviceErr
				},
			}

			h := New(svc, logger.NewLogger(), &mockAudit{})

			req := httptest.NewRequest(http.MethodGet, "/", nil)

			req = withURLParams(req, map[string]string{
				"type": "gauge",
				"name": "Alloc",
			})

			rr := httptest.NewRecorder()

			h.GetMetricHandler().ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			if tt.expectedCode == http.StatusOK {
				assert.Equal(t, "10.5", rr.Body.String())
			}
		})
	}
}

func TestPingHandler(t *testing.T) {
	tests := []struct {
		name         string
		pingErr      error
		expectedCode int
	}{
		{
			name:         "success",
			expectedCode: http.StatusOK,
		},
		{
			name:         "db error",
			pingErr:      errors.New("db"),
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockMetricService{
				pingFunc: func(context.Context) error {
					return tt.pingErr
				},
			}

			h := New(svc, logger.NewLogger(), &mockAudit{})

			req := httptest.NewRequest(http.MethodGet, "/ping", nil)

			rr := httptest.NewRecorder()

			h.PingHandler().ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
		})
	}
}

func TestMapToMetricView(t *testing.T) {
	input := map[string]float64{
		"b": 2.2,
		"a": 1.1,
		"c": 3.3,
	}

	got := mapToMetricView(
		input,
		func(v float64) string {
			return strconv.FormatFloat(v, 'f', -1, 64)
		},
	)

	expected := []templates.MetricView{
		{
			Name:  "a",
			Value: "1.1",
		},
		{
			Name:  "b",
			Value: "2.2",
		},
		{
			Name:  "c",
			Value: "3.3",
		},
	}

	assert.Equal(t, expected, got)
}

func TestRootHandler(t *testing.T) {
	svc := &mockMetricService{
		getListGaugesFunc: func() map[string]float64 {
			return map[string]float64{
				"Alloc": 10.5,
			}
		},
		getListCountersFunc: func() map[string]int64 {
			return map[string]int64{
				"PollCount": 42,
			}
		},
	}

	h := New(
		svc,
		logger.NewLogger(),
		&mockAudit{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	rr := httptest.NewRecorder()

	h.RootHandler().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(
		t,
		rr.Header().Get("Content-Type"),
		"text/html",
	)
}
