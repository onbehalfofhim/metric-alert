package handler

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/onbehalfofhim/metric-alert/internal/models"
)

// Обработчик корневого запроса
func RootHandler(storage *models.MemStorage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/html; charset=utf-8")
		res.WriteHeader(http.StatusOK)

		res.Write([]byte("<html><head><title>Metrics</title></head><body>"))
		res.Write([]byte("<h1><span style='font-family:Comic Sans MS,cursive'>Current Metrics</span></h1>"))

		//создаем таблицу текущих метрик
		res.Write([]byte("<table align='center' border='1' cellpadding='1' cellspacing='1' style='width:500px'>"))
		res.Write([]byte("<thead><tr><th scope='col'>Name</th><th scope='col'>Value</th></tr></thread>"))
		res.Write([]byte("<tbody><tr><td colspan='2' style='text-align:center'>gauges</td></tr>"))

		//выводим метрики типа gauge
		gauges := storage.GetListGauge()
		gKeys := make([]string, 0, len(gauges))
		for k := range gauges {
			gKeys = append(gKeys, k)
		}
		sort.Strings(gKeys)
		for _, name := range gKeys {
			v := gauges[name]
			res.Write([]byte("<tr><td>" + name + "</td><td>" + strconv.FormatFloat(v, 'f', -1, 64) + "</td></tr>"))
		}

		res.Write([]byte("<tr><td colspan='2' style='text-align:center'>counters</td></tr>"))
		//выводим метрики типа gauge
		counters := storage.GetListCounter()
		cKeys := make([]string, 0, len(counters))
		for k := range counters {
			cKeys = append(cKeys, k)
		}
		sort.Strings(cKeys)
		for _, name := range cKeys {
			v := counters[name]
			res.Write([]byte("<tr><td>" + name + "</td><td>" + strconv.FormatInt(v, 10) + "</td></tr>"))
		}

		res.Write([]byte("</tbody>"))
		res.Write([]byte("</ul></body></html>"))
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
		// res.Write([]byte("OK"))
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
