package pipeline

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	chunker "github.com/voznikaetnepriyazn/Hack-Change/agent/rag/chunk"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/rag/storage"
)

type fakeChunker struct {
	chunks []chunker.Chunk
	err    error
	calls  int
}

func (f *fakeChunker) Chunk(context.Context, chunker.Document) ([]chunker.Chunk, error) {
	f.calls++
	return f.chunks, f.err
}

type fakeEmbedder struct {
	vectors [][]float32
	err     error
	inputs  []string
	calls   int
}

func (f *fakeEmbedder) Embed(context.Context, string) ([]float32, error) {
	panic("Embed should not be called by document ingestion")
}

func (f *fakeEmbedder) EmbedQuery(context.Context, string) ([]float32, error) {
	panic("EmbedQuery should not be called by document ingestion")
}

func (f *fakeEmbedder) EmbedBatch(_ context.Context, texts []string) ([][]float32, error) {
	f.calls++
	f.inputs = append([]string(nil), texts...)
	return f.vectors, f.err
}

type fakeStore struct {
	chunks []chunker.Chunk
	err    error
	calls  int
}

func (f *fakeStore) SaveBatch(_ context.Context, chunks []chunker.Chunk) error {
	f.calls++
	f.chunks = append([]chunker.Chunk(nil), chunks...)
	return f.err
}

func (f *fakeStore) Search(context.Context, []float32, int) ([]chunker.Chunk, error) {
	return nil, errors.New("unexpected Search call")
}

var _ storage.Store = (*fakeStore)(nil)

func TestIngestDocumentEmbedsAndSavesAllChunksAsBatch(t *testing.T) {
	chunks := []chunker.Chunk{
		{ID: uuid.New(), Content: "first"},
		{ID: uuid.New(), Content: "second"},
	}
	ch := &fakeChunker{chunks: chunks}
	emb := &fakeEmbedder{vectors: [][]float32{{1, 2}, {3, 4}}}
	store := &fakeStore{}
	p := NewPipeline(ch, emb, store)

	err := p.IngestDocument(context.Background(), chunker.Document{Content: "document"})
	if err != nil {
		t.Fatalf("IngestDocument() error = %v", err)
	}
	if !reflect.DeepEqual(emb.inputs, []string{"first", "second"}) {
		t.Errorf("embedded texts = %v, want [first second]", emb.inputs)
	}
	if store.calls != 1 {
		t.Fatalf("SaveBatch calls = %d, want 1", store.calls)
	}
	if len(store.chunks) != 2 {
		t.Fatalf("saved chunks = %d, want 2", len(store.chunks))
	}
	for i := range chunks {
		if !reflect.DeepEqual(store.chunks[i].Embedding, emb.vectors[i]) {
			t.Errorf("chunk %d embedding = %v, want %v", i, store.chunks[i].Embedding, emb.vectors[i])
		}
	}
}

func TestIngestDocumentStopsAtFirstFailedStage(t *testing.T) {
	chunkErr := errors.New("chunk failure")
	embedErr := errors.New("embed failure")
	storeErr := errors.New("storage failure")

	tests := []struct {
		name        string
		chunker     *fakeChunker
		embedder    *fakeEmbedder
		store       *fakeStore
		wantErr     error
		wantMessage string
		wantEmbeds  int
		wantSaves   int
	}{
		{
			name:        "chunk failure",
			chunker:     &fakeChunker{err: chunkErr},
			embedder:    &fakeEmbedder{},
			store:       &fakeStore{},
			wantErr:     chunkErr,
			wantMessage: "chunk failed",
		},
		{
			name:        "embed failure",
			chunker:     &fakeChunker{chunks: []chunker.Chunk{{ID: uuid.New(), Content: "text"}}},
			embedder:    &fakeEmbedder{err: embedErr},
			store:       &fakeStore{},
			wantErr:     embedErr,
			wantMessage: "embed failed",
			wantEmbeds:  1,
		},
		{
			name:        "save failure",
			chunker:     &fakeChunker{chunks: []chunker.Chunk{{ID: uuid.New(), Content: "text"}}},
			embedder:    &fakeEmbedder{vectors: [][]float32{{1}}},
			store:       &fakeStore{err: storeErr},
			wantErr:     storeErr,
			wantMessage: "save chunks failed",
			wantEmbeds:  1,
			wantSaves:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPipeline(tt.chunker, tt.embedder, tt.store)
			err := p.IngestDocument(context.Background(), chunker.Document{Content: "document"})
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("IngestDocument() error = %v, want wrapped %v", err, tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantMessage) {
				t.Errorf("error = %q, want it to contain %q", err, tt.wantMessage)
			}
			if tt.embedder.calls != tt.wantEmbeds {
				t.Errorf("EmbedBatch calls = %d, want %d", tt.embedder.calls, tt.wantEmbeds)
			}
			if tt.store.calls != tt.wantSaves {
				t.Errorf("SaveBatch calls = %d, want %d", tt.store.calls, tt.wantSaves)
			}
		})
	}
}

func TestIngestDocumentHandlesEmptyChunksWithoutCallingLaterStages(t *testing.T) {
	emb := &fakeEmbedder{}
	store := &fakeStore{}
	p := NewPipeline(&fakeChunker{}, emb, store)

	if err := p.IngestDocument(context.Background(), chunker.Document{}); err != nil {
		t.Fatalf("IngestDocument() error = %v", err)
	}
	if emb.calls != 0 || store.calls != 0 {
		t.Errorf("later stages called: EmbedBatch=%d SaveBatch=%d, want both 0", emb.calls, store.calls)
	}
}

func TestIngestDocumentRejectsInvalidEmbeddingResultsBeforeSaving(t *testing.T) {
	tests := []struct {
		name    string
		vectors [][]float32
		wantErr error
	}{
		{name: "wrong vector count", vectors: nil, wantErr: ErrEmbeddingCount},
		{name: "empty vector", vectors: [][]float32{{}}, wantErr: ErrInvalidEmbedding},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := &fakeChunker{chunks: []chunker.Chunk{{ID: uuid.New(), Content: "text"}}}
			store := &fakeStore{}
			p := NewPipeline(ch, &fakeEmbedder{vectors: tt.vectors}, store)

			err := p.IngestDocument(context.Background(), chunker.Document{Content: "document"})
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("IngestDocument() error = %v, want %v", err, tt.wantErr)
			}
			if store.calls != 0 {
				t.Errorf("SaveBatch calls = %d, want 0 for invalid embeddings", store.calls)
			}
		})
	}
}

func TestIngestDocumentRejectsMissingDependencies(t *testing.T) {
	tests := []struct {
		name string
		p    *Pipeline
	}{
		{name: "nil pipeline"},
		{name: "nil chunker", p: &Pipeline{embedder: &fakeEmbedder{}, storage: &fakeStore{}}},
		{name: "nil embedder", p: &Pipeline{chunker: &fakeChunker{}, storage: &fakeStore{}}},
		{name: "nil storage", p: &Pipeline{chunker: &fakeChunker{}, embedder: &fakeEmbedder{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.IngestDocument(context.Background(), chunker.Document{}); !errors.Is(err, ErrInvalidConfig) {
				t.Errorf("IngestDocument() error = %v, want %v", err, ErrInvalidConfig)
			}
		})
	}
}
