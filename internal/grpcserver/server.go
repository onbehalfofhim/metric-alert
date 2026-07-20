package grpcserver

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/onbehalfofhim/metric-alert/api/proto"
	"github.com/onbehalfofhim/metric-alert/internal/logger"
	"github.com/onbehalfofhim/metric-alert/internal/models"
	"github.com/onbehalfofhim/metric-alert/internal/service/metric"
)

type MetricsServer struct {
	pb.UnimplementedMetricsServer
	svc    *metric.MetricsService
	logger *logger.Logger
}

func NewMetricsServer(svc *metric.MetricsService, logger *logger.Logger) *MetricsServer {
	return &MetricsServer{
		svc:    svc,
		logger: logger,
	}
}

func (s *MetricsServer) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	reqMetrics := req.GetMetrics()

	if req == nil || len(reqMetrics) == 0 {
		return nil, status.Error(codes.InvalidArgument, "metrics list is empty")
	}

	s.logger.Info("received UpdateMetrics request",
		"count", len(reqMetrics),
	)

	metrics := make([]models.Metric, 0, len(reqMetrics))
	for _, m := range reqMetrics {
		var metric models.Metric

		switch m.GetType() {
		case pb.Metric_GAUGE:
			metric = models.NewGauge(m.GetId(), m.GetValue())
		case pb.Metric_COUNTER:
			metric = models.NewCounter(m.GetId(), m.GetDelta())
		default:
			return nil, status.Errorf(codes.InvalidArgument, "unknown metric type: %v", m.GetType())
		}

		metrics = append(metrics, metric)
	}

	if err := s.svc.UpdateBatch(ctx, metrics); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update metrics: %v", err)
	}

	s.logger.Info("metrics stored")

	return &pb.UpdateMetricsResponse{}, nil
}
