package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "net/http/pprof"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/onbehalfofhim/metric-alert/internal/audit"
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

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

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
			return fmt.Errorf("can't connect to DB: %w", err)
		}
		defer db.Close()

		if err := migrations.ApplyMigrations(db, "file://migrations"); err != nil {
			logger.Error("Error apply migrations", "error", err)
			return fmt.Errorf("can't apply migrations: %w", err)
		}

		storage = postgres.New(db)
		logger.Info("storage type: Postgres")

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
		logger.Info("storage type: Inmemory", "reestore from file", cfg.Restore, "file", cfg.FilePath)
	}

	auditService := service.NewAuditService(logger)

	if cfg.AuditFile != "" || cfg.AuditURL != "" {
		if cfg.AuditFile != "" {
			auditService.Register(audit.NewFileObserver(cfg.AuditFile, logger))
		}
		if cfg.AuditURL != "" {
			auditService.Register(audit.NewURLObserver(cfg.AuditURL, logger))
		}
	} else {
		logger.Info("audit service is not enabled, skipping notification")
	}

	service := service.NewMetricService(storage)
	handler := handler.New(service, logger, auditService)

	return http.ListenAndServe(cfg.RunAddr, handler.Route(logger, cfg.Key))
}
