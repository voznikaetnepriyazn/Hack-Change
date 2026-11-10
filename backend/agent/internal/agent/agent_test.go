package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/config"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/event"
	tool "github.com/voznikaetnepriyazn/Hack-Change/agent/tools"
)

func newTestAgent(t *testing.T, handler http.HandlerFunc) *Agent {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	clientConfig := openai.DefaultConfig("test-token")
	clientConfig.BaseURL = server.URL + "/v1"
	return NewAgent(
		openai.NewClientWithConfig(clientConfig),
		tool.NewRegistry(),
		"gpt-4o-mini",
		false,
	)
}

func writeAgentCompletion(w http.ResponseWriter, content string) {
	response := openai.ChatCompletionResponse{
		ID:     "agent-test",
		Object: "chat.completion",
		Model:  "gpt-4o-mini",
		Choices: []openai.ChatCompletionChoice{{
			Index: 0,
			Message: openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: content,
			},
			FinishReason: openai.FinishReasonStop,
		}},
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func TestNewAgentAndFluentOptions(t *testing.T) {
	agent := NewAgent(nil, nil, "test-model", true)
	if agent.baseCfg.Model != "test-model" || !agent.baseCfg.Stream {
		t.Errorf("base config model/stream = %q/%t, want test-model/true", agent.baseCfg.Model, agent.baseCfg.Stream)
	}
	if agent.baseCfg.Limits != config.DefaultLimits() {
		t.Errorf("base config limits = %+v, want defaults %+v", agent.baseCfg.Limits, config.DefaultLimits())
	}
	if agent.baseCfg.Client != nil || agent.baseCfg.Registry != nil {
		t.Error("constructor did not preserve supplied nil dependencies")
	}
	if agent.baseCfg.On == nil {
		t.Error("constructor must initialize the default event handler")
	}
	if agent.sessions == nil || len(agent.sessions) != 0 {
		t.Errorf("initial sessions = %#v, want empty initialized map", agent.sessions)
	}

	customLimits := config.Limits{MaxIterations: 2, LLMTimeout: time.Second}
	handler := func(event.Event) {}
	got := agent.WithLimits(customLimits).WithStream(false).WithEventHandler(handler)
	if got != agent {
		t.Fatal("fluent configuration methods should return the same agent")
	}
	if agent.baseCfg.Limits != customLimits {
		t.Errorf("configured limits = %+v, want %+v", agent.baseCfg.Limits, customLimits)
	}
	if agent.baseCfg.Stream {
		t.Error("WithStream(false) did not disable streaming")
	}
	agent.WithEventHandler(nil)
	if agent.baseCfg.On != nil {
		t.Error("WithEventHandler(nil) did not clear the event handler")
	}
}

func TestConfigForSessionCreatesIndependentLoopDetectors(t *testing.T) {
	agent := NewAgent(nil, nil, "test-model", false)

	first := agent.configForSession("session-a")
	second := agent.configForSession("session-b")
	nextRun := agent.configForSession("session-a")
	if first.Loop == nil || second.Loop == nil {
		t.Fatal("run configs must have loop detectors")
	}
	if first.Loop == second.Loop {
		t.Error("different runs share a loop detector")
	}
	if first.Loop == nextRun.Loop {
		t.Error("separate runs for the same session share a loop detector")
	}
	if first.Session == second.Session {
		t.Error("different session IDs share conversation history")
	}
	if first.Session != nextRun.Session {
		t.Error("repeated runs for one session do not reuse conversation history")
	}
}

func TestRunMaintainsHistoryPerSession(t *testing.T) {
	var mu sync.Mutex
	requests := make(map[string][][]openai.ChatCompletionMessage)
	agent := newTestAgent(t, func(w http.ResponseWriter, r *http.Request) {
		var request openai.ChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		var latestUser string
		for i := len(request.Messages) - 1; i >= 0; i-- {
			if request.Messages[i].Role == openai.ChatMessageRoleUser {
				latestUser = request.Messages[i].Content
				break
			}
		}
		mu.Lock()
		requests[latestUser] = append(requests[latestUser], request.Messages)
		mu.Unlock()
		writeAgentCompletion(w, "answer to "+latestUser)
	})

	output, err := agent.Run(context.Background(), "session-a", "first")
	if err != nil || output != "answer to first" {
		t.Fatalf("Run() = %q, %v; want answer to first, nil", output, err)
	}
	result := agent.RunWithResult(context.Background(), "session-a", "second")
	if result.Err != nil || result.Output != "answer to second" || result.Iterations != 1 {
		t.Fatalf("RunWithResult() = %+v, want second answer in one iteration", result)
	}
	output, err = agent.Run(context.Background(), "session-b", "other")
	if err != nil || output != "answer to other" {
		t.Fatalf("Run() for second session = %q, %v; want answer to other, nil", output, err)
	}

	mu.Lock()
	defer mu.Unlock()
	secondRequest := requests["second"]
	if len(secondRequest) != 1 {
		t.Fatalf("received %d requests for second message, want 1", len(secondRequest))
	}
	if len(secondRequest[0]) != 4 {
		t.Fatalf("second request history length = %d, want 4 messages", len(secondRequest[0]))
	}
	if secondRequest[0][1].Content != "first" ||
		secondRequest[0][2].Role != openai.ChatMessageRoleAssistant ||
		secondRequest[0][2].Content != "answer to first" ||
		secondRequest[0][3].Content != "second" {
		t.Errorf("second request history does not preserve prior turn: %#v", secondRequest[0])
	}

	otherRequest := requests["other"]
	if len(otherRequest) != 1 {
		t.Fatalf("received %d requests for other session, want 1", len(otherRequest))
	}
	if len(otherRequest[0]) != 2 || otherRequest[0][1].Content != "other" {
		t.Errorf("new session inherited another session's history: %#v", otherRequest[0])
	}
}

func TestRunWithChannelForwardsStreamDeltasAndEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"hello \"},\"finish_reason\":null}]}\n\n")
		fmt.Fprint(w, "data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"world\"},\"finish_reason\":\"stop\"}]}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	t.Cleanup(server.Close)

	clientConfig := openai.DefaultConfig("test-token")
	clientConfig.BaseURL = server.URL + "/v1"
	agent := NewAgent(openai.NewClientWithConfig(clientConfig), tool.NewRegistry(), "gpt-4o-mini", true)
	agent.WithLimits(config.Limits{
		MaxIterations: 1,
		LLMTimeout:    time.Second,
	})

	var eventMu sync.Mutex
	var receivedEvents []event.Event
	agent.WithEventHandler(func(e event.Event) {
		eventMu.Lock()
		defer eventMu.Unlock()
		receivedEvents = append(receivedEvents, e)
	})

	deltas := make(chan string, 2)
	output, err := agent.RunWithChannel(context.Background(), "channel-session", "say hello", deltas)
	if err != nil {
		t.Fatalf("RunWithChannel() returned unexpected error: %v", err)
	}
	if output != "hello world" {
		t.Errorf("RunWithChannel() output = %q, want %q", output, "hello world")
	}
	close(deltas)
	var gotDeltas []string
	for delta := range deltas {
		gotDeltas = append(gotDeltas, delta)
	}
	if len(gotDeltas) != 2 || gotDeltas[0] != "hello " || gotDeltas[1] != "world" {
		t.Errorf("stream deltas = %v, want [hello  world]", gotDeltas)
	}

	eventMu.Lock()
	defer eventMu.Unlock()
	var gotDeltaEvents int
	for _, e := range receivedEvents {
		if _, ok := e.(event.StreamDeltaEvent); ok {
			gotDeltaEvents++
		}
	}
	if gotDeltaEvents != 2 {
		t.Errorf("underlying handler received %d stream delta events, want 2", gotDeltaEvents)
	}
}
