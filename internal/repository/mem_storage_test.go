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

func TestMemStorage_GetGauge(t *testing.T) {
	tests := []struct {
		name       string
		metricName string
		want       float64
		wantErr    string
	}{
		{"positive", "test", 7.6, ""},
		{"negative", "test2", 0, "metric not found"},
	}

	s := NewMemStorage()
	s.UpdateGauge("test", 7.6)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := s.GetGauge(tt.metricName)
			if gotErr != nil {
				assert.EqualError(t, gotErr, tt.wantErr)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemStorage_GetCounter(t *testing.T) {
	tests := []struct {
		name       string
		metricName string
		want       int64
		wantErr    string
	}{
		{"positive", "test", 7, ""},
		{"negative", "test2", 0, "metric not found"},
	}

	s := NewMemStorage()
	s.UpdateCounter("test", 7)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := s.GetCounter(tt.metricName)
			if gotErr != nil {
				assert.EqualError(t, gotErr, tt.wantErr)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemStorage_GetListGauges(t *testing.T) {
	tests := []struct {
		name     string
		initial  map[string]float64
		expected map[string]float64
	}{
		{
			name:     "empty storage",
			initial:  map[string]float64{},
			expected: map[string]float64{},
		},
		{
			name: "single metric",
			initial: map[string]float64{
				"alloc": 10.5,
			},
			expected: map[string]float64{
				"alloc": 10.5,
			},
		},
		{
			name: "multiple metrics",
			initial: map[string]float64{
				"alloc": 10.5,
				"heap":  20.3,
			},
			expected: map[string]float64{
				"alloc": 10.5,
				"heap":  20.3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMemStorage()

			for k, v := range tt.initial {
				_ = s.UpdateGauge(k, v)
			}

			result := s.GetListGauges()

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMemStorage_GetListCounters(t *testing.T) {
	tests := []struct {
		name     string
		initial  map[string]int64
		expected map[string]int64
	}{
		{
			name:     "empty storage",
			initial:  map[string]int64{},
			expected: map[string]int64{},
		},
		{
			name: "single metric",
			initial: map[string]int64{
				"test": 10,
			},
			expected: map[string]int64{
				"test": 10,
			},
		},
		{
			name: "multiple metrics",
			initial: map[string]int64{
				"test":  10,
				"test2": 20,
			},
			expected: map[string]int64{
				"test":  10,
				"test2": 20,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMemStorage()

			for k, v := range tt.initial {
				_ = s.UpdateCounter(k, v)
			}

			result := s.GetListCounters()

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMemStorage_GetListGauges_ReturnsCopy(t *testing.T) {
	s := NewMemStorage()

	require.NoError(t, s.UpdateGauge("alloc", 10))

	result := s.GetListGauges()

	result["alloc"] = 999

	v, err := s.GetGauge("alloc")

	require.NoError(t, err)
	assert.Equal(t, float64(10), v)
}

func TestMemStorage_GetListCounters_ReturnsCopy(t *testing.T) {
	s := NewMemStorage()

	require.NoError(t, s.UpdateCounter("test", 10))

	result := s.GetListCounters()

	result["test"] = 999

	v, err := s.GetCounter("test")

	require.NoError(t, err)
	assert.Equal(t, int64(10), v)
}
