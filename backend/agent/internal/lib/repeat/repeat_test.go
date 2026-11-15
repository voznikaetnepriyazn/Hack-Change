package repeat

import (
	"strings"
	"testing"
)

func TestNewRepetitionDetector(t *testing.T) {
	tests := []struct {
		name        string
		bufSize     int
		wantBufSize int
	}{
		{name: "default for zero", bufSize: 0, wantBufSize: 256},
		{name: "default for negative", bufSize: -1, wantBufSize: 256},
		{name: "custom size", bufSize: 64, wantBufSize: 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewRepetitionDetector(tt.bufSize)
			if detector.bufSize != tt.wantBufSize {
				t.Errorf("bufSize = %d, want %d", detector.bufSize, tt.wantBufSize)
			}
			if len(detector.buf) != 0 {
				t.Errorf("initial buffer length = %d, want 0", len(detector.buf))
			}
			if detector.triggered {
				t.Error("initial triggered = true, want false")
			}
		})
	}
}

func TestFeedDetectsRepeatedSingleCharacterAtWindowSize(t *testing.T) {
	const windowSize = 32
	detector := NewRepetitionDetector(windowSize)

	if got := detector.Feed(strings.Repeat("x", windowSize-1)); got {
		t.Fatal("Feed() detected repetition before the window was full")
	}
	if got := detector.Feed("x"); !got {
		t.Fatal("Feed() did not detect repetition when the window became full")
	}
}

func TestFeedDetectsRepeatedTextAcrossChunks(t *testing.T) {
	detector := NewRepetitionDetector(24)
	stream := strings.Repeat("abc", 20)
	chunkSizes := []int{1, 5, 2, 7, 3, 4}
	offset := 0

	for i := 0; offset < len(stream); i++ {
		size := chunkSizes[i%len(chunkSizes)]
		if offset+size > len(stream) {
			size = len(stream) - offset
		}
		if detector.Feed(stream[offset : offset+size]) {
			return
		}
		offset += size
	}

	t.Fatal("Feed() did not detect repeated text split across chunks")
}

func TestFeedDoesNotDetectNonRepeatingStream(t *testing.T) {
	detector := NewRepetitionDetector(64)
	stream := make([]byte, 256)
	for i := range stream {
		stream[i] = byte(i)
	}

	for offset := 0; offset < len(stream); {
		end := offset + 13
		if end > len(stream) {
			end = len(stream)
		}
		if detector.Feed(string(stream[offset:end])) {
			t.Fatalf("Feed() detected repetition in non-repeating stream ending at byte %d", end)
		}
		offset = end
	}
}

func TestFeedAccumulatesChunksUntilWindowIsFull(t *testing.T) {
	const windowSize = 24
	detector := NewRepetitionDetector(windowSize)

	for i := 0; i < windowSize-1; i++ {
		if got := detector.Feed("z"); got {
			t.Fatalf("Feed() detected repetition at byte %d, before the window was full", i+1)
		}
	}
	if got := detector.Feed("z"); !got {
		t.Fatal("Feed() did not detect repeated text after accumulating the full window")
	}
}

func TestFeedBoundsBufferForLongStreams(t *testing.T) {
	const windowSize = 32
	detector := NewRepetitionDetector(windowSize)

	for i := 0; i < 100; i++ {
		detector.Feed("abcdefgh")
		if got := len(detector.buf); got > windowSize {
			t.Fatalf("buffer length = %d after chunk %d, want at most %d", got, i, windowSize)
		}
	}
}

func TestFeedDoesNotTreatWholeSmallWindowAsRepeatedPattern(t *testing.T) {
	tests := []struct {
		window int
		text   string
	}{
		{window: 1, text: "x"},
		{window: 2, text: "ab"},
		{window: 3, text: "abc"},
		{window: 8, text: "abcdefgh"},
	}

	for _, tt := range tests {
		t.Run(strings.Repeat("x", tt.window), func(t *testing.T) {
			detector := NewRepetitionDetector(tt.window)
			if got := detector.Feed(tt.text); got {
				t.Fatalf("Feed(%q) detected a repetition without at least two matching pattern copies", tt.text)
			}
		})
	}
}

func TestFeedDetectionRemainsTriggered(t *testing.T) {
	detector := NewRepetitionDetector(8)
	if !detector.Feed(strings.Repeat("a", 8)) {
		t.Fatal("Feed() did not detect repeated pattern")
	}
	if !detector.Feed("different") {
		t.Fatal("Feed() returned false after detection was already triggered")
	}
}
