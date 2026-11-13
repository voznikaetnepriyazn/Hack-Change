package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
)

const maxEmbeddingResponseBytes = 16 << 20

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

	const batchSize = 16
	result := make([][]float32, 0, len(texts))

	for start := 0; start < len(texts); start += batchSize {
		end := start + batchSize
		if end > len(texts) {
			end = len(texts)
		}
		vectors, err := e.createEmbeddings(ctx, texts[start:end], documentInput)
		if err != nil {
			return nil, err
		}
		result = append(result, vectors...)
	}

	return result, nil
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
	payload := struct {
		Input          []string `json:"input"`
		Model          string   `json:"model"`
		InputType      string   `json:"input_type"`
		Modality       []string `json:"modality"`
		EncodingFormat string   `json:"encoding_format"`
		Dimensions     int      `json:"dimensions"`
		Truncate       string   `json:"truncate"`
		EmbeddingType  string   `json:"embedding_type"`
	}{
		Input:          texts,
		Model:          e.model,
		InputType:      inputType,
		Modality:       []string{"text"},
		EncodingFormat: "float",
		Dimensions:     outputDimensions,
		Truncate:       "END",
		EmbeddingType:  "float",
	}
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode embedding request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		e.baseURL+"/embeddings",
		bytes.NewReader(requestBody),
	)
	if err != nil {
		return nil, fmt.Errorf("create embedding request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, maxEmbeddingResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read embedding response: %w", err)
	}
	if int64(len(responseBody)) > maxEmbeddingResponseBytes {
		return nil, fmt.Errorf("embedding response exceeds %d bytes", maxEmbeddingResponseBytes)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf(
			"embedding request failed: HTTP %d %s: %s",
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			strings.TrimSpace(string(responseBody)),
		)
	}

	var response struct {
		Data []struct {
			Index     int       `json:"index"`
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("decode embedding response: %w", err)
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
		vector, err := truncateAndNormalizeEmbedding(item.Embedding)
		if err != nil {
			return nil, fmt.Errorf("normalize embedding for input %d: %w", item.Index, err)
		}
		vectors[item.Index] = vector
	}

	for i, vector := range vectors {
		if vector == nil {
			return nil, fmt.Errorf("%w: missing vector for input %d", ErrInvalidResponse, i)
		}
	}

	return vectors, nil
}

// NVIDIA documents prefix truncation as supported when followed by L2 normalization.
func truncateAndNormalizeEmbedding(vector []float32) ([]float32, error) {
	if len(vector) <= outputDimensions {
		return vector, nil
	}

	truncated := append([]float32(nil), vector[:outputDimensions]...)
	var normSquared float64
	for _, value := range truncated {
		normSquared += float64(value) * float64(value)
	}
	if normSquared == 0 {
		return nil, fmt.Errorf("%w: zero vector after truncation", ErrInvalidResponse)
	}

	norm := float32(math.Sqrt(normSquared))
	for i := range truncated {
		truncated[i] /= norm
	}
	return truncated, nil
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
