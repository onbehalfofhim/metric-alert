package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

			assert.Equal(t, test.want, s.GetGauge(test.value.metric))
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

			assert.Equal(t, test.want, s.GetCounter(test.value.metric))
		})
	}
}

func Test_RootHandler(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{name: "root query", request: "/", want: want{statusCode: 404}},
		{name: "short update query", request: "/update/", want: want{statusCode: 404}},
		{name: "without metric name", request: "/update/gauge//3.6", want: want{statusCode: 404}},
		{name: "without metric name #2", request: "/update/counter//3", want: want{statusCode: 404}},
		{name: "wrong value", request: "/update/counter/metric1/3.6", want: want{statusCode: 400}},
		{name: "wrong value #2", request: "/update/gauge/metric1/abc", want: want{statusCode: 400}},
		{name: "wrong metric type", request: "/update/unknown/metric1/5", want: want{statusCode: 400}},
		{name: "correct query", request: "/update/gauge/metric1/6.7", want: want{statusCode: 200}},
		{name: "correct query #2", request: "/update/counter/metric1/6", want: want{statusCode: 200}},
	}

	s := models.NewMemStorage()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			handler.RootHandler(*s, w, request)

			result := w.Result()
			err := result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
		})
	}
}
