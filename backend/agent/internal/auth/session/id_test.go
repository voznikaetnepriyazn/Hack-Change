package session

import (
	"strings"
	"testing"
)

func TestNewIDReturnsUniqueURLSafeTokens(t *testing.T) {
	first, err := NewID()
	if err != nil {
		t.Fatalf("NewID() returned error: %v", err)
	}
	second, err := NewID()
	if err != nil {
		t.Fatalf("NewID() returned error: %v", err)
	}

	if first == second {
		t.Fatal("NewID() returned duplicate tokens")
	}
	if len(first) != 43 {
		t.Errorf("token length = %d, want 43", len(first))
	}
	if strings.ContainsAny(first, "+/=") {
		t.Errorf("token %q is not URL-safe unpadded base64", first)
	}
}
