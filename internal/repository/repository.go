package repository

import (
	"context"
	"errors"

	"github.com/onbehalfofhim/metric-alert/internal/models"
)

// Интерфейс для взаимодействия с хранилищем метрик
type Storage interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, value int64) error

	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)

	GetListGauges() map[string]float64
	GetListCounters() map[string]int64

	Ping(ctx context.Context) error

	UpdateBatch(ctx context.Context, metrics []models.Metric) error
}

var (
	ErrMetricNotFound = errors.New("metric not found")
)
