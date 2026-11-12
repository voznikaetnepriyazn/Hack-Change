package storage

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	chunker "github.com/voznikaetnepriyazn/Hack-Change/agent/rag/chunk"
)

type fakeDatabase struct {
	tx          *fakeTransaction
	rows        pgx.Rows
	beginErr    error
	queryErr    error
	beginCalls  int
	queryCalls  int
	queryString string
	queryArgs   []any
}

func (f *fakeDatabase) Begin(context.Context) (transaction, error) {
	f.beginCalls++
	if f.beginErr != nil {
		return nil, f.beginErr
	}
	return f.tx, nil
}

func (f *fakeDatabase) Query(_ context.Context, query string, args ...any) (pgx.Rows, error) {
	f.queryCalls++
	f.queryString = query
	f.queryArgs = args
	if f.queryErr != nil {
		return nil, f.queryErr
	}
	return f.rows, nil
}

type execCall struct {
	query string
	args  []any
}

type fakeTransaction struct {
	calls       []execCall
	execErrAt   int
	execErr     error
	commitErr   error
	rollbackErr error
	committed   bool
	rolledBack  bool
}

func (f *fakeTransaction) Exec(_ context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	f.calls = append(f.calls, execCall{query: query, args: args})
	if f.execErrAt > 0 && len(f.calls) == f.execErrAt {
		return pgconn.CommandTag{}, f.execErr
	}
	return pgconn.NewCommandTag("INSERT 0 1"), nil
}

func (f *fakeTransaction) Commit(context.Context) error {
	f.committed = true
	return f.commitErr
}

func (f *fakeTransaction) Rollback(context.Context) error {
	f.rolledBack = true
	return f.rollbackErr
}

type fakeRow struct {
	id       uuid.UUID
	content  string
	metadata map[string]any
}

type fakeRows struct {
	rows       []fakeRow
	index      int
	scanErr    error
	rowsErr    error
	closed     bool
	closeCalls int
}

func (f *fakeRows) Close() {
	f.closed = true
	f.closeCalls++
}

func (f *fakeRows) Err() error                                   { return f.rowsErr }
func (f *fakeRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (f *fakeRows) FieldDescriptions() []pgconn.FieldDescription { return nil }

func (f *fakeRows) Next() bool {
	if f.index >= len(f.rows) {
		f.Close()
		return false
	}
	f.index++
	return true
}

func (f *fakeRows) Scan(dest ...any) error {
	if f.scanErr != nil {
		return f.scanErr
	}
	row := f.rows[f.index-1]
	id, ok := dest[0].(*uuid.UUID)
	if !ok {
		return fmt.Errorf("unexpected id scan destination %T", dest[0])
	}
	*id = row.id
	content, ok := dest[1].(*string)
	if !ok {
		return fmt.Errorf("unexpected content scan destination %T", dest[1])
	}
	*content = row.content
	metadata, ok := dest[2].(*map[string]any)
	if !ok {
		return fmt.Errorf("unexpected metadata scan destination %T", dest[2])
	}
	*metadata = row.metadata
	return nil
}

func (f *fakeRows) Values() ([]any, error) { return nil, nil }
func (f *fakeRows) RawValues() [][]byte    { return nil }
func (f *fakeRows) Conn() *pgx.Conn        { return nil }
func (f *fakeRows) TypeMap() *pgtype.Map   { return pgtype.NewMap() }

func newTestStore(db database) *pgStore {
	return &pgStore{db: db}
}

func validChunk(content string) chunker.Chunk {
	return chunker.Chunk{
		ID:        uuid.New(),
		Content:   content,
		Metadata:  map[string]any{"source": "test"},
		Embedding: make([]float32, EmbeddingDimensions),
	}
}

func TestSaveBatchUsesOneTransactionAndUpsertsChunks(t *testing.T) {
	tx := &fakeTransaction{}
	db := &fakeDatabase{tx: tx}
	store := newTestStore(db)
	chunks := []chunker.Chunk{validChunk("one"), validChunk("two")}

	if err := store.SaveBatch(context.Background(), chunks); err != nil {
		t.Fatalf("SaveBatch() error = %v", err)
	}
	if db.beginCalls != 1 {
		t.Errorf("Begin calls = %d, want 1", db.beginCalls)
	}
	if len(tx.calls) != len(chunks) {
		t.Fatalf("transaction Exec calls = %d, want %d", len(tx.calls), len(chunks))
	}
	if !tx.committed || tx.rolledBack {
		t.Errorf("transaction committed=%v rolledBack=%v, want committed only", tx.committed, tx.rolledBack)
	}
	if !strings.Contains(tx.calls[0].query, "ON CONFLICT (id) DO UPDATE") {
		t.Errorf("insert query does not upsert conflicting chunks: %s", tx.calls[0].query)
	}
	if got, ok := tx.calls[0].args[2].(interface{ Slice() []float32 }); !ok || len(got.Slice()) != EmbeddingDimensions {
		t.Errorf("stored embedding has unexpected type or dimensions: %T", tx.calls[0].args[2])
	}
}

func TestSaveBatchRollsBackWhenAChunkFails(t *testing.T) {
	saveErr := errors.New("database write failed")
	tx := &fakeTransaction{execErrAt: 2, execErr: saveErr}
	store := newTestStore(&fakeDatabase{tx: tx})

	err := store.SaveBatch(context.Background(), []chunker.Chunk{validChunk("one"), validChunk("two")})
	if !errors.Is(err, saveErr) {
		t.Errorf("SaveBatch() error = %v, want wrapped %v", err, saveErr)
	}
	if tx.committed {
		t.Error("transaction committed after an insert failure")
	}
	if !tx.rolledBack {
		t.Error("transaction was not rolled back after an insert failure")
	}
}

func TestSaveBatchRejectsInvalidChunksBeforeBeginningTransaction(t *testing.T) {
	db := &fakeDatabase{tx: &fakeTransaction{}}
	store := newTestStore(db)
	tests := []struct {
		name  string
		chunk chunker.Chunk
	}{
		{name: "missing id", chunk: chunker.Chunk{Content: "text", Embedding: make([]float32, EmbeddingDimensions)}},
		{name: "empty content", chunk: chunker.Chunk{ID: uuid.New(), Embedding: make([]float32, EmbeddingDimensions)}},
		{name: "wrong embedding dimension", chunk: chunker.Chunk{ID: uuid.New(), Content: "text", Embedding: []float32{1, 2}}},
		{name: "non-finite embedding", chunk: func() chunker.Chunk {
			c := validChunk("text")
			c.Embedding[0] = float32(math.NaN())
			return c
		}()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := store.SaveBatch(context.Background(), []chunker.Chunk{tt.chunk}); !errors.Is(err, ErrInvalidChunk) {
				t.Errorf("SaveBatch() error = %v, want %v", err, ErrInvalidChunk)
			}
		})
	}
	if db.beginCalls != 0 {
		t.Errorf("Begin calls = %d for invalid chunks, want 0", db.beginCalls)
	}
}

