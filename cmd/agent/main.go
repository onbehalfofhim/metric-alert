package main

import (
	"time"

	"github.com/onbehalfofhim/metric-alert/internal/agent"
	"github.com/onbehalfofhim/metric-alert/internal/config"
	"github.com/onbehalfofhim/metric-alert/internal/logger"
)

func main() {
	cfg, err := config.ParseAgentFlags()
	logger := logger.NewLogger()

	if err != nil {
		logger.Error("failed to set environment variables", "error", err)
	}

	// создание сборщика метрик
	collector := agent.NewCollector()

	// создание клиента для отправки метрик
	client := agent.NewSender(cfg.RunAddr)

	// горутина для сборка метрик
	go func() {
		for {
			collector.CollectMetrics()
			time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
		}
	}()

	// отправка метрик
	for {
		time.Sleep(time.Duration(cfg.ReportInterval) * time.Second)

		metrics, delta := collector.PrepareMetrics()

		err := client.SendBatch(metrics)
		if err != nil {
			logger.Error("failed to send metrics batch", "error", err)
		}
		collector.CommitPollCount(delta)
	}
}
