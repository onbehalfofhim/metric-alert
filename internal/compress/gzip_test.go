package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldCompress(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        bool
	}{
		{
			name:        "json",
			contentType: "application/json",
			want:        true,
		},
		{
			name:        "json with charset",
			contentType: "application/json; charset=utf-8",
			want:        true,
		},
		{
			name:        "html",
			contentType: "text/html",
			want:        true,
		},
		{
			name:        "html with charset",
			contentType: "text/html; charset=utf-8",
			want:        true,
		},
		{
			name:        "plain text",
			contentType: "text/plain",
			want:        false,
		},
		{
			name:        "empty",
			contentType: "",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, shouldCompress(tt.contentType))
		})
	}
}

func TestCompressWriter_WriteHeader(t *testing.T) {
	tests := []struct {
		name                    string
		contentType             string
		statusCode              int
		wantContentEncodingGzip bool
	}{
		{
			name:                    "json 200",
			contentType:             "application/json",
			statusCode:              http.StatusOK,
			wantContentEncodingGzip: true,
		},
		{
			name:                    "html 200",
			contentType:             "text/html",
			statusCode:              http.StatusOK,
			wantContentEncodingGzip: true,
		},
		{
			name:                    "plain text 200",
			contentType:             "text/plain",
			statusCode:              http.StatusOK,
			wantContentEncodingGzip: false,
		},
		{
			name:                    "json 500",
			contentType:             "application/json",
			statusCode:              http.StatusInternalServerError,
			wantContentEncodingGzip: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			cw := NewCompressWriter(rec)

			rec.Header().Set("Content-Type", tt.contentType)

			cw.WriteHeader(tt.statusCode)

			got := rec.Header().Get("Content-Encoding")

			if tt.wantContentEncodingGzip {
				assert.Equal(t, "gzip", got)
				assert.NotNil(t, cw.zw)
			} else {
				assert.Empty(t, got)
				assert.Nil(t, cw.zw)
			}
		})
	}
}

func TestCompressWriter_Write(t *testing.T) {
	rec := httptest.NewRecorder()

	cw := NewCompressWriter(rec)

	rec.Header().Set("Content-Type", "application/json")

	_, err := cw.Write([]byte(`{"test":"value"}`))
	require.NoError(t, err)

	require.NoError(t, cw.Close())

	assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))

	gz, err := gzip.NewReader(bytes.NewReader(rec.Body.Bytes()))
	require.NoError(t, err)
	defer func() {
		_ = gz.Close()
	}()

	body, err := io.ReadAll(gz)
	require.NoError(t, err)

	assert.JSONEq(t, `{"test":"value"}`, string(body))
}

func TestCompressReader_Read(t *testing.T) {
	var buf bytes.Buffer

	zw := gzip.NewWriter(&buf)

	_, err := zw.Write([]byte("hello world"))
	require.NoError(t, err)

	require.NoError(t, zw.Close())

	reader, err := NewCompressReader(io.NopCloser(&buf))
	require.NoError(t, err)

	data, err := io.ReadAll(reader)
	require.NoError(t, err)

	assert.Equal(t, "hello world", string(data))
}

func TestNewCompressReader_InvalidGzip(t *testing.T) {
	_, err := NewCompressReader(
		io.NopCloser(strings.NewReader("not gzip")),
	)

	require.Error(t, err)
}
