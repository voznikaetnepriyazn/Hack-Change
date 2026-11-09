package run

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/agent/session"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/config"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/event"
	detection "github.com/voznikaetnepriyazn/Hack-Change/agent/internal/lib/loop"
	tool "github.com/voznikaetnepriyazn/Hack-Change/agent/tools"
)

type runTestTool struct {
	name    string
	execute func(context.Context, string) (string, error)
}

func (t runTestTool) Name() string                { return t.name }
func (t runTestTool) Description() string         { return "test tool" }
func (t runTestTool) Parameters() json.RawMessage { return json.RawMessage(`{"type":"object"}`) }
func (t runTestTool) Execute(ctx context.Context, args string) (string, error) {
	return t.execute(ctx, args)
}

func newRunTestConfig(t *testing.T, handler http.HandlerFunc) config.AgentConfig {
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
			MaxIterations:   4,
			LLMTimeout:      time.Second,
			ToolConcurrency: 2,
		},
		Session: session.NewSession("system prompt"),
	}
}

func writeCompletion(w http.ResponseWriter, message openai.ChatCompletionMessage, finishReason string) {
	response := openai.ChatCompletionResponse{
		ID:     "test-completion",
		Object: "chat.completion",
		Model:  "gpt-4o-mini",
		Choices: []openai.ChatCompletionChoice{
			{
				Index:        0,
				Message:      message,
				FinishReason: openai.FinishReason(finishReason),
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func TestRunReturnsAssistantResponse(t *testing.T) {
	cfg := newRunTestConfig(t, func(w http.ResponseWriter, _ *http.Request) {
		writeCompletion(w, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: "Hello!",
		}, "stop")
	})
	var events []event.Event
	cfg.On = func(e event.Event) { events = append(events, e) }

	result := Run(context.Background(), cfg, "Hi")
	if result.Err != nil {
		t.Fatalf("Run() returned unexpected error: %v", result.Err)
	}
	if result.Output != "Hello!" || result.Iterations != 1 || len(result.ToolsUsed) != 0 {
		t.Errorf("Run() = %+v, want output Hello!, one iteration, no tools", result)
	}

	messages := cfg.Session.Messages()
	if len(messages) != 3 ||
		messages[0].Role != openai.ChatMessageRoleSystem ||
		messages[1].Role != openai.ChatMessageRoleUser ||
		messages[1].Content != "Hi" ||
		messages[2].Role != openai.ChatMessageRoleAssistant ||
		messages[2].Content != "Hello!" {
		t.Errorf("session messages = %#v, want system, user, assistant history", messages)
	}
	if len(events) != 2 {
		t.Fatalf("received %d events, want request and response events", len(events))
	}
	if got, ok := events[0].(event.LLMRequestEvent); !ok || got.Step != 1 {
		t.Errorf("first event = %#v, want LLMRequestEvent for step 1", events[0])
	}
	if got, ok := events[1].(event.LLMResponseEvent); !ok || got.Step != 1 || got.ToolCalls != 0 {
		t.Errorf("second event = %#v, want LLMResponseEvent for step 1 without tools", events[1])
	}
}

func TestRunExecutesToolAndContinuesConversation(t *testing.T) {
	var requestCount int32
	var secondRequestMessages []openai.ChatCompletionMessage
	cfg := newRunTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		requestNumber := atomic.AddInt32(&requestCount, 1)
		var request openai.ChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}

		switch requestNumber {
		case 1:
			writeCompletion(w, openai.ChatCompletionMessage{
				Role: openai.ChatMessageRoleAssistant,
				ToolCalls: []openai.ToolCall{
					{
						ID:   "call-1",
						Type: openai.ToolTypeFunction,
						Function: openai.FunctionCall{
							Name:      "lookup",
							Arguments: `{"id":7}`,
						},
					},
				},
			}, "tool_calls")
		case 2:
			secondRequestMessages = request.Messages
			writeCompletion(w, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: "Found it.",
			}, "stop")
		default:
			http.Error(w, "unexpected extra request", http.StatusInternalServerError)
		}
	})

	var gotArgs string
	cfg.Registry.RegisterReadOnly(runTestTool{
		name: "lookup",
		execute: func(_ context.Context, args string) (string, error) {
			gotArgs = args
			return `{"name":"record"}`, nil
		},
	})
	var events []event.Event
	cfg.On = func(e event.Event) { events = append(events, e) }

	result := Run(context.Background(), cfg, "Find record 7")
	if result.Err != nil {
		t.Fatalf("Run() returned unexpected error: %v", result.Err)
	}
	if result.Output != "Found it." || result.Iterations != 2 {
		t.Errorf("Run() output/iterations = %q/%d, want Found it./2", result.Output, result.Iterations)
	}
	if len(result.ToolsUsed) != 1 || result.ToolsUsed[0] != "lookup" {
		t.Errorf("ToolsUsed = %v, want [lookup]", result.ToolsUsed)
	}
	if gotArgs != `{"id":7}` {
		t.Errorf("tool received args %q, want %q", gotArgs, `{"id":7}`)
	}
	if atomic.LoadInt32(&requestCount) != 2 {
		t.Errorf("LLM request count = %d, want 2", requestCount)
	}

	var foundToolResult bool
	for _, message := range secondRequestMessages {
		if message.Role == openai.ChatMessageRoleTool &&
			message.ToolCallID == "call-1" &&
			message.Content == `{"name":"record"}` {
			foundToolResult = true
		}
	}
	if !foundToolResult {
		t.Errorf("second request messages do not contain the tool result: %#v", secondRequestMessages)
	}

	var startEvent, endEvent bool
	for _, e := range events {
		switch got := e.(type) {
		case event.ToolStartEvent:
			startEvent = got.CallID == "call-1" && got.Name == "lookup"
		case event.ToolEndEvent:
			endEvent = got.CallID == "call-1" && got.Name == "lookup" && got.Result == `{"name":"record"}`
		}
	}
	if !startEvent || !endEvent {
		t.Errorf("tool events start/end = %t/%t, want both matching tool events", startEvent, endEvent)
	}
}

