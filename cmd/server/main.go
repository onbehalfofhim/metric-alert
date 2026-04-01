package main

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/onbehalfofhim/metric-alert/internal/config"
	"github.com/onbehalfofhim/metric-alert/internal/handler"
	"github.com/onbehalfofhim/metric-alert/internal/logger"
	"github.com/onbehalfofhim/metric-alert/internal/repository"
	"github.com/onbehalfofhim/metric-alert/internal/repository/file"
	"github.com/onbehalfofhim/metric-alert/internal/repository/inmemory"
	"github.com/onbehalfofhim/metric-alert/internal/repository/postgres"
	"github.com/onbehalfofhim/metric-alert/internal/service"
	"github.com/onbehalfofhim/metric-alert/migrations"
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
	var storage repository.Storage

	if cfg.DatabaseDSN != "" {
		db, err := sql.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			logger.Error("Error connect to data base", "error", err)
		}
		defer db.Close()

		if err := migrations.ApplyMigrations(db, "file://./../../migrations"); err != nil {
			logger.Error("Error apply migrations", "error", err)
		}

		storage = postgres.New(db)

	} else {
		storage = inmemory.NewMemStorage()

		fileStorage, err := file.NewFileStorage(storage, cfg.FilePath, logger)
		if err != nil {
			return fmt.Errorf("can't open file: %w", err)
		}

		defer fileStorage.Close()

		fileStorage.RunBackup(cfg.StoreInterval)

		if cfg.Restore {
			err := fileStorage.LoadFromFile()
			if err != nil {
				return fmt.Errorf("can't load from file: %w", err)
			}
		}
	}

	service := service.NewMetricService(storage)
	handler := handler.New(service, logger)

	return http.ListenAndServe(cfg.RunAddr, handler.Route(logger))
}
