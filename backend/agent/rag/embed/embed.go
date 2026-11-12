package embed

import (
	"context"
	"errors"
	"net/url"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

const (
	defaultBaseURL = "https://integrate.api.nvidia.com/v1/embeddings"
	defaultModel   = "nvidia/nemotron-3.5-embed-1b"
	documentInput  = "passage"
	queryInput     = "query"
)

var (
	ErrInvalidConfig   = errors.New("invalid embedding client configuration")
	ErrEmptyInput      = errors.New("embedding input must not be empty")
	ErrInvalidResponse = errors.New("invalid embedding response")
	ErrEmbeddingCount  = errors.New("embedding response count does not match input count")
)

type Embedder struct {
	client  *openai.Client
	model   string
	baseURL string
}

// NewEmbedder creates a client for the NVIDIA NIM embeddings API.
// An empty baseURL uses the NVIDIA API endpoint.
func NewEmbedder(apiKey, baseURL, model string) *Embedder {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	model = strings.TrimSpace(model)
	if model == "" {
		model = defaultModel
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL

	return &Embedder{
		client:  openai.NewClientWithConfig(config),
		model:   model,
		baseURL: baseURL,
	}
}

type Embedding interface {
	Embed(ctx context.Context, text string) ([]float32, error)
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
	EmbedQuery(ctx context.Context, query string) ([]float32, error)
}

var _ Embedding = (*Embedder)(nil)

func (e *Embedder) validate() error {
	if e == nil || e.client == nil || e.model == "" || e.baseURL == "" {
		return ErrInvalidConfig
	}
	if !validHTTPSURL(e.baseURL) || strings.ContainsAny(e.model, "?#") {
		return ErrInvalidConfig
	}
	return nil
}

func validHTTPSURL(value string) bool {
	parsed, err := url.ParseRequestURI(value)
	return err == nil &&
		parsed.Scheme == "https" &&
		parsed.Host != "" &&
		parsed.RawQuery == "" &&
		parsed.Fragment == ""
}
