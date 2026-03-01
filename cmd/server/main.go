package main

import (
	// "fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/onbehalfofhim/metric-alert/internal/handler"
	"github.com/onbehalfofhim/metric-alert/internal/models"
)

func Route(storage *models.MemStorage) http.Handler {
	r := chi.NewRouter()

	r.Get("/", handler.RootHandler(storage))
	r.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", handler.UpdateHandler(storage))
	})
	r.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", handler.GetMetricHandler(storage))
	})

	return r
}

func main() {
	parseFlags()

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	// fmt.Printf("Running server on %s\n", flagRunAddr)
	storage := models.NewMemStorage()

	return http.ListenAndServe(flagRunAddr, Route(storage))
}
