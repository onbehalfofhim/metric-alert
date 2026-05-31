package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/onbehalfofhim/metric-alert/internal/logger"
	"github.com/onbehalfofhim/metric-alert/internal/models"
	"github.com/onbehalfofhim/metric-alert/internal/repository"
	"github.com/onbehalfofhim/metric-alert/internal/service"
	"github.com/onbehalfofhim/metric-alert/internal/templates"
)

type Handler struct {
	service *service.MetricsService
	logger  *logger.Logger
	audit   service.AuditPublisher
}

func New(service *service.MetricsService, logger *logger.Logger, audit service.AuditPublisher) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
		audit:   audit,
	}
}

func mapToMetricView[T any](m map[string]T, format func(T) string) []templates.MetricView {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	result := make([]templates.MetricView, 0, len(keys))
	for _, name := range keys {
		result = append(result, templates.MetricView{
			Name:  name,
			Value: format(m[name]),
		})
	}

	return result
}

// Обработчик корневого запроса
func (h *Handler) RootHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		//форматируем метрики типа gauge
		gauges := mapToMetricView(
			h.service.GetListGauges(),
			func(v float64) string {
				return strconv.FormatFloat(v, 'f', -1, 64)
			},
		)

		//форматируем метрики типа counter
		counters := mapToMetricView(
			h.service.GetListCounters(),
			func(v int64) string {
				return strconv.FormatInt(v, 10)
			},
		)

		data := templates.MetricsPageData{
			Gauges:   gauges,
			Counters: counters,
		}

		res.Header().Set("Content-Type", "text/html; charset=utf-8")
		res.WriteHeader(http.StatusOK)

		err := templates.RenderMetricsPage(res, data)
		if err != nil {
			h.logger.Error("failed to render the metrics page", "error", err)

			http.Error(res,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
		}
	}
}

// Обрабтчик запроса на обнолвение метрик
func (h *Handler) UpdateHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// Проверка на заполненность имени метрики
		metricType := chi.URLParam(req, "type")
		metricName := chi.URLParam(req, "name")
		metricValue := chi.URLParam(req, "value")
		if metricName == "" {
			http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		err := h.service.UpdateMetric(metricType, metricName, metricValue)
		if err != nil {
			http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		res.WriteHeader(http.StatusOK)
		h.notifyAudit(req.RemoteAddr, metricName, nil)
	}
}

func (h *Handler) UpdateHandlerJSON() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Content-Type") != "application/json" {
			http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var m models.Metric
		if err := json.NewDecoder(req.Body).Decode(&m); err != nil {
			http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if m.ID == "" {
			http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		err := h.service.UpdateMetricJSON(m)
		if err != nil {
			http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		res.WriteHeader(http.StatusOK)
		h.notifyAudit(req.RemoteAddr, m.ID, nil)
	}
}

// Обработчик запроса на выдачу значения конкретной метрики
func (h *Handler) GetMetricHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		metricType := chi.URLParam(req, "type")
		metricName := chi.URLParam(req, "name")

		value, err := h.service.GetMetric(metricType, metricName)
		if err != nil {
			switch err {

			case repository.ErrMetricNotFound:
				http.Error(res,
					http.StatusText(http.StatusNotFound),
					http.StatusNotFound,
				)
			case service.ErrInvalidType:
				http.Error(res,
					http.StatusText(http.StatusBadRequest),
					http.StatusBadRequest,
				)
			default:
				h.logger.Error("failed to get metric from server", "error", err)

				http.Error(res,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)
			}

			return
		}
		res.WriteHeader(http.StatusOK)
		res.Write([]byte(value))

	}
}

func (h *Handler) GetMetricHandlerJSON() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")

		if req.Header.Get("Content-Type") != "application/json" {
			http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var m models.Metric
		if err := json.NewDecoder(req.Body).Decode(&m); err != nil {
			http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if m.ID == "" {
			http.Error(res, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		resp, err := h.service.GetMetricJSON(m.MType, m.ID)
		if err != nil {
			switch err {

			case repository.ErrMetricNotFound:
				http.Error(res,
					http.StatusText(http.StatusNotFound),
					http.StatusNotFound,
				)
			case service.ErrInvalidType:
				http.Error(res,
					http.StatusText(http.StatusBadRequest),
					http.StatusBadRequest,
				)
			default:
				h.logger.Error("failed to get metric from server", "error", err)

				http.Error(res,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)
			}

			return
		}

		enc := json.NewEncoder(res)
		if err := enc.Encode(resp); err != nil {
			http.Error(res, "cannot encode response body", http.StatusInternalServerError)
		}

		res.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) PingHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
		defer cancel()

		if err := h.service.Ping(ctx); err != nil {
			h.logger.Error("failed to connect to data base", "error", err)

			http.Error(res,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
			return
		}

		res.Header().Set("Content-Type", "text/html; charset=utf-8")
		res.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) UpdateBatchHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Content-Type") != "application/json" {
			http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var metrics []models.Metric
		if err := json.NewDecoder(req.Body).Decode(&metrics); err != nil {
			http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		err := h.service.UpdateBatch(req.Context(), metrics)
		if err != nil {
			http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		res.WriteHeader(http.StatusOK)

		h.notifyAudit(req.RemoteAddr, "", metrics)
	}
}

func (h *Handler) notifyAudit(ip string, name string, metrics []models.Metric) {
	if name == "" && metrics == nil {
		h.logger.Info("Cant't send notification: did't get name of metric(s)")
	}

	var auditMessage models.AuditMessage
	if name != "" {
		auditMessage = models.AuditMessage{
			TS:      time.Now().Unix(),
			Metrics: []string{name},
			IPAddr:  ip,
		}
	} else {
		auditMessage = models.AuditMessage{
			TS:      time.Now().Unix(),
			Metrics: make([]string, len(metrics)),
			IPAddr:  ip,
		}

		for i, metric := range metrics {
			auditMessage.Metrics[i] = metric.ID
		}
	}

	h.audit.Notify(auditMessage)
}
