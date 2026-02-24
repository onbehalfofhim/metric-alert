package models

// Интерфейс для взаимодействия с хранилищем метрик
type Storage interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
}

// Структура для хранения метрик
type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

// Конструктор для структуры хранилищая метрик
func (s *MemStorage) NewMemStorage() {
	s.gauges = make(map[string]float64, 0)
	s.counters = make(map[string]int64, 0)
}

// Метод обновления метрики с типом gauge
func (s *MemStorage) UpdateGauge(name string, value float64) {
	s.gauges[name] = value
}

// Метод обновления метрики с типом counter
func (s *MemStorage) UpdateCounter(name string, value int64) {
	s.counters[name] += value
}

// Метод получения метрики с типом gauge
func (s *MemStorage) GetGauge(name string) float64 {
	return s.gauges[name]
}

// Метод получения метрики с типом counter
func (s *MemStorage) GetCounter(name string) int64 {
	return s.counters[name]
}
