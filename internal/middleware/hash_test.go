package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/onbehalfofhim/metric-alert/internal/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashVerifier_NoHashHeader(t *testing.T) {
	called := false

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := HashVerifier("secret")(next)

	req := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader("test body"),
	)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHashVerifier_ValidHash(t *testing.T) {
	body := "test body"
	key := "secret"

	called := false

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true

		data, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		assert.Equal(t, body, string(data))

		w.WriteHeader(http.StatusOK)
	})

	handler := HashVerifier(key)(next)

	req := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader(body),
	)

	hash := crypto.HashSHA256([]byte(body), key)
	req.Header.Set("HashSHA256", hash)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHashVerifier_InvalidHash(t *testing.T) {
	called := false

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	handler := HashVerifier("secret")(next)

	req := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader("test body"),
	)

	req.Header.Set("HashSHA256", "invalid-hash")

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHashVerifier_BodyRestored(t *testing.T) {
	body := "important payload"
	key := "secret"

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		assert.Equal(t, body, string(data))
		w.WriteHeader(http.StatusOK)
	})

	handler := HashVerifier(key)(next)

	req := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader(body),
	)

	req.Header.Set(
		"HashSHA256",
		crypto.HashSHA256([]byte(body), key),
	)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHMACEqual(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{
			name: "equal",
			a:    "abc",
			b:    "abc",
			want: true,
		},
		{
			name: "different",
			a:    "abc",
			b:    "abd",
			want: false,
		},
		{
			name: "different length",
			a:    "abc",
			b:    "abcd",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, hmacEqual(tt.a, tt.b))
		})
	}
}
