package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onbehalfofhim/metric-alert/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestLoggingResponseWriter_Write(t *testing.T) {
	rec := httptest.NewRecorder()

	respData := &responseData{}

	lw := &loggingResponseWriter{
		ResponseWriter: rec,
		responseData:   respData,
	}

	n, err := lw.Write([]byte("hello"))

	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, 5, respData.size)
	assert.Equal(t, "hello", rec.Body.String())
}

func TestLoggingResponseWriter_WriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()

	respData := &responseData{}

	lw := &loggingResponseWriter{
		ResponseWriter: rec,
		responseData:   respData,
	}

	lw.WriteHeader(http.StatusCreated)

	assert.Equal(t, http.StatusCreated, respData.status)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestRequestLogger(t *testing.T) {

	log := logger.NewLogger()
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("response"))
	})
	handler := RequestLogger(log)(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.True(t, called)
	assert.Equal(t, http.StatusAccepted, rec.Code)
	assert.Equal(t, "response", rec.Body.String())

}
