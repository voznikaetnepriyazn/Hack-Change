package chunker

import (
	"context"
	"crypto/sha256"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/hupe1980/agentmesh/pkg/loader"
)

var ErrInvalidConfig = errors.New("invalid recursive chunker configuration")

// RecursiveChunker implements Chunker using agentmesh's recursive character splitter.
type RecursiveChunker struct {
	splitter *loader.RecursiveCharacterSplitter
	overlap  int
}

var _ Chunker = (*RecursiveChunker)(nil)

// NewRecursiveChunker creates a chunker with a maximum final chunk size and overlap,
// measured in Unicode code points.
func NewRecursiveChunker(chunkSize, overlap int) (*RecursiveChunker, error) {
	if chunkSize <= 0 || overlap < 0 || overlap >= chunkSize {
		return nil, ErrInvalidConfig
	}

	splitter := loader.NewRecursiveCharacterSplitter(
		chunkSize-overlap,
		0,
		loader.WithSeparators([]string{"\n\n", "\n", ". ", " ", ""}),
	)

	return &RecursiveChunker{
		splitter: splitter,
		overlap:  overlap,
	}, nil
}

// Chunk splits a document and adds stable IDs and chunk position metadata.
func (c *RecursiveChunker) Chunk(ctx context.Context, doc Document) ([]Chunk, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	libDocs := []loader.Document{
		{
			Content:  doc.Content,
			Metadata: doc.Metadata,
		},
	}

	splitDocs, err := c.splitter.SplitDocuments(libDocs)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var chunks []Chunk
	var previousText string
	for _, sd := range splitDocs {
		text := strings.TrimSpace(sd.Content)
		if text == "" {
			continue
		}

		chunkContent := text
		if previousText != "" && c.overlap > 0 {
			chunkContent = suffixRunes(previousText, c.overlap) + chunkContent
		}

		metadata := cloneMetadata(sd.Metadata)
		idContent := chunkContent
		if source, ok := metadata["source"].(string); ok && source != "" {
			idContent = source + "\x00" + chunkContent
		}
		hash := sha256.Sum256([]byte(idContent))
		id := uuid.NewSHA1(uuid.NameSpaceOID, hash[:])

		chunks = append(chunks, Chunk{
			ID:       id,
			Content:  chunkContent,
			Metadata: metadata,
		})
		previousText = text
	}

	for i := range chunks {
		if chunks[i].Metadata == nil {
			chunks[i].Metadata = make(map[string]any)
		}
		chunks[i].Metadata["chunk_index"] = i
		chunks[i].Metadata["chunk_count"] = len(chunks)
	}

	return chunks, nil
}

// suffixRunes returns the last count runes of text, or the entire text if it has fewer than count runes.
func suffixRunes(text string, count int) string {
	runes := []rune(text)
	if len(runes) > count {
		runes = runes[len(runes)-count:]
	}
	return string(runes)
}

// cloneMetadata creates a shallow copy of the metadata map, or returns nil if the input is nil.
func cloneMetadata(metadata map[string]any) map[string]any {
	if metadata == nil {
		return nil
	}
	cloned := make(map[string]any, len(metadata))
	for key, value := range metadata {
		cloned[key] = value
	}
	return cloned
}
