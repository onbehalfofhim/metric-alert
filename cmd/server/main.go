package main

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "net/http/pprof"

	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"

	pb "github.com/onbehalfofhim/metric-alert/api/proto"
	"github.com/onbehalfofhim/metric-alert/internal/audit"
	"github.com/onbehalfofhim/metric-alert/internal/buildinfo"
	"github.com/onbehalfofhim/metric-alert/internal/config"
	"github.com/onbehalfofhim/metric-alert/internal/crypto"
	"github.com/onbehalfofhim/metric-alert/internal/grpcserver"
	"github.com/onbehalfofhim/metric-alert/internal/handler"
	"github.com/onbehalfofhim/metric-alert/internal/logger"
	"github.com/onbehalfofhim/metric-alert/internal/repository"
	"github.com/onbehalfofhim/metric-alert/internal/repository/file"
	"github.com/onbehalfofhim/metric-alert/internal/repository/inmemory"
	"github.com/onbehalfofhim/metric-alert/internal/repository/postgres"
	"github.com/onbehalfofhim/metric-alert/internal/service"
	"github.com/onbehalfofhim/metric-alert/internal/service/auditservice"
	"github.com/onbehalfofhim/metric-alert/internal/service/metric"
	"github.com/onbehalfofhim/metric-alert/migrations"
)

func main() {
	cfg, error := config.ParseServerFlags()
	logger := logger.NewLogger()

	if error != nil {
		logger.Error("Error in parse flags and variables", "error", error)
	}

	buildinfo.Print()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	if err := run(cfg, logger); err != nil {
		logger.Error("Error in run server", "error", err)
	}
}

func run(cfg config.ServerConfig, logger *logger.Logger) error {
	var storage repository.Storage

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	var db *sql.DB
	var fileStorage *file.FileStorage
	var err error

	if cfg.DatabaseDSN != "" {
		db, err = sql.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			logger.Error("Error connect to data base", "error", err)
			return fmt.Errorf("can't connect to DB: %w", err)
		}

		if err = migrations.ApplyMigrations(db, "file://migrations"); err != nil {
			logger.Error("Error apply migrations", "error", err)
			return fmt.Errorf("can't apply migrations: %w", err)
		}

		storage = postgres.New(db)
		logger.Info("storage type: Postgres")
	} else {
		storage = inmemory.NewMemStorage()

		fileStorage, err = file.NewFileStorage(storage, cfg.FilePath, logger)
		if err != nil {
			return fmt.Errorf("can't open file: %w", err)
		}

		fileStorage.RunBackup(cfg.StoreInterval)

		if cfg.Restore {
			err = fileStorage.LoadFromFile()
			if err != nil {
				return fmt.Errorf("can't load from file: %w", err)
			}
		}
		logger.Info("storage type: Inmemory", "reestore from file", cfg.Restore, "file", cfg.FilePath)
	}

	var auditService service.Auditer

	if cfg.AuditFile != "" || cfg.AuditURL != "" {
		auditService = auditservice.NewAuditService(logger)
		if cfg.AuditFile != "" {
			auditService.Register(audit.NewFileObserver(cfg.AuditFile, logger))
		}
		if cfg.AuditURL != "" {
			auditService.Register(audit.NewURLObserver(cfg.AuditURL, logger))
		}
	} else {
		auditService = &auditservice.AuditService{}
		logger.Info("audit service is not enabled, skipping notification")
	}

	service := metric.NewMetricService(storage)
	handler := handler.New(service, logger, auditService)

	// получение приватного ключа для дешифровки входящих запросов
	var privateKey *rsa.PrivateKey
	if cfg.CryptoKey != "" {
		privateKey, err = crypto.LoadPrivateKey(cfg.CryptoKey)
		if err != nil {
			return fmt.Errorf("failed to load private key: %w", err)
		}
		logger.Info("private key loaded for decryption", "path", cfg.CryptoKey)
	}

	var grpcSrv *grpc.Server
	errChGRPC := make(chan error, 1)
	if cfg.GRPCAddr != "" {
		opts := []grpc.ServerOption{}

		if cfg.TrustedSubnet != "" {
			interceptor, err := grpcserver.SubnetCheckInterceptor(cfg.TrustedSubnet)
			if err != nil {
				return fmt.Errorf("create subnet interceptor: %w", err)
			}
			opts = append(opts, grpc.UnaryInterceptor(interceptor))
		}

		grpcSrv = grpc.NewServer(opts...)
		pb.RegisterMetricsServer(grpcSrv, grpcserver.NewMetricsServer(service, logger))

		go func() {
			listen, err := net.Listen("tcp", cfg.GRPCAddr)
			if err != nil {
				logger.Error("failed to listen gRPC", "error", err)
				errChGRPC <- err
				return
			}

			logger.Info("starting gRPC server", "Addr", cfg.GRPCAddr)

			if err := grpcSrv.Serve(listen); err != nil {
				logger.Error("failed to serve gRPC", "error", err)
				errChGRPC <- err
			}
		}()
	}

	srv := &http.Server{
		Addr:    cfg.RunAddr,
		Handler: handler.Route(logger, cfg.Key, privateKey, cfg.TrustedSubnet),
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting HTTP server", "addr", cfg.RunAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case err := <-errChGRPC:
		return err
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.Info("stopping servers ...")

	if grpcSrv != nil {
		grpcSrv.GracefulStop()
		logger.Info("gRPC server stopped")
	}

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	if fileStorage != nil {
		if err := fileStorage.Close(); err != nil {
			logger.Error("close file storage", "error", err)
		}
	}

	if db != nil {
		if err := db.Close(); err != nil {
			logger.Error("close db", "error", err)
		}
	}

	logger.Info("HTTP server stopped")

	return nil
}
