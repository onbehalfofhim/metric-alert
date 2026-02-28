package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"

	"github.com/onbehalfofhim/metric-alert/internal/handler"
	"github.com/onbehalfofhim/metric-alert/internal/models"
)

func TestMemStorage_UpdateGauge(t *testing.T) {
	type fields struct {
		metric string
		value  float64
	}
	tests := []struct {
		name  string
		value fields
		want  float64
	}{
		{
			name:  "first update metric",
			value: fields{metric: "metric1", value: 7.8},
			want:  7.8,
		},
		{
			name:  "second update metric",
			value: fields{metric: "metric1", value: 5.7},
			want:  5.7,
		},
		{
			name:  "zero value",
			value: fields{metric: "metric2", value: 0},
			want:  0,
		},
		{
			name:  "negative value",
			value: fields{metric: "metric3", value: -8.5},
			want:  -8.5,
		},
	}
	s := models.NewMemStorage()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s.UpdateGauge(test.value.metric, test.value.value)
			v, ok := s.GetGauge(test.value.metric)
			if ok {
				assert.Equal(t, test.want, v)
			}
		})
	}
}

func TestMemStorage_UpdateCounter(t *testing.T) {
	type fields struct {
		metric string
		value  int64
	}
	tests := []struct {
		name  string
		value fields
		want  int64
	}{
		{
			name:  "first update metric",
			value: fields{metric: "metric1", value: 543},
			want:  543,
		},
		{
			name:  "second update metric",
			value: fields{metric: "metric1", value: 7},
			want:  550,
		},
		{
			name:  "negative value",
			value: fields{metric: "metric1", value: -7},
			want:  543,
		},
	}

	s := models.NewMemStorage()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s.UpdateCounter(test.value.metric, test.value.value)

			v, ok := s.GetCounter(test.value.metric)
			if ok {
				assert.Equal(t, test.want, v)
			}
		})
	}
}

func Test_RootHandler(t *testing.T) {
	s := models.NewMemStorage()

	r := chi.NewRouter()
	r.Get("/", handler.RootHandler(s))

	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(r)
	// останавливаем сервер после завершения теста
	defer srv.Close()

	tests := []struct {
		name                string
		expectedContentType string
		expectedCode        int
	}{
		{name: "base test", expectedContentType: "text/html; charset=utf-8", expectedCode: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := resty.New().R()
			req.Method = http.MethodGet
			req.URL = srv.URL

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, tt.expectedCode, resp.StatusCode(), "Response code didn't match expected")
			assert.Equal(t, tt.expectedContentType, resp.Header().Get("Content-Type"))
		})
	}

}

func Test_UpdateHandler(t *testing.T) {
	tests := []struct {
		name         string
		request      string
		expectedCode int
	}{
		{name: "short update query", request: "/update/", expectedCode: http.StatusNotFound},
		{name: "without metric name", request: "/update/gauge//3.6", expectedCode: http.StatusNotFound},
		{name: "without metric name #2", request: "/update/counter//3", expectedCode: http.StatusNotFound},
		{name: "wrong value", request: "/update/counter/metric1/3.6", expectedCode: http.StatusBadRequest},
		{name: "wrong value #2", request: "/update/gauge/metric1/abc", expectedCode: http.StatusBadRequest},
		{name: "wrong metric type", request: "/update/unknown/metric1/5", expectedCode: http.StatusBadRequest},
		{name: "correct query", request: "/update/gauge/metric1/6.7", expectedCode: http.StatusOK},
		{name: "correct query #2", request: "/update/counter/metric1/6", expectedCode: http.StatusOK},
	}

	s := models.NewMemStorage()

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", handler.UpdateHandler(s))

	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(r)
	// останавливаем сервер после завершения теста
	defer srv.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// делаем запрос с помощью библиотеки resty к адресу запущенного сервера,
			// который хранится в поле URL соответствующей структуры
			req := resty.New().R()
			req.Method = http.MethodPost
			req.URL = srv.URL + tt.request

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, tt.expectedCode, resp.StatusCode(), "Response code didn't match expected")
		})
	}
}

func Test_GetMetricHandler(t *testing.T) {
	s := models.NewMemStorage()
	s.UpdateGauge("metric1", 8.7)
	s.UpdateCounter("metric2", -9)
	s.UpdateCounter("METric", 3)

	r := chi.NewRouter()
	r.Get("/value/{type}/{name}", handler.GetMetricHandler(s))

	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(r)
	// останавливаем сервер после завершения теста
	defer srv.Close()

	tests := []struct {
		name         string
		request      string
		expectedCode int
		want         string
	}{
		{
			name:         "request gauge",
			request:      "/value/gauge/metric1",
			expectedCode: http.StatusOK,
			want:         "8.7",
		},
		{
			name:         "request negative counter",
			request:      "/value/counter/metric2",
			expectedCode: http.StatusOK,
			want:         "-9",
		},
		{
			name:         "check metric name - toLowerCase",
			request:      "/value/counter/metric",
			expectedCode: http.StatusOK,
			want:         "3",
		},
		{
			name:         "metric not exists",
			request:      "/value/counter/metric3",
			expectedCode: http.StatusNotFound,
			want:         "Metric not found\n",
		},
		{
			name:         "bad metric type",
			request:      "/value/unknown/metric1",
			expectedCode: http.StatusBadRequest,
			want:         "Bad metric's type\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// делаем запрос с помощью библиотеки resty к адресу запущенного сервера,
			// который хранится в поле URL соответствующей структуры
			req := resty.New().R()
			req.Method = http.MethodGet
			req.URL = srv.URL + tt.request

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, tt.expectedCode, resp.StatusCode(), "Response code didn't match expected")
			assert.Equal(t, tt.want, string(resp.Body()), "Response value didn't match expected")
		})
	}

}
