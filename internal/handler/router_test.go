package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onbehalfofhim/metric-alert/internal/handler"
	"github.com/onbehalfofhim/metric-alert/internal/repository"
	"github.com/onbehalfofhim/metric-alert/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Route(t *testing.T) {
	storage := repository.NewMemStorage()
	service := service.NewMetricService(storage)
	h := handler.New(service)

	router := h.Route()

	tests := []struct {
		name   string
		method string
		path   string
		code   int
	}{
		{"root", "GET", "/", http.StatusOK},
		{"update", "POST", "/update/gauge/test/10", http.StatusOK},
		{"wrong metric name", "GET", "/value/gauge/test2", http.StatusNotFound},
		{"get metric", "GET", "/value/gauge/test", http.StatusOK},
		{"metric type", "GET", "/value/gau/test", http.StatusBadRequest},
		{"not found", "GET", "/unknown", http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.code, rr.Code)
		})
	}
}
