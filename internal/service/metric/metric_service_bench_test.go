package metric

import (
	"context"
	"strconv"
	"testing"

	"github.com/onbehalfofhim/metric-alert/internal/models"
	"github.com/onbehalfofhim/metric-alert/internal/repository/inmemory"
)

func BenchmarkMetricServiceUpdateMetric(b *testing.B) {
	repo := inmemory.NewMemStorage()
	service := NewMetricService(repo)

	b.ResetTimer()

	b.Run("gauge", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = service.UpdateMetric(
				"gauge",
				"Alloc",
				"123.45",
			)
		}
	})

	b.Run("counter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = service.UpdateMetric(
				"counter",
				"Alloc",
				"123",
			)
		}
	})
}

func BenchmarkUpdateMetricJSON(b *testing.B) {
	storage := inmemory.NewMemStorage()
	service := NewMetricService(storage)

	v := 123.45

	metric := models.Metric{
		ID:    "Alloc",
		MType: "gauge",
		Value: &v,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = service.UpdateMetricJSON(metric)
	}
}

func BenchmarkUpdateBatch(b *testing.B) {
	storage := inmemory.NewMemStorage()
	service := NewMetricService(storage)

	metrics := make([]models.Metric, 1000)

	for i := range metrics {
		v := float64(i)

		metrics[i] = models.Metric{
			ID:    strconv.Itoa(i),
			MType: "gauge",
			Value: &v,
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = service.UpdateBatch(
			context.Background(),
			metrics,
		)
	}
}
