package agent

import (
	"fmt"
	"log"
	"net/http"

	"github.com/onbehalfofhim/metric-alert/internal/models"
)

type Sender struct {
	Client *http.Client
	URL    string
}

func (s *Sender) NewSender(url string) {
	s.URL = url
	s.Client = &http.Client{}
}

func (s *Sender) Send(metrics []models.Metric) {
	for _, v := range metrics {
		uri := fmt.Sprintf("%s/", s.URL)
		if v.MType == "gauge" && v.Value != nil {
			uri = fmt.Sprintf("%s/update/%s/%s/%v", s.URL, v.MType, v.ID, *v.Value)
		}
		if v.MType == "counter" && v.Delta != nil {
			uri = fmt.Sprintf("%s/update/%s/%s/%v", s.URL, v.MType, v.ID, *v.Delta)
		}
		fmt.Println(uri)

		resp, err := s.Client.Post(uri, "text/plain", nil)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(resp.Status)
	}
}
