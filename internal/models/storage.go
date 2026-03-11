package models

import (
	"strings"
	"sync"
)

// Интерфейс для взаимодействия с хранилищем метрик
type Storage interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
}

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
func (s *MemStorage) UpdateGauge(name string, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.gauges[name] = value
}

// Метод обновления метрики с типом counter
func (s *MemStorage) UpdateCounter(name string, value int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counters[name] += value
}

// Метод получения метрики с типом gauge
func (s *MemStorage) GetGauge(name string) (float64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var value float64
	var ok bool

	for k := range s.gauges {
		gName := strings.ToLower(k)
		if gName == name {
			value, ok = s.gauges[k]
		}
	}
	return value, ok
}

// Метод получения метрики с типом counter
func (s *MemStorage) GetCounter(name string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var value int64
	var ok bool

	for k := range s.counters {
		cName := strings.ToLower(k)
		if cName == name {
			value, ok = s.counters[k]
		}
	}
	return value, ok
}

func (s *MemStorage) GetListGauge() map[string]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]float64, len(s.gauges))

	for k, v := range s.gauges {
		result[k] = v
	}

	return result
}

func (s *MemStorage) GetListCounter() map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]int64, len(s.counters))

	for k, v := range s.counters {
		result[k] = v
	}

	return result
}
