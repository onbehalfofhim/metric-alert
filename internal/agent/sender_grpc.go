package agent

import (
	"context"
	"fmt"

	pb "github.com/onbehalfofhim/metric-alert/api/proto"
	"github.com/onbehalfofhim/metric-alert/internal/logger"
	"github.com/onbehalfofhim/metric-alert/internal/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type GRPCSender struct {
	client  pb.MetricsClient
	conn    *grpc.ClientConn
	localIP string
	logger  *logger.Logger
}

func NewGRPCMetricsSender(serverAddr string, logger *logger.Logger) (*GRPCSender, error) {
	conn, err := grpc.NewClient(
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	client := pb.NewMetricsClient(conn)

	localIP, err := getLocalIP(serverAddr)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &GRPCSender{
		client:  client,
		conn:    conn,
		localIP: localIP,
		logger:  logger,
	}, nil
}

func (s *GRPCSender) Close() error {
	return s.conn.Close()
}

func (s *GRPCSender) SendMetrics(ctx context.Context, metrics []models.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	pbMetrics := make([]*pb.Metric, 0, len(metrics))
	for _, m := range metrics {
		var pbMetric pb.Metric

		pbMetric.SetId(m.ID)

		switch m.MType {
		case models.Gauge:
			pbMetric.SetType(pb.Metric_GAUGE)
			if m.Value != nil {
				pbMetric.SetValue(*m.Value)
			}
		case models.Counter:
			pbMetric.SetType(pb.Metric_COUNTER)
			if m.Delta != nil {
				pbMetric.SetDelta(*m.Delta)
			}
		}

		pbMetrics = append(pbMetrics, &pbMetric)
	}

	req := &pb.UpdateMetricsRequest_builder{
		Metrics: pbMetrics,
	}

	reqCtx := ctx
	if s.localIP != "" {
		reqCtx = metadata.AppendToOutgoingContext(reqCtx, "x-real-ip", s.localIP)
	}

	s.logger.Info("sending metrics",
		"count", len(metrics),
		"ip", s.localIP,
	)
	_, err := s.client.UpdateMetrics(reqCtx, req.Build())
	if err != nil {
		s.logger.Error("grpc request failed", "error", err)
		return err
	}

	s.logger.Info("metrics sent successfully")

	return nil
}
