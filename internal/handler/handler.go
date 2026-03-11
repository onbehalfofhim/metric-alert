package handler

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/onbehalfofhim/metric-alert/internal/models"
	"github.com/onbehalfofhim/metric-alert/internal/templates"
)

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
func RootHandler(storage *models.MemStorage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		//форматируем метрики типа gauge
		gauges := mapToMetricView(
			storage.GetListGauge(),
			func(v float64) string {
				return strconv.FormatFloat(v, 'f', -1, 64)
			},
		)

		//форматируем метрики типа counter
		counters := mapToMetricView(
			storage.GetListCounter(),
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
func UpdateHandler(storage *models.MemStorage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// Проверка на заполненность имени метрики
		metricType := chi.URLParam(req, "type")
		metricName := chi.URLParam(req, "name")
		metricValue := chi.URLParam(req, "value")
		if metricName == "" {
			http.Error(res, "Missing metric's name", http.StatusNotFound)
			return
		}

		// Проверка типа метрики
		switch metricType {
		case "gauge":
			// Проверка на корректность значения
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(res, "Invalid gauge value", http.StatusBadRequest)
				return
			}
			// Обновление метрики
			storage.UpdateGauge(metricName, value)
		case "counter":
			// Проверка на корректность значения
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(res, "Invalid counter value", http.StatusBadRequest)
				return
			}
			// Обновление метрики
			storage.UpdateCounter(metricName, value)
		default:
			http.Error(res, "Bad metric's type", http.StatusBadRequest)
			return
		}

		res.WriteHeader(http.StatusOK)
	}
}

// Обработчик запроса на выдачу значения конкретной метрики
func GetMetricHandler(storage *models.MemStorage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		metricType := chi.URLParam(req, "type")
		metricName := strings.ToLower(chi.URLParam(req, "name"))

		switch metricType {
		case "gauge":
			// Проверка на корректность значения
			value, ok := storage.GetGauge(metricName)
			if !ok {
				http.Error(res, "Metric not found", http.StatusNotFound)
				return
			}
			res.WriteHeader(http.StatusOK)
			res.Write([]byte(strconv.FormatFloat(value, 'f', -1, 64)))
		case "counter":
			// Проверка на корректность значения
			value, ok := storage.GetCounter(metricName)
			if !ok {
				http.Error(res, "Metric not found", http.StatusNotFound)
				return
			}
			res.WriteHeader(http.StatusOK)
			res.Write([]byte(strconv.FormatInt(value, 10)))
		default:
			http.Error(res, "Bad metric's type", http.StatusBadRequest)
			return
		}

	}
}
