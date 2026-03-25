package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipMiddleware_CompressResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	})

	ts := httptest.NewServer(GzipMiddleware(handler))
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
	req.Header.Set("Accept-Encoding", "gzip")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))

	gr, err := gzip.NewReader(resp.Body)
	require.NoError(t, err)
	defer gr.Close()

	body, _ := io.ReadAll(gr)
	assert.JSONEq(t, `{"ok":true}`, string(body))
}

func TestGzipMiddleware_SkipCompression(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("hello"))
	})

	ts := httptest.NewServer(GzipMiddleware(handler))
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Empty(t, resp.Header.Get("Content-Encoding"))

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "hello", string(body))
}

func TestGzipMiddleware_DecompressRequest(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, "test-data", string(body))
		w.WriteHeader(http.StatusOK)
	})

	ts := httptest.NewServer(GzipMiddleware(handler))
	defer ts.Close()

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write([]byte("test-data"))
	gz.Close()

	req, _ := http.NewRequest(http.MethodPost, ts.URL, &buf)
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