func TestRunRecordsInvalidToolCallAndContinues(t *testing.T) {
	var requestCount int32
	cfg := newRunTestConfig(t, func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&requestCount, 1) == 1 {
			writeCompletion(w, openai.ChatCompletionMessage{
				Role: openai.ChatMessageRoleAssistant,
				ToolCalls: []openai.ToolCall{
					{
						ID:   "call-invalid",
						Type: openai.ToolTypeFunction,
						Function: openai.FunctionCall{
							Name:      "not_registered",
							Arguments: `{}`,
						},
					},
				},
			}, "tool_calls")
			return
		}
		writeCompletion(w, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: "Recovered.",
		}, "stop")
	})

	result := Run(context.Background(), cfg, "Do something")
	if result.Err != nil {
		t.Fatalf("Run() returned unexpected error: %v", result.Err)
	}
	if result.Output != "Recovered." || result.Iterations != 2 || len(result.ToolsUsed) != 0 {
		t.Errorf("Run() = %+v, want recovered answer after two iterations without executed tools", result)
	}

	messages := cfg.Session.Messages()
	if len(messages) < 4 {
		t.Fatalf("session has %d messages, want invalid tool call and its error result recorded", len(messages))
	}
	lastToolResult := messages[len(messages)-2]
	if lastToolResult.Role != openai.ChatMessageRoleTool ||
		lastToolResult.ToolCallID != "call-invalid" ||
		!strings.Contains(lastToolResult.Content, `unknown tool "not_registered"`) {
		t.Errorf("recorded tool error = %#v, want unknown tool error for call-invalid", lastToolResult)
	}
}

func TestRunReturnsMaxIterationsError(t *testing.T) {
	var requestCount int32
	cfg := newRunTestConfig(t, func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		writeCompletion(w, openai.ChatCompletionMessage{
			Role: openai.ChatMessageRoleAssistant,
			ToolCalls: []openai.ToolCall{
				{
					ID:   fmt.Sprintf("call-%d", requestCount),
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      "lookup",
						Arguments: `{}`,
					},
				},
			},
		}, "tool_calls")
	})
	cfg.Limits.MaxIterations = 2
	cfg.Registry.RegisterReadOnly(runTestTool{
		name: "lookup",
		execute: func(context.Context, string) (string, error) {
			return "ok", nil
		},
	})

	result := Run(context.Background(), cfg, "keep going")
	if result.Err == nil || result.Err.Error() != "max iterations exceeded (2)" {
		t.Fatalf("Run() error = %v, want max iterations exceeded (2)", result.Err)
	}
	if result.Iterations != 2 || len(result.ToolsUsed) != 2 {
		t.Errorf("Run() iterations/tools = %d/%v, want 2 iterations and two tool calls", result.Iterations, result.ToolsUsed)
	}
}

func TestRunHandlesLLMErrors(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		wantErr    string
		timeout    time.Duration
	}{
		{
			name: "API error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				http.Error(w, `{"error":{"message":"unavailable"}}`, http.StatusServiceUnavailable)
			},
			wantErr: "LLM error on step 1:",
		},
		{
			name: "empty response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{"choices":[]}`)
			},
			wantErr: "empty response from LLM on step 1",
		},
		{
			name: "LLM timeout",
			handler: func(w http.ResponseWriter, r *http.Request) {
				select {
				case <-r.Context().Done():
				case <-time.After(100 * time.Millisecond):
				}
			},
			wantErr: "LLM timeout on step 1",
			timeout: 20 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := newRunTestConfig(t, tt.handler)
			if tt.timeout > 0 {
				cfg.Limits.LLMTimeout = tt.timeout
			}
			result := Run(context.Background(), cfg, "request")
			if result.Err == nil {
				t.Fatal("Run() returned nil error")
			}
			if !strings.HasPrefix(result.Err.Error(), tt.wantErr) {
				t.Errorf("Run() error = %q, want prefix %q", result.Err, tt.wantErr)
			}
		})
	}
}

func TestRunReturnsErrorForExpiredContext(t *testing.T) {
	cfg := newRunTestConfig(t, func(http.ResponseWriter, *http.Request) {
		t.Error("LLM request should not be made for an expired context")
	})
	cfg.Limits.Deadline = time.Second
	if cfg.Loop == nil {
		cfg.Loop = detection.NewLoopDetector(4, 2)
	}
	if err := cfg.Loop.Record("old", "input", "output"); err != nil {
		t.Fatalf("setup loop detector: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := Run(ctx, cfg, "request")

	if result.Err == nil || !errors.Is(result.Err, context.Canceled) {
		t.Fatalf("Run() error = %v, want wrapped context.Canceled", result.Err)
	}
	if result.Iterations != 0 {
		t.Errorf("Iterations = %d, want 0", result.Iterations)
	}
	if len(cfg.Session.Messages()) != 2 {
		t.Errorf("session history length = %d, want system prompt and user input", len(cfg.Session.Messages()))
	}
	if err := cfg.Loop.Record("old", "input", "output"); err != nil {
		t.Errorf("loop detector was not reset before run: %v", err)
	}
}
