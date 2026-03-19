package service

import (
	"strconv"
	"strings"

	"github.com/onbehalfofhim/metric-alert/internal/models"
	"github.com/onbehalfofhim/metric-alert/pkg/errors"
)

// Интерфейс для взаимодействия с хранилищем метрик
type Storage interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, value int64) error

	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)

	GetListGauges() map[string]float64
	GetListCounters() map[string]int64
}

type MetricsService struct {
	storage Storage
}

func NewMetricService(storage Storage) *MetricsService {
	return &MetricsService{storage: storage}
}

func (s *MetricsService) UpdateMetric(mType, name, value string) error {
	switch mType {
	case "gauge":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return errors.ErrInvalidValue
		}
		name = strings.ToLower(name)
		return s.storage.UpdateGauge(name, v)

	case "counter":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return errors.ErrInvalidValue
		}
		name = strings.ToLower(name)
		return s.storage.UpdateCounter(name, v)
	}

	return errors.ErrInvalidType
}

func (s *MetricsService) UpdateMetricJSON(metric models.Metric) error {
	switch metric.MType {
	case "gauge":
		if metric.Value == nil {
			return errors.ErrInvalidValue
		}
		name := strings.ToLower(metric.ID)
		return s.storage.UpdateGauge(name, *metric.Value)
	case "counter":
		if metric.Delta == nil {
			return errors.ErrInvalidValue
		}
		name := strings.ToLower(metric.ID)
		return s.storage.UpdateCounter(name, *metric.Delta)
	}

	return errors.ErrInvalidType
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

	return "", errors.ErrInvalidType
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

	return models.Metric{}, errors.ErrInvalidType
}

func (s *MetricsService) GetListGauges() map[string]float64 {
	return s.storage.GetListGauges()
}

func (s *MetricsService) GerListCounters() map[string]int64 {
	return s.storage.GetListCounters()
}
