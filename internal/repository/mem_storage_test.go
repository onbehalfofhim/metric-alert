package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemStorage_UpdateGauge(t *testing.T) {
	type fields struct {
		metric string
		value  float64
	}
	tests := []struct {
		name  string
		value fields
		want  float64
	}{
		{
			name:  "first update metric",
			value: fields{metric: "metric1", value: 7.8},
			want:  7.8,
		},
		{
			name:  "second update metric",
			value: fields{metric: "metric1", value: 5.7},
			want:  5.7,
		},
		{
			name:  "zero value",
			value: fields{metric: "metric2", value: 0},
			want:  0,
		},
		{
			name:  "negative value",
			value: fields{metric: "metric3", value: -8.5},
			want:  -8.5,
		},
	}
	s := NewMemStorage()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := s.UpdateGauge(test.value.metric, test.value.value)
			require.NoError(t, err)

			v, err := s.GetGauge(test.value.metric)
			require.NoError(t, err)

			assert.Equal(t, test.want, v)
		})
	}
}

func TestMemStorage_UpdateCounter(t *testing.T) {
	type fields struct {
		metric string
		value  int64
	}
	tests := []struct {
		name  string
		value fields
		want  int64
	}{
		{
			name:  "first update metric",
			value: fields{metric: "metric1", value: 543},
			want:  543,
		},
		{
			name:  "second update metric",
			value: fields{metric: "metric1", value: 7},
			want:  550,
		},
		{
			name:  "negative value",
			value: fields{metric: "metric1", value: -7},
			want:  543,
		},
	}

	s := NewMemStorage()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := s.UpdateCounter(test.value.metric, test.value.value)
			require.NoError(t, err)

			v, err := s.GetCounter(test.value.metric)
			require.NoError(t, err)

			assert.Equal(t, test.want, v)
		})
	}
}
