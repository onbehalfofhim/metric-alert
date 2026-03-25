package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(cfg, logger, ctx); err != nil {
		logger.Error("Error in run server", "error", err)
	}
}

func run(cfg config.ServerConfig, logger *logger.Logger, ctx context.Context) error {
	memStorage := inmemory.NewMemStorage()

	storage, err := file.NewFileStorage(memStorage, cfg.FilePath, logger)
	if err != nil {
		return fmt.Errorf("can't open file: %w", err)
	}
	defer storage.Close()

	service := service.NewMetricService(memStorage)
	handler := handler.New(service, logger)

	go storage.RunBackup(ctx, cfg.StoreInterval)

	if cfg.Restore {
		err := storage.LoadFromFile()
		if err != nil {
			return fmt.Errorf("can't load from file: %w", err)
		}
	}

	return http.ListenAndServe(cfg.RunAddr, handler.Route(logger))
}
