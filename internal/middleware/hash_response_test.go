package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onbehalfofhim/metric-alert/internal/crypto"
	"github.com/stretchr/testify/assert"
)

func TestHashResponseWriter_Write(t *testing.T) {
	rec := httptest.NewRecorder()

	rw := &hashResponseWriter{
		w:          rec,
		statusCode: http.StatusOK,
	}

	n, err := rw.Write([]byte("hello"))

	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, "hello", rw.buf.String())
}

func TestHashResponseWriter_WriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()

	rw := &hashResponseWriter{
		w: rec,
	}

	rw.WriteHeader(http.StatusAccepted)

	assert.Equal(t, http.StatusAccepted, rw.statusCode)
}

func TestHashResponseWriter_Header(t *testing.T) {
	rec := httptest.NewRecorder()

	rw := &hashResponseWriter{
		w: rec,
	}

	rw.Header().Set("Content-Type", "application/json")

	assert.Equal(
		t,
		"application/json",
		rec.Header().Get("Content-Type"),
	)
}

func TestHashSigner_SetsHashHeader(t *testing.T) {
	key := "secret"
	body := `{"status":"ok"}`

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	})

	handler := HashSigner(key)(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	expectedHash := crypto.HashSHA256([]byte(body), key)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, expectedHash, rec.Header().Get("HashSHA256"))
	assert.Equal(t, body, rec.Body.String())
}

func TestHashSigner_PreservesStatusCode(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	})

	handler := HashSigner("secret")(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "created", rec.Body.String())
}

func TestHashSigner_EmptyBody(t *testing.T) {
	key := "secret"

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	handler := HashSigner(key)(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	expectedHash := crypto.HashSHA256(nil, key)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, expectedHash, rec.Header().Get("HashSHA256"))
	assert.Empty(t, rec.Body.String())
}
