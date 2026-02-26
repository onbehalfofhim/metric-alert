package agent_test

import (
	"github.com/onbehalfofhim/metric-alert/internal/agent"
	"github.com/onbehalfofhim/metric-alert/internal/models"
	"testing"
)

func TestSender_Send(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		url string
		// Named input parameters for target function.
		metrics []models.Metric
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := agent.NewSender(tt.url)
			gotErr := s.Send(tt.metrics)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Send() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Send() succeeded unexpectedly")
			}
		})
	}
}
