package handler

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/onbehalfofhim/metric-alert/internal/logger"
	"github.com/onbehalfofhim/metric-alert/internal/repository/inmemory"
	"github.com/onbehalfofhim/metric-alert/internal/service"
)

func BenchmarkRootHandler(b *testing.B) {
	storage := inmemory.NewMemStorage()

	for i := 0; i < 10000; i++ {
		name := strconv.Itoa(i)

		_ = storage.UpdateGauge(name, float64(i))
		_ = storage.UpdateCounter(name, int64(i))
	}

	service := service.NewMetricService(storage)

	handler := &Handler{
		service: service,
		// logger можно замокать или взять slog.Default()
		logger: logger.NewLogger(),
	}

	h := handler.RootHandler()

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)
	}
}
