package audit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/onbehalfofhim/metric-alert/internal/logger"
	"github.com/onbehalfofhim/metric-alert/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewURLObserver(t *testing.T) {
	logger := logger.NewLogger()

	o := NewURLObserver("http://localhost:8080", logger)

	assert.Equal(t, "http://localhost:8080", o.url)
	assert.Equal(t, logger, o.logger)
	assert.NotNil(t, o.client)
	assert.Equal(t, 5*time.Second, o.client.Timeout)
}

func TestURLObserver_GetID(t *testing.T) {

	o := &URLObserver{
		url: "http://localhost:8080",
	}
	assert.Equal(
		t,
		"audit-observer-http://localhost:8080",
		o.GetID(),
	)
}

func TestSendToURL(t *testing.T) {
	tests := []struct {
		name    string
		url     func(t *testing.T) string
		wantErr bool
	}{
		{
			name: "success",
			url: func(t *testing.T) string {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

					var msg models.AuditMessage
					err := json.NewDecoder(r.Body).Decode(&msg)
					require.NoError(t, err)

					assert.Equal(t, "127.0.0.0", msg.IPAddr)

					w.WriteHeader(http.StatusOK)
				}))

				t.Cleanup(ts.Close)

				return ts.URL
			},
		},
		{
			name: "invalid url",
			url: func(t *testing.T) string {
				return "://invalid-url"
			},
			wantErr: true,
		},
		{
			name: "connection refused",
			url: func(t *testing.T) string {
				return "http://127.0.0.1:1"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &URLObserver{
				url: tt.url(t),
				client: &http.Client{
					Timeout: time.Second,
				},
			}

			msg := models.AuditMessage{
				IPAddr: "127.0.0.0",
			}

			err := o.sendToURL(msg)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestNotify(t *testing.T) {
	var called atomic.Bool

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called.Store(true)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	o := &URLObserver{
		url:    ts.URL,
		logger: logger.NewLogger(),
		client: ts.Client(),
	}

	o.Notify(models.AuditMessage{
		IPAddr: "127.0.0.0",
	})

	assert.True(t, called.Load())
}
