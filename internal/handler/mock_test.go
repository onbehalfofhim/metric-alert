package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/onbehalfofhim/metric-alert/internal/audit"
	"github.com/onbehalfofhim/metric-alert/internal/models"
)

type mockMetricService struct {
	updateMetricFunc     func(string, string, string) error
	updateMetricJSONFunc func(models.Metric) error
	updateBatchFunc      func(context.Context, []models.Metric) error

	getMetricFunc     func(string, string) (string, error)
	getMetricJSONFunc func(string, string) (models.Metric, error)

	getListGaugesFunc   func() map[string]float64
	getListCountersFunc func() map[string]int64

	pingFunc func(context.Context) error
}

func (m *mockMetricService) UpdateMetric(t, n, v string) error {
	if m.updateMetricFunc != nil {
		return m.updateMetricFunc(t, n, v)
	}
	return nil
}

func (m *mockMetricService) UpdateMetricJSON(metric models.Metric) error {
	if m.updateMetricJSONFunc != nil {
		return m.updateMetricJSONFunc(metric)
	}
	return nil
}

func (m *mockMetricService) UpdateBatch(ctx context.Context, metrics []models.Metric) error {
	if m.updateBatchFunc != nil {
		return m.updateBatchFunc(ctx, metrics)
	}
	return nil
}

func (m *mockMetricService) GetMetric(t, n string) (string, error) {
	if m.getMetricFunc != nil {
		return m.getMetricFunc(t, n)
	}
	return "", nil
}

func (m *mockMetricService) GetMetricJSON(t, n string) (models.Metric, error) {
	if m.getMetricJSONFunc != nil {
		return m.getMetricJSONFunc(t, n)
	}
	return models.Metric{}, nil
}

func (m *mockMetricService) GetListGauges() map[string]float64 {
	if m.getListGaugesFunc != nil {
		return m.getListGaugesFunc()
	}
	return nil
}

func (m *mockMetricService) GetListCounters() map[string]int64 {
	if m.getListCountersFunc != nil {
		return m.getListCountersFunc()
	}
	return nil
}

func (m *mockMetricService) Ping(ctx context.Context) error {
	if m.pingFunc != nil {
		return m.pingFunc(ctx)
	}
	return nil
}

type mockAudit struct {
	called bool
	msg    models.AuditMessage
}

func (m *mockAudit) Notify(msg models.AuditMessage) {
	m.called = true
	m.msg = msg
}

func (m *mockAudit) Register(audit.AuditObserver)   {}
func (m *mockAudit) Deregister(audit.AuditObserver) {}
func (m *mockAudit) ObserversAmount() int { return 0}

func withURLParams(req *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()

	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}

	ctx := context.WithValue(
		req.Context(),
		chi.RouteCtxKey,
		rctx,
	)

	return req.WithContext(ctx)
}
