package agent

import (
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
