package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/onbehalfofhim/metric-alert/internal/logger"
	"github.com/onbehalfofhim/metric-alert/internal/middleware"
)

func (h *Handler) Route(log *logger.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestLogger(log))
	r.Use(middleware.GzipMiddleware)

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
