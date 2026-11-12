package chunker

import (
	"context"
	"errors"
	"strings"
	"testing"
	"unicode/utf8"
)

var _ Chunker = (*RecursiveChunker)(nil)

func TestNewRecursiveChunkerRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		chunkSize int
		overlap   int
	}{
		{name: "zero chunk size", chunkSize: 0},
		{name: "negative chunk size", chunkSize: -1},
		{name: "negative overlap", chunkSize: 10, overlap: -1},
		{name: "overlap equals chunk size", chunkSize: 10, overlap: 10},
		{name: "overlap exceeds chunk size", chunkSize: 10, overlap: 11},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRecursiveChunker(tt.chunkSize, tt.overlap)
			if !errors.Is(err, ErrInvalidConfig) {
				t.Errorf("NewRecursiveChunker() error = %v, want %v", err, ErrInvalidConfig)
			}
			if got != nil {
				t.Errorf("NewRecursiveChunker() returned a chunker for invalid configuration")
			}
		})
	}
}

func TestRecursiveChunkerSplitsWithOverlapAndPreservesContent(t *testing.T) {
	const (
		chunkSize = 10
		overlap   = 3
		content   = "abcdefghijklmnopqrstuvwxyz"
	)
	chunker, err := NewRecursiveChunker(chunkSize, overlap)
	if err != nil {
		t.Fatalf("NewRecursiveChunker() returned error: %v", err)
	}

	chunks, err := chunker.Chunk(context.Background(), Document{Content: content})
	if err != nil {
		t.Fatalf("Chunk() returned error: %v", err)
	}
	if len(chunks) < 2 {
		t.Fatalf("Chunk() returned %d chunks, want multiple", len(chunks))
	}

	var rebuilt strings.Builder
	for i, part := range chunks {
		runeLen := utf8.RuneCountInString(part.Content)
		if runeLen > chunkSize {
			t.Errorf("chunk %d has %d runes, exceeds configured size %d: %q", i, runeLen, chunkSize, part.Content)
		}
		if strings.TrimSpace(part.Content) != part.Content || part.Content == "" {
			t.Errorf("chunk %d is empty or has surrounding whitespace: %q", i, part.Content)
		}
		if part.ID == ([16]byte{}) {
			t.Errorf("chunk %d has a zero ID", i)
		}
		if got := part.Metadata["chunk_index"]; got != i {
			t.Errorf("chunk %d metadata chunk_index = %v, want %d", i, got, i)
		}
		if got := part.Metadata["chunk_count"]; got != len(chunks) {
			t.Errorf("chunk %d metadata chunk_count = %v, want %d", i, got, len(chunks))
		}

		if i == 0 {
			rebuilt.WriteString(part.Content)
			continue
		}
		previous := []rune(chunks[i-1].Content)
		current := []rune(part.Content)
		if len(previous) < overlap || len(current) < overlap {
			t.Fatalf("chunks %d and %d are too short to contain the configured overlap", i-1, i)
		}
		for j := 0; j < overlap; j++ {
			if previous[len(previous)-overlap+j] != current[j] {
				t.Fatalf("chunks %d and %d do not share the expected %d-rune overlap: %q, %q",
					i-1, i, overlap, chunks[i-1].Content, part.Content)
			}
		}
		rebuilt.WriteString(string(current[overlap:]))
	}

	if got := rebuilt.String(); got != content {
		t.Errorf("reconstructed content = %q, want %q", got, content)
	}
}

func TestRecursiveChunkerUsesUnicodeCodePointsForLimits(t *testing.T) {
	chunker, err := NewRecursiveChunker(5, 1)
	if err != nil {
		t.Fatalf("NewRecursiveChunker() returned error: %v", err)
	}

	const content = "абвгдежзий"
	chunks, err := chunker.Chunk(context.Background(), Document{Content: content})
	if err != nil {
		t.Fatalf("Chunk() returned error: %v", err)
	}
	if len(chunks) < 2 {
		t.Fatalf("Chunk() returned %d chunks, want multiple", len(chunks))
	}

	for i, part := range chunks {
		if got := utf8.RuneCountInString(part.Content); got > 5 {
			t.Errorf("chunk %d has %d runes, exceeds configured size 5: %q", i, got, part.Content)
		}
		if !utf8.ValidString(part.Content) {
			t.Errorf("chunk %d contains invalid UTF-8: %q", i, part.Content)
		}
	}
}

func TestRecursiveChunkerPreservesDocumentMetadata(t *testing.T) {
	chunker, err := NewRecursiveChunker(12, 2)
	if err != nil {
		t.Fatalf("NewRecursiveChunker() returned error: %v", err)
	}
	metadata := map[string]any{"source": "guide.md", "chunk_index": 99, "chunk_count": 99}
	chunks, err := chunker.Chunk(context.Background(), Document{
		Content:  "first paragraph has several words\n\nsecond paragraph has others",
		Metadata: metadata,
	})
	if err != nil {
		t.Fatalf("Chunk() returned error: %v", err)
	}
	if len(chunks) == 0 {
		t.Fatal("Chunk() returned no chunks for non-empty document")
	}

	for i, part := range chunks {
		if got := part.Metadata["source"]; got != "guide.md" {
			t.Errorf("chunk %d source metadata = %v, want guide.md", i, got)
		}
		if got := part.Metadata["chunk_index"]; got != i {
			t.Errorf("chunk %d chunk_index = %v, want %d", i, got, i)
		}
		if got := part.Metadata["chunk_count"]; got != len(chunks) {
			t.Errorf("chunk %d chunk_count = %v, want %d", i, got, len(chunks))
		}
	}
	if metadata["chunk_index"] != 99 || metadata["chunk_count"] != 99 {
		t.Errorf("Chunk() mutated input metadata: %+v", metadata)
	}
}

func TestRecursiveChunkerSkipsEmptyContent(t *testing.T) {
	chunker, err := NewRecursiveChunker(20, 2)
	if err != nil {
		t.Fatalf("NewRecursiveChunker() returned error: %v", err)
	}

	for _, content := range []string{"", " \t\n "} {
		chunks, err := chunker.Chunk(context.Background(), Document{Content: content})
		if err != nil {
			t.Errorf("Chunk(%q) returned error: %v", content, err)
		}
		if len(chunks) != 0 {
			t.Errorf("Chunk(%q) returned %d chunks, want 0", content, len(chunks))
		}
	}
}

func TestRecursiveChunkerReturnsCanceledContext(t *testing.T) {
	chunker, err := NewRecursiveChunker(20, 2)
	if err != nil {
		t.Fatalf("NewRecursiveChunker() returned error: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	chunks, err := chunker.Chunk(ctx, Document{Content: "some input"})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Chunk() error = %v, want %v", err, context.Canceled)
	}
	if chunks != nil {
		t.Errorf("Chunk() = %v, want nil when context is canceled", chunks)
	}
}
