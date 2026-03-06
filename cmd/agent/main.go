package main

import (
	"fmt"
	"time"

	"github.com/onbehalfofhim/metric-alert/internal/agent"
	"github.com/onbehalfofhim/metric-alert/internal/config"
	"github.com/onbehalfofhim/metric-alert/internal/models"
)

func main() {
	cfg := config.ParseAgentFlags()

	// создание сборщика метрик
	collector := models.NewCollector()

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

		err := client.Send(metrics)
		if err != nil {
			fmt.Printf("Send falied: %s", err)
		}
		collector.CommitPollCount(delta)
	}
}
