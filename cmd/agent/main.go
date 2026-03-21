package main

import (
	"log"
	"time"

	"github.com/onbehalfofhim/metric-alert/internal/agent"
	"github.com/onbehalfofhim/metric-alert/internal/config"
)

func main() {
	cfg := config.ParseAgentFlags()

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

		var err error
		if cfg.SendType == "simple" {
			err = client.Send(metrics)
		} else {
			err = client.SendJSON(metrics)
		}

		if err != nil {
			log.Printf("Send falied: %s", err)
		}
		collector.CommitPollCount(delta)
	}
}
