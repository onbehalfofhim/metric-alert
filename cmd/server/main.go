package main

import (
	// "fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Интерфейс для взаимодействия с хранилищем метрик
type Storage interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
}

// Структура для хранения метрик
type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

// Конструктор для структуры хранилищая метрик
func (s *MemStorage) NewMemStorage() {
	s.gauges = make(map[string]float64, 0)
	s.counters = make(map[string]int64, 0)
}

// Метод обновления метрики с типом gauge
func (s *MemStorage) UpdateGauge(name string, value float64) {
	s.gauges[name] = value
}

// Метод обновления метрики с типом counter
func (s *MemStorage) UpdateCounter(name string, value int64) {
	s.counters[name] += value
}

// Обработчик корневого запроса
func rootHandler(storage MemStorage, res http.ResponseWriter, req *http.Request) {
	// Проверка префикса запроса и вызов обработчика
	if strings.HasPrefix(req.URL.Path, "/update/") {
		updateHandler(storage, res, req)
		return
	}

	http.NotFound(res, req)
}

// Обрабтчик запроса на обнолвение метрик
func updateHandler(storage MemStorage, res http.ResponseWriter, req *http.Request) {
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

func main() {
	storage := MemStorage{}
	storage.NewMemStorage()

	log.Fatal(http.ListenAndServe(":8080", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			rootHandler(storage, w, r)
		},
	)))
}
