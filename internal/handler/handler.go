package handler

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/onbehalfofhim/metric-alert/internal/models"
	"github.com/onbehalfofhim/metric-alert/internal/service"
	"github.com/onbehalfofhim/metric-alert/internal/templates"
	"github.com/onbehalfofhim/metric-alert/pkg/errors"
)

type Handler struct {
	service *service.MetricsService
}

func New(service *service.MetricsService) *Handler {
	return &Handler{
		service: service,
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
			h.service.GerListCounters(),
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
			http.Error(res, err.Error(), http.StatusInternalServerError)
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
			http.Error(res, "Missing metric's name", http.StatusNotFound)
			return
		}

		err := h.service.UpdateMetric(metricType, metricName, metricValue)
		if err != nil {
			http.Error(res, "Invalid type or metric name", http.StatusBadRequest)
			return
		}

		res.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) UpdateHandlerJSON() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Content-Type") != "application/json" {
			http.Error(res, "invalid content type", http.StatusBadRequest)
			return
		}

		var m models.Metric
		if err := json.NewDecoder(req.Body).Decode(&m); err != nil {
			http.Error(res, "bad json", http.StatusBadRequest)
			return
		}

		if m.ID == "" {
			http.Error(res, "missing metric's name", http.StatusNotFound)
			return
		}

		err := h.service.UpdateMetricJson(m)
		if err != nil {
			http.Error(res, "invalid type, metric value or name", http.StatusBadRequest)
			return
		}

		res.WriteHeader(http.StatusOK)
	}
}

// Обработчик запроса на выдачу значения конкретной метрики
func (h *Handler) GetMetricHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		metricType := chi.URLParam(req, "type")
		metricName := strings.ToLower(chi.URLParam(req, "name"))

		value, err := h.service.GetMetric(metricType, metricName)
		if err != nil {
			switch err {

			case errors.ErrMetricNotFound:
				http.Error(res, "Metric not found", http.StatusNotFound)

			case errors.ErrInvalidType:
				http.Error(res, "Bad metric's type", http.StatusBadRequest)

			default:
				http.Error(res, "Server error", http.StatusInternalServerError)
			}

			return
		}
		res.WriteHeader(http.StatusOK)
		res.Write([]byte(value))

	}
}

func (h *Handler) GetMetricHandlerJSON() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Content-Type") != "application/json" {
			http.Error(res, "invalid content type", http.StatusBadRequest)
			return
		}

		var m models.Metric
		if err := json.NewDecoder(req.Body).Decode(&m); err != nil {
			http.Error(res, "bad json", http.StatusBadRequest)
			return
		}

		if m.ID == "" {
			http.Error(res, "missing metric's name", http.StatusNotFound)
			return
		}

		resp, err := h.service.GetMetricJSON(m.MType, strings.ToLower(m.ID))
		if err != nil {
			switch err {
			case errors.ErrMetricNotFound:
				http.Error(res, "Metric not found", http.StatusNotFound)

			case errors.ErrInvalidType:
				http.Error(res, "Bad metric's type", http.StatusBadRequest)

			default:
				http.Error(res, "Server error", http.StatusInternalServerError)
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
