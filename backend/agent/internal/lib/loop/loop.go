package loop

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// LoopDetector - cycle detector
type LoopDetector struct {
	window    int
	threshold int
	history   []string
}

// Constructor
func NewLoopDetector(window, threshold int) *LoopDetector {
	if window <= 0 {
		window = 12
	}
	if threshold <= 0 {
		threshold = 4
	}
	return &LoopDetector{
		window:    window,
		threshold: threshold,
		history:   make([]string, 0, window),
	}
}

// Record loop detections
func (d *LoopDetector) Record(toolName string, input, output string) error {
	sig := sha256sum(toolName + "|" + input + "|" + output)
	d.history = append(d.history, sig)

	if len(d.history) > d.window {
		d.history = d.history[len(d.history)-d.window:]
	}

	counts := make(map[string]int)
	for _, s := range d.history {
		counts[s]++
		if counts[s] >= d.threshold {
			return fmt.Errorf("tool call loop detected: %s called %d times with same result",
				toolName, counts[s])
		}
	}
	return nil
}

// Reset
func (d *LoopDetector) Reset() {
	d.history = d.history[:0]
}

// sha256sum returns the SHA256 hash.
func sha256sum(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:8])
}
