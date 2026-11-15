package chunker

import (
	"context"

	"github.com/google/uuid"
)

type Chunk struct {
	ID        uuid.UUID
	Content   string
	Metadata  map[string]any
	Embedding []float32
}

func NewChunk(id uuid.UUID, content string, metadata map[string]any) *Chunk {
	return &Chunk{
		ID:        id,
		Content:   content,
		Metadata:  metadata,
		Embedding: nil,
	}
}

type Document struct {
	Content  string
	Metadata map[string]any
}

type Chunker interface {
	Chunk(ctx context.Context, doc Document) ([]Chunk, error)
}
