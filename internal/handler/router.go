package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) Route() http.Handler {
	r := chi.NewRouter()

	r.Get("/", h.RootHandler())
	r.Route("/update", func(r chi.Router) {
		r.Post("/", h.UpdateHandlerJSON())
		r.Post("/{type}/{name}/{value}", h.UpdateHandler())
	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", h.GetMetricHandlerJSON())
		r.Get("/{type}/{name}", h.GetMetricHandler())
	})

	return r
}
