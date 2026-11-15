package answerquestions

import (
	"context"
	"errors"
	"strings"
	"testing"

	chunker "github.com/voznikaetnepriyazn/Hack-Change/agent/rag/chunk"
)

type fakeQueryEmbedder struct {
	query  string
	vector []float32
	err    error
}

func (f *fakeQueryEmbedder) EmbedQuery(_ context.Context, query string) ([]float32, error) {
	f.query = query
	return f.vector, f.err
}

type fakeChunkSearcher struct {
	vector []float32
	topK   int
	chunks []chunker.Chunk
	err    error
}

func (f *fakeChunkSearcher) Search(_ context.Context, vector []float32, topK int) ([]chunker.Chunk, error) {
	f.vector = vector
	f.topK = topK
	return f.chunks, f.err
}

func TestAnswerToolReturnsRelevantExcerptsAndSources(t *testing.T) {
	queryVector := []float32{0.25, 0.75}
	embedder := &fakeQueryEmbedder{vector: queryVector}
	searcher := &fakeChunkSearcher{chunks: []chunker.Chunk{
		{
			Content:  "  First relevant passage.  ",
			Metadata: map[string]any{"source": "playbook-master/policies.md"},
		},
		{Content: "Second relevant passage."},
	}}
	tool := NewAnswerTool(embedder, searcher)

	result, err := tool.Execute(context.Background(), `{"query":"  How does vacation work?  "}`)
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}
	if embedder.query != "How does vacation work?" {
		t.Errorf("embedded query = %q, want trimmed question", embedder.query)
	}
	if searcher.topK != topK {
		t.Errorf("search topK = %d, want %d", searcher.topK, topK)
	}
	if len(searcher.vector) != len(queryVector) ||
		searcher.vector[0] != queryVector[0] ||
		searcher.vector[1] != queryVector[1] {
		t.Errorf("search vector = %v, want %v", searcher.vector, queryVector)
	}
	for _, expected := range []string{
		"First relevant passage.",
		"Second relevant passage.",
		"Источник: playbook-master/policies.md",
	} {
		if !strings.Contains(result, expected) {
			t.Errorf("Execute() result does not contain %q: %s", expected, result)
		}
	}
}

func TestAnswerToolReportsNoMatches(t *testing.T) {
	tool := NewAnswerTool(
		&fakeQueryEmbedder{vector: []float32{0.25}},
		&fakeChunkSearcher{},
	)

	result, err := tool.Execute(context.Background(), `{"query":"unknown question"}`)
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}
	if !strings.Contains(result, "не найден") {
		t.Errorf("Execute() result = %q, want not-found message", result)
	}
}

func TestAnswerToolRejectsInvalidQueriesAndConfiguration(t *testing.T) {
	tool := NewAnswerTool(&fakeQueryEmbedder{}, &fakeChunkSearcher{})
	for _, args := range []string{`{"query":"  "}`, `{`} {
		if _, err := tool.Execute(context.Background(), args); err == nil {
			t.Errorf("Execute(%q) returned no error", args)
		}
	}

	if _, err := NewAnswerTool(nil, nil).Execute(context.Background(), `{"query":"question"}`); !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("Execute() with missing dependencies error = %v, want %v", err, ErrInvalidConfig)
	}
}

func TestAnswerToolPropagatesEmbedAndSearchErrors(t *testing.T) {
	embedErr := errors.New("embed unavailable")
	tool := NewAnswerTool(
		&fakeQueryEmbedder{err: embedErr},
		&fakeChunkSearcher{},
	)
	if _, err := tool.Execute(context.Background(), `{"query":"question"}`); !errors.Is(err, embedErr) {
		t.Errorf("Execute() embedding error = %v, want %v", err, embedErr)
	}

	searchErr := errors.New("database unavailable")
	tool = NewAnswerTool(
		&fakeQueryEmbedder{vector: []float32{0.5}},
		&fakeChunkSearcher{err: searchErr},
	)
	if _, err := tool.Execute(context.Background(), `{"query":"question"}`); !errors.Is(err, searchErr) {
		t.Errorf("Execute() search error = %v, want %v", err, searchErr)
	}
}
