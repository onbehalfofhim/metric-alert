package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/onbehalfofhim/metric-alert/internal/models"
)

// Обработчик корневого запроса
func RootHandler(storage models.MemStorage, res http.ResponseWriter, req *http.Request) {
	// Проверка префикса запроса и вызов обработчика
	if strings.HasPrefix(req.URL.Path, "/update/") {
		updateHandler(storage, res, req)
		return
	}

	http.NotFound(res, req)
}

// Обрабтчик запроса на обнолвение метрик
func updateHandler(storage models.MemStorage, res http.ResponseWriter, req *http.Request) {
	// Проверка типа запроса
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	// Проверка запроса на соответствие типу: /update/<type>/<name>/<value>
	path := strings.TrimPrefix(req.URL.Path, "/update/")
	queryParts := strings.Split(path, "/")

	if len(queryParts) != 3 {
		http.Error(res, "Invalid URL format", http.StatusNotFound)
		return
	}
	// Проверка на заполненность имени метрики
	metricType := queryParts[0]
	metricName := queryParts[1]
	metricValue := queryParts[2]
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
