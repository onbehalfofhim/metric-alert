package models

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
