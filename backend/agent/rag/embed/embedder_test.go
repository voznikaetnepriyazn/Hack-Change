package embed

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"strings"
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newTestEmbedder(baseURL, token, model string, client openai.HTTPDoer) *Embedder {
	config := openai.DefaultConfig(token)
	config.BaseURL = strings.TrimRight(baseURL, "/")
	config.HTTPClient = client
	return &Embedder{
		client:  openai.NewClientWithConfig(config),
		model:   model,
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestEmbedSendsNVIDIAEmbeddingsRequest(t *testing.T) {
	const (
		baseURL = "https://integrate.api.nvidia.com/v1"
		model   = "nvidia/nv-embedqa-e5-v5"
		token   = "test-api-key"
	)
	client := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost {
			t.Errorf("HTTP method = %q, want POST", req.Method)
		}
		if got, want := req.URL.String(), baseURL+"/embeddings"; got != want {
			t.Errorf("request URL = %q, want %q", got, want)
		}
		if got, want := req.Header.Get("Authorization"), "Bearer "+token; got != want {
			t.Errorf("Authorization = %q, want %q", got, want)
		}
		if got := req.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", got)
		}

		var payload struct {
			Input          []string `json:"input"`
			Model          string   `json:"model"`
			InputType      string   `json:"input_type"`
			Modality       []string `json:"modality"`
			Dimensions     int      `json:"dimensions"`
			EncodingFormat string   `json:"encoding_format"`
			Truncate       string   `json:"truncate"`
			EmbeddingType  string   `json:"embedding_type"`
		}
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if len(payload.Input) != 1 || payload.Input[0] != "hello world" {
			t.Errorf("request input = %v, want [hello world]", payload.Input)
		}
		if payload.Model != model {
			t.Errorf("request model = %q, want %q", payload.Model, model)
		}
		if payload.InputType != "passage" {
			t.Errorf("request input_type = %q, want passage", payload.InputType)
		}
		if len(payload.Modality) != 1 || payload.Modality[0] != "text" {
			t.Errorf("request modality = %v, want [text]", payload.Modality)
		}
		if payload.Dimensions != outputDimensions {
			t.Errorf("request dimensions = %d, want %d", payload.Dimensions, outputDimensions)
		}
		if payload.EncodingFormat != "float" {
			t.Errorf("request encoding_format = %q, want float", payload.EncodingFormat)
		}
		if payload.Truncate != "END" {
			t.Errorf("request truncate = %q, want END", payload.Truncate)
		}
		if payload.EmbeddingType != "float" {
			t.Errorf("request embedding_type = %q, want float", payload.EmbeddingType)
		}
		return jsonResponse(http.StatusOK, `{"data":[{"index":0,"embedding":[0.1,0.2,0.3]}]}`), nil
	})
	embedder := newTestEmbedder(baseURL, token, model, client)

	got, err := embedder.Embed(context.Background(), "hello world")
	if err != nil {
		t.Fatalf("Embed() returned error: %v", err)
	}
	want := []float32{0.1, 0.2, 0.3}
	if !equalVector(got, want) {
		t.Errorf("Embed() = %v, want %v", got, want)
	}
}

func TestEmbedBatchReturnsVectorsInInputOrder(t *testing.T) {
	client := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var payload struct {
			Input     []string `json:"input"`
			InputType string   `json:"input_type"`
		}
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if len(payload.Input) != 2 || payload.Input[0] != "query one" || payload.Input[1] != "query two" {
			t.Errorf("request input = %v, want [query one query two]", payload.Input)
		}
		if payload.InputType != "passage" {
			t.Errorf("request input_type = %q, want passage", payload.InputType)
		}
		return jsonResponse(http.StatusOK, `{"data":[{"index":1,"embedding":[3,4]},{"index":0,"embedding":[1,2]}]}`), nil
	})
	embedder := newTestEmbedder(defaultBaseURL, "token", defaultModel, client)

	got, err := embedder.EmbedBatch(context.Background(), []string{"query one", "query two"})
	if err != nil {
		t.Fatalf("EmbedBatch() returned error: %v", err)
	}
	want := [][]float32{{1, 2}, {3, 4}}
	if len(got) != len(want) {
		t.Fatalf("EmbedBatch() returned %d vectors, want %d", len(got), len(want))
	}
	for i := range want {
		if !equalVector(got[i], want[i]) {
			t.Errorf("EmbedBatch()[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestEmbedBatchTruncatesAndRenormalizes2048DimensionalVectors(t *testing.T) {
	rawVector := make([]float32, 2048)
	for i := range rawVector {
		rawVector[i] = 100
		if i < outputDimensions {
			rawVector[i] = 2
		}
	}
	body, err := json.Marshal(map[string]any{
		"data": []map[string]any{{"index": 0, "embedding": rawVector}},
	})
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	client := roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusOK, string(body)), nil
	})
	embedder := newTestEmbedder(defaultBaseURL, "token", defaultModel, client)

	vectors, err := embedder.EmbedBatch(context.Background(), []string{"document"})
	if err != nil {
		t.Fatalf("EmbedBatch() returned error: %v", err)
	}
	if got := len(vectors[0]); got != outputDimensions {
		t.Fatalf("embedding dimensions = %d, want %d", got, outputDimensions)
	}

	var normSquared float64
	for _, value := range vectors[0] {
		normSquared += float64(value) * float64(value)
	}
	if math.Abs(math.Sqrt(normSquared)-1) > 1e-6 {
		t.Errorf("truncated embedding norm = %f, want 1", math.Sqrt(normSquared))
	}
	if got, want := vectors[0][0], float32(1/math.Sqrt(outputDimensions)); math.Abs(float64(got-want)) > 1e-6 {
		t.Errorf("first truncated component = %f, want %f", got, want)
	}
}

