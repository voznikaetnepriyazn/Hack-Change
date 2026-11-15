package config

import (
	"testing"
	"time"
)

func TestDefaultLimits(t *testing.T) {
	got := DefaultLimits()
	if got.MaxIterations <= 0 || got.Deadline <= 0 || got.LLMTimeout <= 0 || got.ToolConcurrency <= 0 {
		t.Fatalf("DefaultLimits() returned non-positive limits: %+v", got)
	}
	if got.MaxIterations != 20 || got.Deadline != 5*time.Minute ||
		got.LLMTimeout != 4*time.Minute || got.ToolConcurrency != 3 {
		t.Errorf("DefaultLimits() = %+v, want 20 iterations, 5m deadline, 4m LLM timeout, concurrency 3", got)
	}
}

func TestLimitsWithDefaults(t *testing.T) {
	tests := []struct {
		name string
		in   Limits
		want Limits
	}{
		{
			name: "defaults unset fields",
			in:   Limits{MaxIterations: 2, LLMTimeout: time.Second},
			want: Limits{
				MaxIterations:   2,
				Deadline:        0,
				LLMTimeout:      time.Second,
				ToolConcurrency: 3,
			},
		},
		{
			name: "negative deadline uses default",
			in:   Limits{Deadline: -time.Second},
			want: Limits{
				MaxIterations:   20,
				Deadline:        5 * time.Minute,
				LLMTimeout:      4 * time.Minute,
				ToolConcurrency: 3,
			},
		},
		{
			name: "preserves fully configured values",
			in: Limits{
				MaxIterations:   7,
				Deadline:        2 * time.Minute,
				LLMTimeout:      30 * time.Second,
				ToolConcurrency: 5,
			},
			want: Limits{
				MaxIterations:   7,
				Deadline:        2 * time.Minute,
				LLMTimeout:      30 * time.Second,
				ToolConcurrency: 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.WithDefaults(); got != tt.want {
				t.Errorf("WithDefaults() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
