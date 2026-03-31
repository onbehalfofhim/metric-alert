package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"

	"github.com/onbehalfofhim/metric-alert/internal/logger"
	"github.com/onbehalfofhim/metric-alert/internal/repository/inmemory"
	"github.com/onbehalfofhim/metric-alert/internal/service"
)

func Test_RootHandler(t *testing.T) {
	storage := inmemory.NewMemStorage()
	service := service.NewMetricService(storage)
	logger := logger.NewLogger()
	h := New(service, logger)

	r := chi.NewRouter()
	r.Get("/", h.RootHandler())

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
		{name: "without metric name", request: "/update/gauge//3.6", expectedCode: http.StatusBadRequest},
		{name: "without metric name #2", request: "/update/counter//3", expectedCode: http.StatusBadRequest},
		{name: "wrong value", request: "/update/counter/metric1/3.6", expectedCode: http.StatusBadRequest},
		{name: "wrong value #2", request: "/update/gauge/metric1/abc", expectedCode: http.StatusBadRequest},
		{name: "wrong metric type", request: "/update/unknown/metric1/5", expectedCode: http.StatusBadRequest},
		{name: "correct query", request: "/update/gauge/metric1/6.7", expectedCode: http.StatusOK},
		{name: "correct query #2", request: "/update/counter/metric1/6", expectedCode: http.StatusOK},
	}

	storage := inmemory.NewMemStorage()
	service := service.NewMetricService(storage)
	logger := logger.NewLogger()
	h := New(service, logger)

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", h.UpdateHandler())

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
	storage := inmemory.NewMemStorage()
	service := service.NewMetricService(storage)
	logger := logger.NewLogger()
	h := New(service, logger)

	service.UpdateMetric("gauge", "metric1", "8.7")
	service.UpdateMetric("counter", "metric2", "-9")
	service.UpdateMetric("counter", "METric", "3")

	r := chi.NewRouter()
	r.Get("/value/{type}/{name}", h.GetMetricHandler())

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
			want:         http.StatusText(http.StatusNotFound),
		},
		{
			name:         "bad metric type",
			request:      "/value/unknown/metric1",
			expectedCode: http.StatusBadRequest,
			want:         http.StatusText(http.StatusBadRequest),
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
			assert.Equal(t, tt.want, strings.TrimSpace(string(resp.Body())), "Response value didn't match expected")
		})
	}

}
