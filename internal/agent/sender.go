package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/onbehalfofhim/metric-alert/internal/models"
)

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

func (s *Sender) SendJSON(metrics []models.Metric) error {
	buf := &bytes.Buffer{}
	for _, v := range metrics {
		uri := fmt.Sprintf("%s/update", s.URL)

		buf.Reset()
		enc := json.NewEncoder(buf)
		if err := enc.Encode(v); err != nil {
			return fmt.Errorf("cannot encode request body: %w", err)
		}

		req, err := http.NewRequest(http.MethodPost, uri, buf)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.Client.Do(req)
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
