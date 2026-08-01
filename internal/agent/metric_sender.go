package agent

import (
	"context"

	"github.com/onbehalfofhim/metric-alert/internal/models"
)

type MetricsSender interface {
	SendMetrics(ctx context.Context, metrics []models.Metric) error
}
