package loop

import (
	"strings"
	"testing"
)

func TestNewLoopDetector(t *testing.T) {
	tests := []struct {
		name          string
		window        int
		threshold     int
		wantWindow    int
		wantThreshold int
	}{
		{
			name:          "defaults both non-positive values",
			window:        0,
			threshold:     -1,
			wantWindow:    12,
			wantThreshold: 4,
		},
		{
			name:          "defaults only non-positive window",
			window:        0,
			threshold:     2,
			wantWindow:    12,
			wantThreshold: 2,
		},
		{
			name:          "defaults only non-positive threshold",
			window:        3,
			threshold:     0,
			wantWindow:    3,
			wantThreshold: 4,
		},
		{
			name:          "preserves positive values",
			window:        5,
			threshold:     3,
			wantWindow:    5,
			wantThreshold: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewLoopDetector(tt.window, tt.threshold)
			if detector.window != tt.wantWindow {
				t.Errorf("window = %d, want %d", detector.window, tt.wantWindow)
			}
			if detector.threshold != tt.wantThreshold {
				t.Errorf("threshold = %d, want %d", detector.threshold, tt.wantThreshold)
			}
			if len(detector.history) != 0 {
				t.Errorf("initial history length = %d, want 0", len(detector.history))
			}
		})
	}
}

func TestRecordDetectsRepeatedCallAtThreshold(t *testing.T) {
	detector := NewLoopDetector(5, 3)

	for i := 1; i < 3; i++ {
		if err := detector.Record("search", "query", "result"); err != nil {
			t.Fatalf("Record() call %d returned unexpected error: %v", i, err)
		}
	}

	err := detector.Record("search", "query", "result")
	if err == nil {
		t.Fatal("Record() at the threshold returned nil, want loop detection error")
	}
	if !strings.Contains(err.Error(), "search called 3 times") {
		t.Errorf("Record() error = %q, want it to report 3 repeated calls", err)
	}
}

func TestRecordTreatsDifferentCallsAsDifferentSignatures(t *testing.T) {
	tests := []struct {
		name       string
		toolName   string
		input      string
		output     string
		variant    string
		changeTool bool
		changeIn   bool
		changeOut  bool
	}{
		{
			name:       "tool name",
			toolName:   "search",
			input:      "query",
			output:     "result",
			variant:    "fetch",
			changeTool: true,
		},
		{
			name:     "input",
			toolName: "search",
			input:    "query",
			output:   "result",
			variant:  "other query",
			changeIn: true,
		},
		{
			name:      "output",
			toolName:  "search",
			input:     "query",
			output:    "result",
			variant:   "other result",
			changeOut: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewLoopDetector(5, 2)
			if err := detector.Record(tt.toolName, tt.input, tt.output); err != nil {
				t.Fatalf("first Record() returned unexpected error: %v", err)
			}

			toolName, input, output := tt.toolName, tt.input, tt.output
			if tt.changeTool {
				toolName = tt.variant
			}
			if tt.changeIn {
				input = tt.variant
			}
			if tt.changeOut {
				output = tt.variant
			}

			if err := detector.Record(toolName, input, output); err != nil {
				t.Errorf("Record() for a different call returned unexpected error: %v", err)
			}
		})
	}
}

func TestRecordOnlyCountsCallsWithinWindow(t *testing.T) {
	detector := NewLoopDetector(3, 3)
	calls := []struct {
		toolName string
		input    string
		output   string
	}{
		{"search", "query", "result"},
		{"search", "query", "result"},
		{"fetch", "url", "page"},
		{"write", "file", "done"},
	}

	for i, call := range calls {
		if err := detector.Record(call.toolName, call.input, call.output); err != nil {
			t.Fatalf("Record() call %d returned unexpected error after old calls left the window: %v", i+1, err)
		}
	}

	if len(detector.history) != detector.window {
		t.Errorf("history length = %d, want window size %d", len(detector.history), detector.window)
	}
}

func TestResetClearsHistory(t *testing.T) {
	detector := NewLoopDetector(5, 2)
	if err := detector.Record("search", "query", "result"); err != nil {
		t.Fatalf("first Record() returned unexpected error: %v", err)
	}

	detector.Reset()
	if len(detector.history) != 0 {
		t.Fatalf("history length after Reset() = %d, want 0", len(detector.history))
	}

	if err := detector.Record("search", "query", "result"); err != nil {
		t.Fatalf("Record() after Reset() returned unexpected error: %v", err)
	}
}
