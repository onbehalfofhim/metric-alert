package main

import (
	"log"
	"net/http"

	"github.com/onbehalfofhim/metric-alert/internal/config"
	"github.com/onbehalfofhim/metric-alert/internal/handler"
	"github.com/onbehalfofhim/metric-alert/internal/models"
)

func main() {
	cfg := config.ParseServerFlags()

	if err := run(cfg); err != nil {
		log.Fatal(err)
	}
}

func run(cfg config.ServerConfig) error {
	storage := models.NewMemStorage()

	return http.ListenAndServe(cfg.RunAddr, handler.Route(storage))
}
