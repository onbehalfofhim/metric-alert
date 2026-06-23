package pool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMetrics - тестовая структура, аналогичная Metrics из domain
type TestMetrics struct {
	ID    string
	MType string
	Delta *int64
	Value *float64
	Hash  string
}

// Reset сбрасывает TestMetrics к начальным значениям
func (s *TestMetrics) Reset() {
	if s == nil {
		return
	}
	s.ID = ""
	s.MType = ""
	if s.Delta != nil {
		*s.Delta = 0
	}
	if s.Value != nil {
		*s.Value = 0
	}
	s.Hash = ""
}

// Пример использования с структурой, аналогичной Metrics из models
func TestPool_GetPut(t *testing.T) {
	// Создаем конструктор
	newFunc := func() *TestMetrics {
		delta := int64(0)
		value := float64(0.0)
		return &TestMetrics{
			Delta: &delta,
			Value: &value,
		}
	}

	// Создаем пул
	pool := New(newFunc)

	// Тестируем
	metrics := pool.Get()
	assert.NotNil(t, metrics)
	assert.Equal(t, "", metrics.ID)
	assert.Equal(t, "", metrics.MType)
	assert.Equal(t, int64(0), *metrics.Delta)
	assert.Equal(t, float64(0.0), *metrics.Value)

	// Изменяем значения
	metrics.ID = "test-metric"
	metrics.MType = "counter"
	*metrics.Delta = 100
	*metrics.Value = 3.14
	metrics.Hash = "test-hash"

	// Возвращаем в пул
	pool.Put(metrics)

	// Получаем снова
	newMetrics := pool.Get()
	assert.Equal(t, "", newMetrics.ID)
	assert.Equal(t, "", newMetrics.MType)
	assert.Equal(t, int64(0), *newMetrics.Delta)
	assert.Equal(t, float64(0.0), *newMetrics.Value)
	assert.Equal(t, "", newMetrics.Hash)
}
