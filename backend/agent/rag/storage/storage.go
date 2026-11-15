package storage

import (
	"context"

	chunker "github.com/voznikaetnepriyazn/Hack-Change/agent/rag/chunk"
)

type Store interface {
	SaveBatch(ctx context.Context, chunks []chunker.Chunk) error
	Search(ctx context.Context, queryVec []float32, topK int) ([]chunker.Chunk, error)
}
