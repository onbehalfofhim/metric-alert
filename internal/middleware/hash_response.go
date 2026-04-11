package middleware

import (
	"bytes"
	"net/http"

	"github.com/onbehalfofhim/metric-alert/internal/crypto"
)

type hashResponseWriter struct {
	w          http.ResponseWriter
	statusCode int
	buf        bytes.Buffer
}

func (rw *hashResponseWriter) Header() http.Header {
	return rw.w.Header()
}

func (rw *hashResponseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
}

func (rw *hashResponseWriter) Write(b []byte) (int, error) {
	return rw.buf.Write(b)
}

func HashSigner(key string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				h.ServeHTTP(w, r)
				return
			}

			rw := &hashResponseWriter{
				w:          w,
				statusCode: http.StatusOK,
			}

			h.ServeHTTP(rw, r)

			hash := crypto.HashSHA256(rw.buf.Bytes(), key)
			w.Header().Set("HashSHA256", hash)
			w.WriteHeader(rw.statusCode)

			_, _ = w.Write(rw.buf.Bytes())
		})
	}
}
