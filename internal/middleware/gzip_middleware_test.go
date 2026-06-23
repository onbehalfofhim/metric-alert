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
		_, _ = w.Write([]byte(`{"ok":true}`)) //nolint:errcheck
	})

	ts := httptest.NewServer(GzipMiddleware(handler))
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
	req.Header.Set("Accept-Encoding", "gzip")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)

	err = resp.Body.Close()
	require.NoError(t, err)

	assert.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))

	gr, err := gzip.NewReader(resp.Body)
	require.NoError(t, err)

	err = gr.Close()
	require.NoError(t, err)

	body, _ := io.ReadAll(gr)
	assert.JSONEq(t, `{"ok":true}`, string(body))
}

func TestGzipMiddleware_SkipCompression(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("hello")) //nolint:errcheck
	})

	ts := httptest.NewServer(GzipMiddleware(handler))
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	err = resp.Body.Close() //nolint:errcheck
	require.NoError(t, err)

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
	_, err := gz.Write([]byte("test-data"))
	require.NoError(t, err)

	err = gz.Close()
	require.NoError(t, err)

	req, _ := http.NewRequest(http.MethodPost, ts.URL, &buf)
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	err = resp.Body.Close()
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
