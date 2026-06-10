package inmemory

import (
	"context"
	"strconv"
	"testing"

	"github.com/onbehalfofhim/metric-alert/internal/models"
)

func prepareStorage(metricsCount int) *MemStorage {
	storage := NewMemStorage()

	for i := 0; i < metricsCount; i++ {
		storage.gauges[strconv.Itoa(i)] = float64(i)
		storage.counters[strconv.Itoa(i)] = int64(i)
	}

	return storage
}

func BenchmarkGetListGauges(b *testing.B) {
	storage := prepareStorage(10000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = storage.GetListGauges()
	}
}

func BenchmarkGetListCounters(b *testing.B) {
	storage := prepareStorage(10000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = storage.GetListCounters()
	}
}

func BenchmarkUpdateBatch(b *testing.B) {
	storage := NewMemStorage()

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
		_ = storage.UpdateBatch(
			context.Background(),
			metrics,
		)
	}
}
