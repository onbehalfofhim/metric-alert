package main

import (
	"fmt"
	"net/http"

	"github.com/onbehalfofhim/metric-alert/internal/config"
	"github.com/onbehalfofhim/metric-alert/internal/handler"
	"github.com/onbehalfofhim/metric-alert/internal/logger"
	"github.com/onbehalfofhim/metric-alert/internal/repository/file"
	"github.com/onbehalfofhim/metric-alert/internal/repository/inmemory"
	"github.com/onbehalfofhim/metric-alert/internal/service"
)

func main() {
	cfg, error := config.ParseServerFlags()
	logger := logger.NewLogger()

	if error != nil {
		logger.Error("Error in parse flags and variables", "error", error)
	}

	if err := run(cfg, logger); err != nil {
		logger.Error("Error in run server", "error", err)
	}
}

func run(cfg config.ServerConfig, logger *logger.Logger) error {
	memStorage := inmemory.NewMemStorage()

	storage, err := file.NewFileStorage(memStorage, cfg.FilePath, logger)
	if err != nil {
		return fmt.Errorf("can't open file: %w", err)
	}
	defer storage.Close()

	service := service.NewMetricService(memStorage)
	handler := handler.New(service, logger)

	storage.RunBackup(cfg.StoreInterval)

	if cfg.Restore {
		err := storage.LoadFromFile()
		if err != nil {
			return fmt.Errorf("can't load from file: %w", err)
		}
	}

	return http.ListenAndServe(cfg.RunAddr, handler.Route(logger))
}
