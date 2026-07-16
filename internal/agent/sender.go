package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/onbehalfofhim/metric-alert/internal/crypto"
	"github.com/onbehalfofhim/metric-alert/internal/models"
	"github.com/onbehalfofhim/metric-alert/internal/retry"
)

type Sender struct {
	Client    *http.Client
	URL       string
	key       string
	publicKey *rsa.PublicKey
}

func NewSender(url, key string, publicKey *rsa.PublicKey) *Sender {
	return &Sender{
		URL:       url,
		Client:    &http.Client{},
		key:       key,
		publicKey: publicKey,
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
		defer func() {
			_ = resp.Body.Close() //nolint:errcheck
		}()
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

	payload := buf.Bytes()

	// Шифрование
	if s.publicKey != nil {
		encrypted, err := crypto.EncryptRSA(payload, s.publicKey)
		if err != nil {
			return fmt.Errorf("can't encrypt request body: %w", err)
		}
		payload = encrypted
	}

	req, err := http.NewRequest(http.MethodPost, s.URL+endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("can't create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	u, err := url.Parse(s.URL)
	if err == nil {
		ip, err := getLocalIP(u.Host)
		if err == nil {
			req.Header.Set("X-Real-IP", ip)
		}
	}

	// чтобы сервер понял, что тело зашифровано
	if s.publicKey != nil {
		req.Header.Set("X-Encrypted", "rsa")
	}

	if s.key != "" {
		hash := crypto.HashSHA256(buf.Bytes(), s.key)
		req.Header.Set("HashSHA256", hash)
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("can't send a post-request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close() //nolint:errcheck
	}()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest {
		return fmt.Errorf("%w: %d", retry.ErrBadRequest, resp.StatusCode)
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

	return retry.Retry(func() error {
		return s.doRequest(metrics, "/updates/")
	})
}

func getLocalIP(serverAddr string) (string, error) {
	conn, err := net.Dial("udp", serverAddr)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}
