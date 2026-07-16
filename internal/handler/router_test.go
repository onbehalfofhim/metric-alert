package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/onbehalfofhim/metric-alert/internal/logger"
)

func TestRoute(t *testing.T) {
	svc := &mockMetricService{}
	audit := &mockAudit{}

	h := New(
		svc,
		logger.NewLogger(),
		audit,
	)

	router := h.Route(
		logger.NewLogger(),
		"",
		nil,
		"",
	)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "root",
			method: http.MethodGet,
			path:   "/",
		},
		{
			name:   "update json",
			method: http.MethodPost,
			path:   "/update/",
		},
		{
			name:   "update metric",
			method: http.MethodPost,
			path:   "/update/gauge/test/1",
		},
		{
			name:   "batch",
			method: http.MethodPost,
			path:   "/updates/",
		},
		{
			name:   "get metric",
			method: http.MethodGet,
			path:   "/value/gauge/test",
		},
		{
			name:   "get metric json",
			method: http.MethodPost,
			path:   "/value/",
		},
		{
			name:   "ping",
			method: http.MethodGet,
			path:   "/ping/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := httptest.NewRequest(
				tt.method,
				tt.path,
				nil,
			)

			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.NotEqual(
				t,
				http.StatusNotFound,
				rr.Code,
			)
		})
	}
}
