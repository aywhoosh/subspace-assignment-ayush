package mocknet_test

import (
	"context"
	"testing"
	"time"

	"github.com/aywhoosh/subspace-assignment-ayush/internal/automation/mocknet"
)

func TestHumanBehavior_RandomDelay(t *testing.T) {
	hb := mocknet.HumanBehavior{
		RandomDelayMin: 100 * time.Millisecond,
		RandomDelayMax: 200 * time.Millisecond,
	}

	ctx := context.Background()
	start := time.Now()
	hb.RandomDelay(ctx)
	elapsed := time.Since(start)

	if elapsed < 100*time.Millisecond {
		t.Errorf("delay too short: %v", elapsed)
	}
	if elapsed > 300*time.Millisecond {
		t.Errorf("delay too long: %v", elapsed)
	}
}

func TestHumanBehavior_Presets(t *testing.T) {
	tests := []struct {
		name   string
		preset mocknet.HumanBehavior
	}{
		{"default", mocknet.DefaultHumanBehavior()},
		{"fast", mocknet.FastHumanBehavior()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preset.RandomDelayMax == 0 {
				t.Error("RandomDelayMax should not be zero")
			}
			if tt.preset.TypingSpeedMax == 0 {
				t.Error("TypingSpeedMax should not be zero")
			}
		})
	}
}

func TestSearchOptions(t *testing.T) {
	opts := mocknet.SearchOptions{
		Title:    "Engineer",
		Company:  "Tech Corp",
		Location: "San Francisco",
		PerPage:  10,
	}

	if opts.Title != "Engineer" {
		t.Errorf("expected title Engineer, got %s", opts.Title)
	}
}

func TestMessage(t *testing.T) {
	msg := mocknet.Message{
		From:      "Alice",
		Content:   "Hello",
		Timestamp: "2024-01-01 10:00:00",
	}

	if msg.From != "Alice" {
		t.Errorf("expected From=Alice, got %s", msg.From)
	}
	if msg.Content != "Hello" {
		t.Errorf("expected Content=Hello, got %s", msg.Content)
	}
}
