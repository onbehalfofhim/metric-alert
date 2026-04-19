package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/onbehalfofhim/metric-alert/internal/agent"
	"github.com/onbehalfofhim/metric-alert/internal/config"
	"github.com/onbehalfofhim/metric-alert/internal/logger"
	"github.com/onbehalfofhim/metric-alert/internal/models"
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
	client := agent.NewSender(cfg.RunAddr, cfg.Key)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		logger.Info("shutting down")
		cancel()
	}()

	// горутина для сбора runtime-метрик
	go func() {
		ticker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				collector.CollectMetrics()

			case <-ctx.Done():
				logger.Info("collector stopped")
				return
			}
		}
	}()

	// горутина для сбора gopsutil-метрик
	go func() {
		ticker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				collector.CollectSystemMetrics()

			case <-ctx.Done():
				logger.Info("system collector stopped")
				return
			}
		}
	}()

	//создание буферизированного канала дляпринятия задач на отправку
	jobs := make(chan []models.Metric, cfg.RateLimit)

	var wg sync.WaitGroup

	// создание и запуск воркеров для отправки метрик на сервер
	for i := 0; i < cfg.RateLimit; i++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case metrics, ok := <-jobs:
					if !ok {
						logger.Info("worker stopped", "id", workerID)
						return
					}

					err := client.SendBatch(metrics)
					if err != nil {
						logger.Error("failed to send metrics batch",
							"worker", workerID,
							"error", err,
						)
					}

				case <-ctx.Done():
					logger.Info("worker received stop signal", "id", workerID)
					return
				}
			}
		}(i + 1)
	}

	// горутина для формирования задачи на отпраку меткри на сервер
	go func() {
		ticker := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				metrics, delta := collector.PrepareMetrics()
				if len(metrics) == 0 {
					continue
				}
				select {
				case jobs <- metrics:
					collector.CommitPollCount(delta)

				case <-ctx.Done():
					return
				}

			case <-ctx.Done():
				logger.Info("sender stopped")
				return
			}
		}
	}()

	<-ctx.Done()
	close(jobs)
	wg.Wait()

	logger.Info("agent stopped gracefully")
}
