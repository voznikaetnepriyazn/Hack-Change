package pipeline

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	chunker "github.com/voznikaetnepriyazn/Hack-Change/agent/rag/chunk"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/rag/embed"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/rag/storage"
)

type Pipeline struct {
	chunker  chunker.Chunker
	embedder embed.Embedding
	storage  storage.Store
}

var (
	ErrInvalidConfig    = errors.New("invalid RAG pipeline configuration")
	ErrEmbeddingCount   = errors.New("embedding response count does not match chunk count")
	ErrInvalidEmbedding = errors.New("invalid embedding vector")
)

func NewPipeline(chunker chunker.Chunker, embedder embed.Embedding, storage storage.Store) *Pipeline {
	return &Pipeline{
		chunker:  chunker,
		embedder: embedder,
		storage:  storage,
	}
}

// IngestDocument processes a document by chunking it, generating embeddings for each chunk, and storing the chunks with their embeddings in the storage backend.
func (p *Pipeline) IngestDocument(ctx context.Context, doc chunker.Document) error {
	if p == nil || p.chunker == nil || p.embedder == nil || p.storage == nil {
		return ErrInvalidConfig
	}

	// 1. Режем
	chunks, err := p.chunker.Chunk(ctx, doc)
	if err != nil {
		return fmt.Errorf("chunk failed: %w", err)
	}

	if len(chunks) == 0 {
		slog.Info("no chunks generated")
		return nil
	}

	// 2. Собираем текст для батча
	texts := make([]string, len(chunks))
	for i, c := range chunks {
		texts[i] = c.Content
	}

	// 3. Получаем эмбеддинги (батчем!)
	embeddings, err := p.embedder.EmbedBatch(ctx, texts)
	if err != nil {
		slog.Error("embed failed", "error", err)
		return fmt.Errorf("embed failed: %w", err)
	}

	if len(embeddings) != len(chunks) {
		slog.Error("embeddings count mismatch", "expected", len(chunks), "actual", len(embeddings))
		return fmt.Errorf("%w: got %d vectors for %d chunks", ErrEmbeddingCount, len(embeddings), len(chunks))
	}

	// 4. Attach vectors, then persist the complete document atomically.
	for i := range chunks {
		if len(embeddings[i]) == 0 {
			return fmt.Errorf("%w: no vector for chunk %d", ErrInvalidEmbedding, i)
		}
		chunks[i].Embedding = embeddings[i]
	}

	if err := p.storage.SaveBatch(ctx, chunks); err != nil {
		return fmt.Errorf("save chunks failed: %w", err)
	}
	return nil
}
