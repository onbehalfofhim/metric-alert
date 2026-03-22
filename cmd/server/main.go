package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/onbehalfofhim/metric-alert/internal/compress"
	"github.com/onbehalfofhim/metric-alert/internal/config"
	"github.com/onbehalfofhim/metric-alert/internal/handler"
	"github.com/onbehalfofhim/metric-alert/internal/logger"
	"github.com/onbehalfofhim/metric-alert/internal/repository/memStorage"
	"github.com/onbehalfofhim/metric-alert/internal/service"
)

func main() {
	cfg := config.ParseServerFlags()

	if err := run(cfg); err != nil {
		log.Fatal(err)
	}
}

func run(cfg config.ServerConfig) error {
	storage := memStorage.NewMemStorage()
	service := service.NewMetricService(storage, cfg.FilePath)
	handler := handler.New(service)

	logger := logger.NewLogger()

	go func() {
		for {
			time.Sleep(time.Duration(cfg.StoreInterval) * time.Second)
			err := service.SaveToFile()
			if err != nil {
				print("cant' save to file")
			}
		}
	}()

	if cfg.Restore {
		err := service.LoadFromFile()
		if err != nil {
			return fmt.Errorf("cannot load from file: %w", err)
		}
	}

	return http.ListenAndServe(cfg.RunAddr, logger.RequestLogger(compress.GzipMiddleware(handler.Route())))
}
