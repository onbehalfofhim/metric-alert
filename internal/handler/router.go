package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/onbehalfofhim/metric-alert/internal/models"
)

func Route(storage *models.MemStorage) http.Handler {
	r := chi.NewRouter()

	r.Get("/", RootHandler(storage))
	r.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", UpdateHandler(storage))
	})
	r.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", GetMetricHandler(storage))
	})

	return r
}
