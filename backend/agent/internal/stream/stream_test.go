package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/agent/session"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/config"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/event"
	tool "github.com/voznikaetnepriyazn/Hack-Change/agent/tools"
)

func testAgentConfig(t *testing.T, handler http.HandlerFunc, onEvent func(event.Event)) config.AgentConfig {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	clientConfig := openai.DefaultConfig("test-token")
	clientConfig.BaseURL = server.URL + "/v1"

	return config.AgentConfig{
		Client:   openai.NewClientWithConfig(clientConfig),
		Registry: tool.NewRegistry(),
		Model:    "gpt-4o-mini",
		Limits: config.Limits{
			LLMTimeout: time.Second,
		},
		Session: session.NewSession("system prompt"),
		On:      onEvent,
	}
}

func writeSSE(w http.ResponseWriter, chunks ...string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.WriteHeader(http.StatusOK)
	for _, chunk := range chunks {
		fmt.Fprintf(w, "data: %s\n\n", chunk)
	}
	fmt.Fprint(w, "data: [DONE]\n\n")
}

func TestStreamLLMCombinesContentAndEmitsEvents(t *testing.T) {
	var request openai.ChatCompletionRequest
	cfg := testAgentConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/chat/completions" {
			t.Errorf("request = %s %s, want POST /v1/chat/completions", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		writeSSE(w,
			`{"choices":[{"index":0,"delta":{"role":"assistant","content":"Hello "},"finish_reason":null}]}`,
			`{"choices":[{"index":0,"delta":{"content":"world"},"finish_reason":"stop"}]}`,
			`{"choices":[]}`,
		)
	}, nil)

	var events []event.Event
	cfg.On = func(e event.Event) { events = append(events, e) }

	got, err := StreamLLM(context.Background(), cfg, 7)
	if err != nil {
		t.Fatalf("StreamLLM() returned unexpected error: %v", err)
	}
	if got.Role != openai.ChatMessageRoleAssistant || got.Content != "Hello world" {
		t.Errorf("StreamLLM() = role %q content %q, want assistant and %q", got.Role, got.Content, "Hello world")
	}
	if request.Model != cfg.Model || !request.Stream {
		t.Errorf("request model/stream = %q/%t, want %q/true", request.Model, request.Stream, cfg.Model)
	}
	if len(request.Messages) != 1 || request.Messages[0].Content != "system prompt" {
		t.Errorf("request messages = %#v, want system prompt", request.Messages)
	}

	if len(events) != 3 {
		t.Fatalf("received %d events, want 3", len(events))
	}
	wantDeltaTexts := []string{"Hello ", "world"}
	for i, wantText := range wantDeltaTexts {
		gotEvent, ok := events[i].(event.StreamDeltaEvent)
		if !ok {
			t.Fatalf("event %d has type %T, want StreamDeltaEvent", i, events[i])
		}
		if gotEvent.Step != 7 || gotEvent.Text != wantText {
			t.Errorf("delta event %d = %+v, want step 7 and text %q", i, gotEvent, wantText)
		}
	}
	endEvent, ok := events[2].(event.StreamEndEvent)
	if !ok {
		t.Fatalf("last event has type %T, want StreamEndEvent", events[2])
	}
	if endEvent.Step != 7 || endEvent.Text != "Hello world" {
		t.Errorf("end event = %+v, want step 7 and complete response", endEvent)
	}
}

func TestStreamLLMCombinesAndSortsToolCallDeltas(t *testing.T) {
	cfg := testAgentConfig(t, func(w http.ResponseWriter, _ *http.Request) {
		writeSSE(w,
			`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":1,"id":"call-b","type":"function","function":{"name":"second","arguments":"{\"b\":"}}]},"finish_reason":null}]}`,
			`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"call-a","type":"function","function":{"name":"first","arguments":"{\"a\":"}}]},"finish_reason":null}]}`,
			`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":1,"function":{"arguments":"2}"}}]},"finish_reason":null}]}`,
			`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"1}"}}]},"finish_reason":"tool_calls"}]}`,
		)
	}, nil)

	got, err := StreamLLM(context.Background(), cfg, 1)
	if err != nil {
		t.Fatalf("StreamLLM() returned unexpected error: %v", err)
	}
	if len(got.ToolCalls) != 2 {
		t.Fatalf("StreamLLM() returned %d tool calls, want 2", len(got.ToolCalls))
	}

	want := []openai.ToolCall{
		{
			ID:   "call-a",
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionCall{
				Name:      "first",
				Arguments: `{"a":1}`,
			},
		},
		{
			ID:   "call-b",
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionCall{
				Name:      "second",
				Arguments: `{"b":2}`,
			},
		},
	}
	for i := range want {
		if got.ToolCalls[i].ID != want[i].ID ||
			got.ToolCalls[i].Type != want[i].Type ||
			got.ToolCalls[i].Function != want[i].Function {
			t.Errorf("tool call %d = %+v, want %+v", i, got.ToolCalls[i], want[i])
		}
	}
}

func TestStreamLLMReportsStreamCreationError(t *testing.T) {
	cfg := testAgentConfig(t, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
	}, nil)

	_, err := StreamLLM(context.Background(), cfg, 1)
	if err == nil {
		t.Fatal("StreamLLM() returned nil error for failed stream request")
	}
	if !strings.HasPrefix(err.Error(), "stream error:") {
		t.Errorf("StreamLLM() error = %q, want stream error context", err)
	}
}

func TestStreamLLMReportsReceiveError(t *testing.T) {
	cfg := testAgentConfig(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "data: not-json\n\n")
	}, nil)

	_, err := StreamLLM(context.Background(), cfg, 1)
	if err == nil {
		t.Fatal("StreamLLM() returned nil error for malformed stream data")
	}
	if !strings.HasPrefix(err.Error(), "stream recv error:") {
		t.Errorf("StreamLLM() error = %q, want stream recv error context", err)
	}
}

func TestStreamLLMDetectsRepetitionAndEmitsEvent(t *testing.T) {
	repeatedText := strings.Repeat("x", 256)
	cfg := testAgentConfig(t, func(w http.ResponseWriter, _ *http.Request) {
		chunk, err := json.Marshal(map[string]any{
			"choices": []any{
				map[string]any{
					"index": 0,
					"delta": map[string]string{"content": repeatedText},
				},
			},
		})
		if err != nil {
			t.Errorf("marshal stream chunk: %v", err)
			return
		}
		writeSSE(w, string(chunk))
	}, nil)

	var events []event.Event
	cfg.On = func(e event.Event) { events = append(events, e) }

	_, err := StreamLLM(context.Background(), cfg, 3)
	if err == nil || err.Error() != "LLM stuck in repetition loop" {
		t.Fatalf("StreamLLM() error = %v, want repetition loop error", err)
	}
	if len(events) != 2 {
		t.Fatalf("received %d events, want delta and repetition events", len(events))
	}
	if got, ok := events[0].(event.StreamDeltaEvent); !ok || got.Step != 3 || got.Text != repeatedText {
		t.Errorf("first event = %#v, want repetition text delta at step 3", events[0])
	}
	if got, ok := events[1].(event.RepetitionDetectedEvent); !ok || got.Step != 3 || got.Text != repeatedText {
		t.Errorf("second event = %#v, want repetition detection at step 3", events[1])
	}
}
