package embed

import (
	"context"
	"fmt"
	"math"

	openai "github.com/sashabaranov/go-openai"
)

func (e *Embedder) Embed(ctx context.Context, text string) ([]float32, error) {
	if err := e.validate(); err != nil {
		return nil, err
	}
	if text == "" {
		return nil, ErrEmptyInput
	}

	vectors, err := e.createEmbeddings(ctx, []string{text}, documentInput)
	if err != nil {
		return nil, err
	}
	return vectors[0], nil
}

func (e *Embedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if err := e.validate(); err != nil {
		return nil, err
	}
	if len(texts) == 0 {
		return nil, ErrEmptyInput
	}
	for i, text := range texts {
		if text == "" {
			return nil, fmt.Errorf("%w at index %d", ErrEmptyInput, i)
		}
	}

	return e.createEmbeddings(ctx, texts, documentInput)
}

func (e *Embedder) EmbedQuery(ctx context.Context, query string) ([]float32, error) {
	if err := e.validate(); err != nil {
		return nil, err
	}
	if query == "" {
		return nil, ErrEmptyInput
	}

	vectors, err := e.createEmbeddings(ctx, []string{query}, queryInput)
	if err != nil {
		return nil, err
	}
	return vectors[0], nil
}

func (e *Embedder) createEmbeddings(ctx context.Context, texts []string, inputType string) ([][]float32, error) {
	response, err := e.client.CreateEmbeddings(ctx, openai.EmbeddingRequestStrings{
		Input: texts,
		Model: openai.EmbeddingModel(e.model),
		ExtraBody: map[string]any{
			"input_type": inputType,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}

	if len(response.Data) != len(texts) {
		return nil, fmt.Errorf("%w: got %d vectors for %d inputs", ErrEmbeddingCount, len(response.Data), len(texts))
	}

	vectors := make([][]float32, len(texts))
	dimensions := 0
	for _, item := range response.Data {
		if item.Index < 0 || item.Index >= len(texts) {
			return nil, fmt.Errorf("%w: response index %d is out of range", ErrInvalidResponse, item.Index)
		}
		if vectors[item.Index] != nil {
			return nil, fmt.Errorf("%w: duplicate response index %d", ErrInvalidResponse, item.Index)
		}
		if err := validateVector(item.Embedding, item.Index); err != nil {
			return nil, err
		}
		if dimensions == 0 {
			dimensions = len(item.Embedding)
		} else if len(item.Embedding) != dimensions {
			return nil, fmt.Errorf("%w: input %d has %d dimensions, want %d", ErrInvalidResponse, item.Index, len(item.Embedding), dimensions)
		}
		vectors[item.Index] = item.Embedding
	}

	for i, vector := range vectors {
		if vector == nil {
			return nil, fmt.Errorf("%w: missing vector for input %d", ErrInvalidResponse, i)
		}
	}

	return vectors, nil
}

func validateVector(vector []float32, inputIndex int) error {
	if len(vector) == 0 {
		return fmt.Errorf("%w: no vector for input %d", ErrInvalidResponse, inputIndex)
	}

	for dimension, value := range vector {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return fmt.Errorf("%w: non-finite value at input %d, dimension %d", ErrInvalidResponse, inputIndex, dimension)
		}
	}
	return nil
}
