package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/onbehalfofhim/metric-alert/internal/logger"
	"github.com/onbehalfofhim/metric-alert/internal/models"
	"github.com/onbehalfofhim/metric-alert/internal/repository/inmemory"
	"github.com/onbehalfofhim/metric-alert/internal/service/audit_service"
	"github.com/onbehalfofhim/metric-alert/internal/service/metric"
)

func ExampleHandler_UpdateHandlerJSON() {
	storage := inmemory.NewMemStorage()
	service := metric.NewMetricService(storage)

	logger := logger.NewLogger()
	auditer := audit_service.NewAuditService(logger)

	h := New(service, logger, auditer)

	val := 42.0
	body, _ := json.Marshal(models.Metric{
		ID:    "cpu",
		MType: models.Gauge,
		Value: &val,
	})

	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()
	h.UpdateHandlerJSON().ServeHTTP(res, req)

	v, _ := storage.GetGauge("cpu")
	fmt.Printf("%.0f\n", v)

	// Output:
	// 42
}
