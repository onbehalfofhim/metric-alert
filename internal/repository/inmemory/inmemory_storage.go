package inmemory

import (
	"context"
	"errors"
	"sync"

	"github.com/onbehalfofhim/metric-alert/internal/models"
	"github.com/onbehalfofhim/metric-alert/internal/repository"
)

// Структура для хранения метрик
type MemStorage struct {
	mu       sync.RWMutex // для разделения потоков при работе с хранилищем метрик
	gauges   map[string]float64
	counters map[string]int64
}

// Конструктор для структуры хранилищая метрик
func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64, 0),
		counters: make(map[string]int64, 0),
	}
}

// Метод обновления метрики с типом gauge
func (s *MemStorage) UpdateGauge(name string, value float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.gauges[name] = value

	return nil
}

// Метод обновления метрики с типом counter
func (s *MemStorage) UpdateCounter(name string, value int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counters[name] += value

	return nil
}

// Метод получения метрики с типом gauge
func (s *MemStorage) GetGauge(name string) (float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.gauges[name]
	if !ok {
		return 0, repository.ErrMetricNotFound
	}

	return v, nil
}

// Метод получения метрики с типом counter
func (s *MemStorage) GetCounter(name string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.counters[name]
	if !ok {
		return 0, repository.ErrMetricNotFound
	}
	return v, nil
}

func (s *MemStorage) GetListGauges() map[string]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]float64, len(s.gauges))

	for k, v := range s.gauges {
		result[k] = v
	}

	return result
}

func (s *MemStorage) GetListCounters() map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]int64, len(s.counters))

	for k, v := range s.counters {
		result[k] = v
	}

	return result
}

func (s *MemStorage) Ping(ctx context.Context) error {
	return nil
}

func (s *MemStorage) UpdateBatch(ctx context.Context, metrics []models.Metric) error {
	for _, metric := range metrics {
		switch metric.MType {
		case "gauge":
			if metric.Value == nil {
				return errors.New("unknown metric value")
			}
			s.UpdateGauge(metric.ID, *metric.Value)

		case "counter":
			if metric.Delta == nil {
				return errors.New("unknown metric value")
			}
			s.UpdateCounter(metric.ID, *metric.Delta)

		default:
			return errors.New("unknown metric type")
		}
	}

	return nil
}
