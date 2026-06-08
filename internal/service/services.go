package service

import (
	"context"

	"github.com/onbehalfofhim/metric-alert/internal/audit"
	"github.com/onbehalfofhim/metric-alert/internal/models"
)

// MetricHandler описывает сервис взаимодействия с хранилищем.
type MetricHandler interface {
	UpdateMetric(mType, name, value string) error
	UpdateMetricJSON(metric models.Metric) error
	UpdateBatch(ctx context.Context, metrics []models.Metric) error

	GetMetric(mType, name string) (string, error)
	GetMetricJSON(mType, name string) (models.Metric, error)

	GetListGauges() map[string]float64
	GetListCounters() map[string]int64

	Ping(ctx context.Context) error
}

// Auditer описвает сервис аудита.
type Auditer interface {
	Notify(message models.AuditMessage)

	Register(o audit.AuditObserver)
	Deregister(o audit.AuditObserver)
}
