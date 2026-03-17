package main

import (
	"log"
	"net/http"

	"github.com/onbehalfofhim/metric-alert/internal/config"
	"github.com/onbehalfofhim/metric-alert/internal/handler"
	"github.com/onbehalfofhim/metric-alert/internal/logger"
	"github.com/onbehalfofhim/metric-alert/internal/repository"
	"github.com/onbehalfofhim/metric-alert/internal/service"
)

func main() {
	cfg := config.ParseServerFlags()

	if err := run(cfg); err != nil {
		log.Fatal(err)
	}
}

func run(cfg config.ServerConfig) error {
	storage := repository.NewMemStorage()
	service := service.NewMetricService(storage)
	handler := handler.New(service)

	logger := logger.NewLogger()

	return http.ListenAndServe(cfg.RunAddr, logger.RequestLogger(handler.Route()))
}
