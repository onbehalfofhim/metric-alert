package main

import (
	// "fmt"
	"log"
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
	storage := models.NewMemStorage()

	log.Fatal(http.ListenAndServe(":8080", Route(storage)))
}