func TestSaveBatchEmptyAndNilStore(t *testing.T) {
	if err := newTestStore(&fakeDatabase{}).SaveBatch(context.Background(), nil); err != nil {
		t.Errorf("SaveBatch(nil) error = %v, want nil", err)
	}
	if err := (*pgStore)(nil).SaveBatch(context.Background(), []chunker.Chunk{validChunk("text")}); !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("SaveBatch() on nil store error = %v, want %v", err, ErrInvalidConfig)
	}
	if err := NewStore(nil).SaveBatch(context.Background(), []chunker.Chunk{validChunk("text")}); !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("SaveBatch() with nil pool error = %v, want %v", err, ErrInvalidConfig)
	}
}

func TestSearchValidatesInputBeforeQuery(t *testing.T) {
	db := &fakeDatabase{rows: &fakeRows{}}
	store := newTestStore(db)
	tests := []struct {
		name   string
		vector []float32
		topK   int
	}{
		{name: "empty vector", topK: 5},
		{name: "wrong dimension", vector: []float32{1, 2}, topK: 5},
		{name: "zero topK", vector: make([]float32, EmbeddingDimensions)},
		{name: "negative topK", vector: make([]float32, EmbeddingDimensions), topK: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := store.Search(context.Background(), tt.vector, tt.topK); !errors.Is(err, ErrInvalidQuery) {
				t.Errorf("Search() error = %v, want %v", err, ErrInvalidQuery)
			}
		})
	}
	if db.queryCalls != 0 {
		t.Errorf("Query calls = %d for invalid inputs, want 0", db.queryCalls)
	}
}

func TestSearchReturnsChunksAndClosesRows(t *testing.T) {
	id := uuid.New()
	rows := &fakeRows{rows: []fakeRow{{id: id, content: "matching content", metadata: map[string]any{"source": "doc"}}}}
	db := &fakeDatabase{rows: rows}
	store := newTestStore(db)
	vector := make([]float32, EmbeddingDimensions)

	got, err := store.Search(context.Background(), vector, 3)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("Search() returned %d chunks, want 1", len(got))
	}
	if got[0].ID != id || got[0].Content != "matching content" || got[0].Metadata["source"] != "doc" {
		t.Errorf("Search() chunk = %+v, want original ID/content/metadata", got[0])
	}
	if !strings.Contains(db.queryString, "ORDER BY embedding <=> $1") {
		t.Errorf("query does not use cosine distance ordering: %s", db.queryString)
	}
	if db.queryCalls != 1 || !rows.closed {
		t.Errorf("query calls=%d rows closed=%v, want one query and closed rows", db.queryCalls, rows.closed)
	}
}

func TestSearchPropagatesQueryAndIterationErrors(t *testing.T) {
	queryErr := errors.New("query failed")
	if _, err := newTestStore(&fakeDatabase{queryErr: queryErr}).Search(context.Background(), make([]float32, EmbeddingDimensions), 1); !errors.Is(err, queryErr) {
		t.Errorf("Search() query error = %v, want wrapped %v", err, queryErr)
	}

	rowsErr := errors.New("row iteration failed")
	rows := &fakeRows{rowsErr: rowsErr}
	if _, err := newTestStore(&fakeDatabase{rows: rows}).Search(context.Background(), make([]float32, EmbeddingDimensions), 1); !errors.Is(err, rowsErr) {
		t.Errorf("Search() rows error = %v, want wrapped %v", err, rowsErr)
	}
	if !rows.closed {
		t.Error("rows were not closed after iteration error")
	}
}
