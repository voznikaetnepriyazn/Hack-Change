package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	pgvector "github.com/pgvector/pgvector-go"

	chunker "github.com/voznikaetnepriyazn/Hack-Change/agent/rag/chunk"
)

const EmbeddingDimensions = 1024

var (
	ErrInvalidConfig = errors.New("invalid PostgreSQL RAG storage configuration")
	ErrInvalidChunk  = errors.New("invalid chunk")
	ErrInvalidQuery  = errors.New("invalid vector search query")
)

type commandExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type queryExecutor interface {
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
}

type transaction interface {
	commandExecutor
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type database interface {
	queryExecutor
	Begin(ctx context.Context) (transaction, error)
}

type poolAdapter struct {
	pool *pgxpool.Pool
}

func (p poolAdapter) Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error) {
	return p.pool.Query(ctx, sql, arguments...)
}

func (p poolAdapter) Begin(ctx context.Context) (transaction, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return pgxTransaction{Tx: tx}, nil
}

type pgxTransaction struct {
	pgx.Tx
}

type pgStore struct {
	db database
}

func NewStore(db *pgxpool.Pool) Store {
	if db == nil {
		return &pgStore{}
	}
	return &pgStore{db: poolAdapter{pool: db}}
}

func (s *pgStore) SaveBatch(ctx context.Context, chunks []chunker.Chunk) error {
	if s == nil || s.db == nil {
		return ErrInvalidConfig
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("saving chunks canceled: %w", err)
	}
	for i, chunk := range chunks {
		if err := validateChunk(chunk); err != nil {
			return fmt.Errorf("chunk %d: %w", i, err)
		}
	}
	if len(chunks) == 0 {
		return nil
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning chunk save transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			if err := tx.Rollback(context.Background()); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
				slog.Error("rolling back chunk save transaction failed", "error", err)
			}
		}
	}()

	for _, chunk := range chunks {
		metadata := chunk.Metadata
		if metadata == nil {
			metadata = make(map[string]any)
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO chunks (id, content, embedding, metadata)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO UPDATE
			SET content = EXCLUDED.content,
			    embedding = EXCLUDED.embedding,
			    metadata = EXCLUDED.metadata
		`, chunk.ID, chunk.Content, pgvector.NewVector(chunk.Embedding), metadata)
		if err != nil {
			return fmt.Errorf("saving chunk %s: %w", chunk.ID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing chunk save transaction: %w", err)
	}
	committed = true
	return nil
}

func (s *pgStore) Search(ctx context.Context, queryVec []float32, topK int) ([]chunker.Chunk, error) {
	if s == nil || s.db == nil {
		return nil, ErrInvalidConfig
	}
	if err := validateVector(queryVec); err != nil || topK <= 0 {
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidQuery, err)
		}
		return nil, fmt.Errorf("%w: topK must be positive", ErrInvalidQuery)
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("search canceled: %w", err)
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, content, metadata
		FROM chunks
		WHERE embedding IS NOT NULL
		ORDER BY embedding <=> $1
		LIMIT $2
	`, pgvector.NewVector(queryVec), topK)
	if err != nil {
		return nil, fmt.Errorf("querying chunks: %w", err)
	}
	defer rows.Close()

	var chunks []chunker.Chunk
	for rows.Next() {
		var c chunker.Chunk
		var metadata map[string]any
		if err := rows.Scan(&c.ID, &c.Content, &metadata); err != nil {
			return nil, fmt.Errorf("scanning chunk row: %w", err)
		}
		if metadata == nil {
			metadata = make(map[string]any)
		}
		c.Metadata = metadata
		chunks = append(chunks, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating chunk rows: %w", err)
	}

	slog.Debug("RAG chunk search completed", "chunks", len(chunks))
	return chunks, nil
}

func validateChunk(chunk chunker.Chunk) error {
	if chunk.ID == uuid.Nil {
		return fmt.Errorf("%w: id must not be nil", ErrInvalidChunk)
	}
	if strings.TrimSpace(chunk.Content) == "" {
		return fmt.Errorf("%w: content must not be empty", ErrInvalidChunk)
	}
	if err := validateVector(chunk.Embedding); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidChunk, err)
	}
	return nil
}

func validateVector(vector []float32) error {
	if len(vector) != EmbeddingDimensions {
		return fmt.Errorf("vector has %d dimensions, want %d", len(vector), EmbeddingDimensions)
	}
	for i, value := range vector {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return fmt.Errorf("vector contains a non-finite value at dimension %d", i)
		}
	}
	return nil
}
