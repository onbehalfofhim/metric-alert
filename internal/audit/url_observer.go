package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/onbehalfofhim/metric-alert/internal/logger"
	"github.com/onbehalfofhim/metric-alert/internal/models"
)

// URLObserver - регистация событий аудита через API.
type URLObserver struct {
	url    string
	logger *logger.Logger
	client *http.Client
}

// NewURLObserver - создает экземпляр наблюдателя для аудита.
func NewURLObserver(url string, logger *logger.Logger) *URLObserver {
	return &URLObserver{
		url:    url,
		logger: logger,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetID - получение идентификатора наблюдателя.
func (o *URLObserver) GetID() string {
	return fmt.Sprintf("audit-observer-%s", o.url)
}

// Notify - реагирование наблюдателя на событие аудита.
func (o *URLObserver) Notify(message models.AuditMessage) {
	if err := o.sendToURL(message); err != nil {
		o.logger.Info("failed to send audit message to URL", err)
		return
	}
	o.logger.Info("audit message sent to URL", "url", o.url)
}

func (o *URLObserver) sendToURL(message models.AuditMessage) error {
	buf := &bytes.Buffer{}

	if err := json.NewEncoder(buf).Encode(message); err != nil {
		return fmt.Errorf("can't encode request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, o.url, buf)
	if err != nil {
		return fmt.Errorf("can't create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close() //nolint:errcheck
	}()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
