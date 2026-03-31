package agent

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollector_CollectMetrics(t *testing.T) {
	c := NewCollector()

	c.CollectMetrics()
	c.CollectMetrics()

	metrics, _ := c.PrepareMetrics()

	if assert.NotEmpty(t, metrics) {
		for _, m := range metrics {
			if m.ID == "PollCount" {
				if m.Delta == nil || *m.Delta != 2 {
					assert.Equal(t, 2, *m.Delta)
				}
			}
		}
	}
}

func TestCollector_PrepareMetrics(t *testing.T) {
	c := NewCollector()

	c.CollectMetrics()

	metrics, _ := c.PrepareMetrics()
	checkValue := 9999

	if assert.NotEmpty(t, metrics) {
		for i := range metrics {
			if metrics[i].ID == "PollCount" && metrics[i].Delta != nil {
				*metrics[i].Delta = int64(checkValue)
			}
		}
	}

	newMetrics, _ := c.PrepareMetrics()
	for _, m := range newMetrics {
		if m.ID == "PollCount" && m.Delta != nil {
			assert.NotEqual(t, int64(checkValue), *m.Delta)
		}
	}

}

func TestCollector_CommitPollCount(t *testing.T) {
	tests := []struct {
		name     string
		initial  int64
		delta    int64
		expected int64
	}{
		{"simple case", 10, 3, 7},
		{"zero delta", 5, 0, 5},
		{"full reset", 8, 8, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Collector{
				pollCount: tt.initial,
			}

			c.CommitPollCount(tt.delta)

			result := atomic.LoadInt64(&c.pollCount)
			assert.Equal(t, tt.expected, result)
		})
	}
}