func TestEmbedQueryUsesQueryInputType(t *testing.T) {
	client := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var payload struct {
			InputType string `json:"input_type"`
		}
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if payload.InputType != "query" {
			t.Errorf("request input_type = %q, want query", payload.InputType)
		}
		return jsonResponse(http.StatusOK, `{"data":[{"index":0,"embedding":[0.5,0.6]}]}`), nil
	})
	embedder := newTestEmbedder(defaultBaseURL, "token", defaultModel, client)

	got, err := embedder.EmbedQuery(context.Background(), "find something")
	if err != nil {
		t.Fatalf("EmbedQuery() returned error: %v", err)
	}
	if !equalVector(got, []float32{0.5, 0.6}) {
		t.Errorf("EmbedQuery() = %v, want [0.5 0.6]", got)
	}
}

func TestEmbedReportsNVIDIAAPIError(t *testing.T) {
	client := roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusUnauthorized, `{"error":{"message":"Incorrect API key provided","type":"invalid_request_error"}}`), nil
	})
	embedder := newTestEmbedder(defaultBaseURL, "bad-token", defaultModel, client)

	_, err := embedder.Embed(context.Background(), "hello")
	if err == nil || !strings.Contains(err.Error(), "Incorrect API key provided") {
		t.Errorf("Embed() error = %v, want API key error", err)
	}
}

func TestEmbedRejectsInvalidOrMismatchedResponses(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		batch    bool
		inputs   []string
		wantErr  error
		contains string
	}{
		{
			name:     "empty response",
			body:     `{"data":[]}`,
			inputs:   []string{"hello"},
			wantErr:  ErrEmbeddingCount,
			contains: "got 0 vectors",
		},
		{
			name:     "empty vector",
			body:     `{"data":[{"index":0,"embedding":[]}]}`,
			inputs:   []string{"hello"},
			wantErr:  ErrInvalidResponse,
			contains: "no vector",
		},
		{
			name:     "batch count mismatch",
			body:     `{"data":[{"index":0,"embedding":[1,2]}]}`,
			batch:    true,
			inputs:   []string{"one", "two"},
			wantErr:  ErrEmbeddingCount,
			contains: "got 1 vectors for 2 inputs",
		},
		{
			name:     "malformed response",
			body:     `{"data":[`,
			inputs:   []string{"hello"},
			contains: "unexpected EOF",
		},
		{
			name:     "inconsistent dimensions",
			body:     `{"data":[{"index":0,"embedding":[1,2]},{"index":1,"embedding":[3]}]}`,
			batch:    true,
			inputs:   []string{"one", "two"},
			wantErr:  ErrInvalidResponse,
			contains: "has 1 dimensions, want 2",
		},
		{
			name:     "response index out of range",
			body:     `{"data":[{"index":1,"embedding":[1]}]}`,
			inputs:   []string{"hello"},
			wantErr:  ErrInvalidResponse,
			contains: "out of range",
		},
		{
			name:     "duplicate response index",
			body:     `{"data":[{"index":0,"embedding":[1]},{"index":0,"embedding":[2]}]}`,
			batch:    true,
			inputs:   []string{"one", "two"},
			wantErr:  ErrInvalidResponse,
			contains: "duplicate response index",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := roundTripFunc(func(*http.Request) (*http.Response, error) {
				return jsonResponse(http.StatusOK, tt.body), nil
			})
			embedder := newTestEmbedder(defaultBaseURL, "token", defaultModel, client)

			var err error
			if tt.batch {
				_, err = embedder.EmbedBatch(context.Background(), tt.inputs)
			} else {
				_, err = embedder.Embed(context.Background(), tt.inputs[0])
			}
			if err == nil {
				t.Fatal("embedding call returned nil error for invalid response")
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.contains) {
				t.Errorf("error = %q, want it to contain %q", err, tt.contains)
			}
		})
	}
}

