package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubnetCheck(t *testing.T) {
	tests := []struct {
		name           string
		trustedSubnet  string
		clientIP       string
		expectedStatus int
		handlerCalled  bool
	}{
		{
			name:           "allowed ip",
			trustedSubnet:  "192.168.1.0/24",
			clientIP:       "192.168.1.15",
			expectedStatus: http.StatusOK,
			handlerCalled:  true,
		},
		{
			name:           "missing header",
			trustedSubnet:  "192.168.1.0/24",
			clientIP:       "",
			expectedStatus: http.StatusForbidden,
			handlerCalled:  false,
		},
		{
			name:           "invalid ip",
			trustedSubnet:  "192.168.1.0/24",
			clientIP:       "not-an-ip",
			expectedStatus: http.StatusForbidden,
			handlerCalled:  false,
		},
		{
			name:           "ip outside subnet",
			trustedSubnet:  "192.168.1.0/24",
			clientIP:       "10.0.0.1",
			expectedStatus: http.StatusForbidden,
			handlerCalled:  false,
		},
		{
			name:           "invalid subnet",
			trustedSubnet:  "invalid",
			clientIP:       "192.168.1.15",
			expectedStatus: http.StatusInternalServerError,
			handlerCalled:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			})

			handler := SubnetCheck(tt.trustedSubnet)(next)

			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if tt.clientIP != "" {
				req.Header.Set("X-Real-IP", tt.clientIP)
			}

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.handlerCalled, called)
		})
	}
}
