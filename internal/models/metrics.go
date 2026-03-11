package models

import "fmt"

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

func NewGauge(id string, v float64) Metric {
	return Metric{
		ID:    id,
		MType: Gauge,
		Value: &v,
	}
}

func NewCounter(id string, d int64) Metric {
	return Metric{
		ID:    id,
		MType: Counter,
		Delta: &d,
	}
}

// переопределение метода для проверки и отладки метрик
func (m Metric) String() string {
	switch m.MType {
	case "gauge":
		if m.Value != nil {
			return fmt.Sprintf("%s = %f", m.ID, *m.Value)
		}
	case "counter":
		if m.Delta != nil {
			return fmt.Sprintf("%s = %d", m.ID, *m.Delta)
		}
	}
	return fmt.Sprintf("%s = <nil>", m.ID)
}
