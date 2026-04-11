package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/onbehalfofhim/metric-alert/internal/crypto"
)

func HashVerifier(key string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				h.ServeHTTP(w, r)
				return
			}

			// проверяем, что клиент отправил hash
			recievedHash := r.Header.Get("HashSHA256")
			if recievedHash == "" {
				h.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)

			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			r.Body.Close()

			expectedHash := crypto.HashSHA256(body, key)

			if !hmacEqual(recievedHash, expectedHash) {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(body))
			h.ServeHTTP(w, r)
		})
	}
}

func hmacEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var res byte
	for i := 0; i < len(a); i++ {
		res |= a[i] ^ b[i]
	}
	return res == 0
}