func TestEmbedRejectsInvalidInputsBeforeCallingProvider(t *testing.T) {
	var calls int
	client := roundTripFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return jsonResponse(http.StatusOK, `{"data":[{"index":0,"embedding":[1]}]}`), nil
	})
	embedder := newTestEmbedder(defaultBaseURL, "token", defaultModel, client)

	if _, err := embedder.Embed(context.Background(), ""); !errors.Is(err, ErrEmptyInput) {
		t.Errorf("Embed(empty) error = %v, want %v", err, ErrEmptyInput)
	}
	if _, err := embedder.EmbedBatch(context.Background(), nil); !errors.Is(err, ErrEmptyInput) {
		t.Errorf("EmbedBatch(nil) error = %v, want %v", err, ErrEmptyInput)
	}
	if _, err := embedder.EmbedBatch(context.Background(), []string{"ok", ""}); !errors.Is(err, ErrEmptyInput) {
		t.Errorf("EmbedBatch(empty element) error = %v, want %v", err, ErrEmptyInput)
	}
	if _, err := embedder.EmbedQuery(context.Background(), ""); !errors.Is(err, ErrEmptyInput) {
		t.Errorf("EmbedQuery(empty) error = %v, want %v", err, ErrEmptyInput)
	}
	if calls != 0 {
		t.Errorf("provider was called %d times for invalid input, want 0", calls)
	}
}

func TestNewEmbedderDefaultsBaseURLAndModel(t *testing.T) {
	embedder := NewEmbedder("token", "", "")
	if embedder.baseURL != "https://integrate.api.nvidia.com/v1" {
		t.Errorf("baseURL = %q, want %q", embedder.baseURL, defaultBaseURL)
	}
	if embedder.model != defaultModel {
		t.Errorf("model = %q, want %q", embedder.model, defaultModel)
	}

	custom := NewEmbedder("token", "https://provider.example/v1/", "custom-model")
	if custom.baseURL != "https://provider.example/v1" {
		t.Errorf("custom baseURL = %q, want https://provider.example/v1", custom.baseURL)
	}
	if custom.model != "custom-model" {
		t.Errorf("custom model = %q, want custom-model", custom.model)
	}

	endpointURL := NewEmbedder("token", "https://provider.example/v1/embeddings", "custom-model")
	if endpointURL.baseURL != "https://provider.example/v1" {
		t.Errorf("baseURL with endpoint suffix = %q, want https://provider.example/v1", endpointURL.baseURL)
	}
}

func TestNewEmbedderWithHTTPClientUsesClientAndSingleEndpointSuffix(t *testing.T) {
	const baseURL = "https://provider.example/v1"
	var gotURL string
	httpClient := &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		gotURL = req.URL.String()
		return jsonResponse(http.StatusOK, `{"data":[{"index":0,"embedding":[0.1]}]}`), nil
	})}
	embedder := NewEmbedderWithHTTPClient(
		"token",
		baseURL+"/embeddings",
		"custom-model",
		httpClient,
	)

	if _, err := embedder.Embed(context.Background(), "hello"); err != nil {
		t.Fatalf("Embed() returned error: %v", err)
	}
	if want := baseURL + "/embeddings"; gotURL != want {
		t.Errorf("request URL = %q, want %q", gotURL, want)
	}
}

func TestEmbedRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		model   string
	}{
		{name: "plain HTTP base URL", baseURL: "http://api.example/v1", model: defaultModel},
		{name: "base URL with query", baseURL: "https://api.example/v1?token=leak", model: defaultModel},
		{name: "model with query", baseURL: defaultBaseURL, model: "model?version=1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			embedder := NewEmbedder("token", tt.baseURL, tt.model)
			if _, err := embedder.Embed(context.Background(), "hello"); !errors.Is(err, ErrInvalidConfig) {
				t.Errorf("Embed() error = %v, want %v", err, ErrInvalidConfig)
			}
		})
	}
}

func equalVector(left, right []float32) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
