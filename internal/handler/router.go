package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/onbehalfofhim/metric-alert/internal/logger"
	"github.com/onbehalfofhim/metric-alert/internal/middleware"
)

func (h *Handler) Route(log *logger.Logger, key string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestLogger(log))

	// проверка входящего тела
	r.Use(middleware.HashVerifier(key))

	// обработка ответа
	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.HashSigner(key))

	r.Get("/", h.RootHandler())
	r.Route("/update", func(r chi.Router) {
		r.Post("/", h.UpdateHandlerJSON())
		r.Post("/{type}/{name}/{value}", h.UpdateHandler())
	})
	r.Route("/updates", func(r chi.Router) {
		r.Post("/", h.UpdateBatchHandler())
	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", h.GetMetricHandlerJSON())
		r.Get("/{type}/{name}", h.GetMetricHandler())
	})
	r.Route("/ping", func(r chi.Router) {
		r.Get("/", h.PingHandler())
	})

	return r
}
