package middleware

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"

	"github.com/onbehalfofhim/metric-alert/internal/crypto"
)

// DecryptRSA returns middleware to decrypt POST body
func DecryptRSA(privateKey *rsa.PrivateKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Пропускаем незашифрованные запросы
			if r.Header.Get("X-Encrypted") != "rsa" {
				next.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			defer func() {
				_ = r.Body.Close()
			}()

			decrypted, err := crypto.DecryptRSA(body, privateKey)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(decrypted))
			next.ServeHTTP(w, r)
		})
	}
}
