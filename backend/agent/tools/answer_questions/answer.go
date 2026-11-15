package answerquestions

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	parser "github.com/voznikaetnepriyazn/Hack-Change/agent/internal/lib/parse"
	chunker "github.com/voznikaetnepriyazn/Hack-Change/agent/rag/chunk"
)

const topK = 4

//go:embed answer.md
var description string

var ErrInvalidConfig = errors.New("invalid answer questions tool configuration")

type QueryEmbedder interface {
	EmbedQuery(context.Context, string) ([]float32, error)
}

type ChunkSearcher interface {
	Search(context.Context, []float32, int) ([]chunker.Chunk, error)
}

type AnswerTool struct {
	embedder QueryEmbedder
	searcher ChunkSearcher
}

var _ interface {
	Name() string
	Description() string
	Parameters() json.RawMessage
	Execute(context.Context, string) (string, error)
} = (*AnswerTool)(nil)

func NewAnswerTool(embedder QueryEmbedder, searcher ChunkSearcher) *AnswerTool {
	return &AnswerTool{
		embedder: embedder,
		searcher: searcher,
	}
}

func (t *AnswerTool) Name() string {
	return "answer_questions"
}

func (t *AnswerTool) Description() string {
	return description
}

func (t *AnswerTool) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"query": {
				"type": "string",
				"description": "Полный вопрос стажёра в свободной форме, на который нужно найти ответ в базе знаний."
			}
		},
		"required": ["query"],
		"additionalProperties": false
	}`)
}

type searchArgs struct {
	Query string `json:"query"`
}

func (t *AnswerTool) Execute(ctx context.Context, argsJSON string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("answer question canceled: %w", err)
	}
	if t == nil || t.embedder == nil || t.searcher == nil {
		return "", ErrInvalidConfig
	}

	args, err := parser.ParseArgs[searchArgs](argsJSON)
	if err != nil {
		slog.Error("Failed to parse answer question arguments", "error", err)
		return "", fmt.Errorf("failed to parse answer question arguments: %w", err)
	}
	query := strings.TrimSpace(args.Query)
	if query == "" {
		return "", errors.New("question query is required")
	}

	queryVector, err := t.embedder.EmbedQuery(ctx, query)
	if err != nil {
		slog.Error("Failed to embed answer question", "error", err)
		return "", fmt.Errorf("failed to embed answer question: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("answer question canceled: %w", err)
	}

	chunks, err := t.searcher.Search(ctx, queryVector, topK)
	if err != nil {
		slog.Error("Failed to search RAG chunks", "error", err)
		return "", fmt.Errorf("failed to search knowledge base: %w", err)
	}
	if len(chunks) == 0 {
		return "В базе знаний не найден подходящий материал для ответа на этот вопрос.", nil
	}

	var result strings.Builder
	result.WriteString("Релевантные фрагменты из базы знаний:\n")
	resultCount := 0
	for _, chunk := range chunks {
		content := strings.TrimSpace(chunk.Content)
		if content == "" {
			continue
		}
		resultCount++
		result.WriteString("\n")
		if source, ok := chunk.Metadata["source"].(string); ok && source != "" {
			fmt.Fprintf(&result, "[%d] Источник: %s\n", resultCount, source)
		} else {
			fmt.Fprintf(&result, "[%d]\n", resultCount)
		}
		result.WriteString(content)
		result.WriteString("\n")
	}
	if resultCount == 0 {
		return "В базе знаний не найден подходящий материал для ответа на этот вопрос.", nil
	}
	return strings.TrimSpace(result.String()), nil
}