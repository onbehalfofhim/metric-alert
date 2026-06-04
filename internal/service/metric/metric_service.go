package metric

import (
	"context"
	"strconv"

	"github.com/onbehalfofhim/metric-alert/internal/models"
	"github.com/onbehalfofhim/metric-alert/internal/repository"
	"github.com/onbehalfofhim/metric-alert/internal/service"
)

type MetricsService struct {
	storage repository.Storage
}

func NewMetricService(storage repository.Storage) *MetricsService {
	return &MetricsService{
		storage: storage,
	}
}

func (s *MetricsService) UpdateMetric(mType, name, value string) error {
	switch mType {
	case "gauge":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return service.ErrInvalidValue
		}
		return s.storage.UpdateGauge(name, v)
	case "counter":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return service.ErrInvalidValue
		}
		return s.storage.UpdateCounter(name, v)
	}

	return service.ErrInvalidType
}

func (s *MetricsService) UpdateMetricJSON(metric models.Metric) error {
	switch metric.MType {
	case "gauge":
		if metric.Value == nil {
			return service.ErrInvalidValue
		}
		return s.storage.UpdateGauge(metric.ID, *metric.Value)
	case "counter":
		if metric.Delta == nil {
			return service.ErrInvalidValue
		}
		return s.storage.UpdateCounter(metric.ID, *metric.Delta)
	}

	return service.ErrInvalidType
}

func (s *MetricsService) UpdateBatch(ctx context.Context, metrics []models.Metric) error {
	return s.storage.UpdateBatch(ctx, metrics)
}

func (s *MetricsService) GetMetric(mType, name string) (string, error) {
	switch mType {
	case "gauge":
		v, err := s.storage.GetGauge(name)
		if err != nil {
			return "", err
		}
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case "counter":
		v, err := s.storage.GetCounter(name)
		if err != nil {
			return "", err
		}
		return strconv.FormatInt(v, 10), nil
	}

	return "", service.ErrInvalidType
}

func (s *MetricsService) GetMetricJSON(mType, name string) (models.Metric, error) {
	value, err := s.GetMetric(mType, name)

	if err != nil {
		return models.Metric{}, err
	}

	switch mType {
	case "gauge":
		v, _ := strconv.ParseFloat(value, 64)
		return models.NewGauge(name, v), nil
	case "counter":
		v, _ := strconv.ParseInt(value, 10, 64)
		return models.NewCounter(name, v), nil
	}

	return models.Metric{}, service.ErrInvalidType
}

func (s *MetricsService) GetListGauges() map[string]float64 {
	return s.storage.GetListGauges()
}

func (s *MetricsService) GetListCounters() map[string]int64 {
	return s.storage.GetListCounters()
}

func (s *MetricsService) Ping(ctx context.Context) error {
	return s.storage.Ping(ctx)
}
