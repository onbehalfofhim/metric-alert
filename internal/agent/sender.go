package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/onbehalfofhim/metric-alert/internal/models"
)

var retryDelays = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}

var ErrBadRequest = errors.New("bad request")

type Sender struct {
	Client *http.Client
	URL    string
}

func NewSender(url string) *Sender {
	return &Sender{
		URL:    url,
		Client: &http.Client{},
	}
}

func (s *Sender) Send(metrics []models.Metric) error {
	for _, v := range metrics {
		uri := fmt.Sprintf("%s/", s.URL)
		if v.MType == "gauge" && v.Value != nil {
			value := strconv.FormatFloat(*v.Value, 'f', -1, 64)
			uri = fmt.Sprintf("%s/update/%s/%s/%v", s.URL, v.MType, v.ID, value)
		}
		if v.MType == "counter" && v.Delta != nil {
			value := strconv.FormatInt(*v.Delta, 10)
			uri = fmt.Sprintf("%s/update/%s/%s/%v", s.URL, v.MType, v.ID, value)
		}

		resp, err := s.Client.Post(uri, "text/plain", nil)
		if err != nil {
			return fmt.Errorf("cannot send a post-request: %w", err)
		}
		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest {
			return fmt.Errorf("bad request: %d", resp.StatusCode)
		}
		resp.Body.Close()
	}

	return nil
}

func (s *Sender) doRequest(body any, endpoint string) error {
	buf := &bytes.Buffer{}
	gz := gzip.NewWriter(buf)

	if err := json.NewEncoder(gz).Encode(body); err != nil {
		return fmt.Errorf("can't encode request body: %w", err)
	}

	if err := gz.Close(); err != nil {
		return fmt.Errorf("can't close gzip: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.URL+endpoint, buf)
	if err != nil {
		return fmt.Errorf("can't create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("can't send a post-request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest {
		return fmt.Errorf("%w: %d", ErrBadRequest, resp.StatusCode)
	}

	if resp.StatusCode >= 500 {
		return fmt.Errorf("server error: %d", resp.StatusCode)
	}

	return nil
}

func (s *Sender) SendJSON(metrics []models.Metric) error {
	for _, m := range metrics {
		if err := s.doRequest(m, "/update/"); err != nil {
			return err
		}
	}

	return nil
}

func (s *Sender) SendBatch(metrics []models.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	return s.retry(func() error {
		return s.doRequest(metrics, "/updates/")
	})
}

func (s *Sender) retry(fn func() error) error {
	var err error

	for i := 0; i <= len(retryDelays); i++ {
		err = fn()
		if err == nil {
			return nil
		}

		if errors.Is(err, ErrBadRequest) {
			return err
		}

		if i < len(retryDelays) {
			time.Sleep(retryDelays[i])
		}
	}

	return fmt.Errorf("all retries failed: %w", err)
}
