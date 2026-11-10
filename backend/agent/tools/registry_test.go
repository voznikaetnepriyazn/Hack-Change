package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sashabaranov/go-openai"
)

type registryTestTool struct {
	name        string
	description string
	parameters  json.RawMessage
	execute     func(context.Context, string) (string, error)
}

func (t *registryTestTool) Name() string                { return t.name }
func (t *registryTestTool) Description() string         { return t.description }
func (t *registryTestTool) Parameters() json.RawMessage { return t.parameters }
func (t *registryTestTool) Execute(ctx context.Context, args string) (string, error) {
	return t.execute(ctx, args)
}

func TestRegistrySchemas(t *testing.T) {
	registry := NewRegistry()
	tools := []*registryTestTool{
		{
			name:        "lookup",
			description: "Find a record",
			parameters:  json.RawMessage(`{"type":"object","properties":{"id":{"type":"string"}}}`),
		},
		{
			name:        "update",
			description: "Update a record",
			parameters:  json.RawMessage(`{"type":"object","required":["id"]}`),
		},
	}
	for _, testTool := range tools {
		registry.RegisterReadOnly(testTool)
	}

	schemas := registry.Schemas()
	if len(schemas) != len(tools) {
		t.Fatalf("Schemas() returned %d tools, want %d", len(schemas), len(tools))
	}

	gotByName := make(map[string]openai.Tool, len(schemas))
	for _, schema := range schemas {
		if schema.Function == nil {
			t.Fatal("Schemas() returned tool with nil Function")
		}
		gotByName[schema.Function.Name] = schema
	}

	for _, want := range tools {
		schema, ok := gotByName[want.name]
		if !ok {
			t.Errorf("Schemas() is missing tool %q", want.name)
			continue
		}
		if schema.Type != openai.ToolTypeFunction {
			t.Errorf("schema type for %q = %q, want %q", want.name, schema.Type, openai.ToolTypeFunction)
		}
		if schema.Function.Description != want.description {
			t.Errorf("description for %q = %q, want %q", want.name, schema.Function.Description, want.description)
		}
		if !reflect.DeepEqual(schema.Function.Parameters, want.parameters) {
			t.Errorf("parameters for %q = %s, want %s", want.name, schema.Function.Parameters, want.parameters)
		}
	}
}

func TestRegistryExecute(t *testing.T) {
	type contextKey struct{}
	wantContextValue := "request-context"
	var gotArgs string

	registry := NewRegistry()
	registry.RegisterMutating(&registryTestTool{
		name: "echo",
		execute: func(ctx context.Context, args string) (string, error) {
			gotArgs = args
			if got := ctx.Value(contextKey{}); got != wantContextValue {
				t.Errorf("tool context value = %v, want %q", got, wantContextValue)
			}
			return `{"ok":true}`, nil
		},
	})

	ctx := context.WithValue(context.Background(), contextKey{}, wantContextValue)
	const args = `{"message":"hello"}`
	if got := registry.Execute(ctx, "echo", args); got != `{"ok":true}` {
		t.Errorf("Execute() = %q, want %q", got, `{"ok":true}`)
	}
	if gotArgs != args {
		t.Errorf("tool received args %q, want %q", gotArgs, args)
	}
}

func TestRegistryExecuteErrors(t *testing.T) {
	registry := NewRegistry()
	registry.RegisterMutating(&registryTestTool{
		name: "fail",
		execute: func(context.Context, string) (string, error) {
			return "", errors.New("execution failed")
		},
	})

	if got, want := registry.Execute(context.Background(), "missing", `{}`), `Error: unknown tool "missing"`; got != want {
		t.Errorf("Execute() for unknown tool = %q, want %q", got, want)
	}
	if got, want := registry.Execute(context.Background(), "fail", `{}`), "Error: execution failed"; got != want {
		t.Errorf("Execute() for failed tool = %q, want %q", got, want)
	}
}

func TestRegistryExecuteBatchRunsReadOnlyToolsInParallelAndPreservesOrder(t *testing.T) {
	const callCount = 3
	release := make(chan struct{})
	started := make(chan string, callCount)
	var active int32
	var maxActive int32

	registry := NewRegistry()
	for _, name := range []string{"first", "second", "third"} {
		name := name
		registry.RegisterReadOnly(&registryTestTool{
			name: name,
			execute: func(context.Context, string) (string, error) {
				current := atomic.AddInt32(&active, 1)
				for {
					previous := atomic.LoadInt32(&maxActive)
					if current <= previous || atomic.CompareAndSwapInt32(&maxActive, previous, current) {
						break
					}
				}
				started <- name
				<-release
				atomic.AddInt32(&active, -1)
				return "result-" + name, nil
			},
		})
	}

	calls := []openai.ToolCall{
		{ID: "call-1", Function: openai.FunctionCall{Name: "first", Arguments: `{"n":1}`}},
		{ID: "call-2", Function: openai.FunctionCall{Name: "second", Arguments: `{"n":2}`}},
		{ID: "call-3", Function: openai.FunctionCall{Name: "third", Arguments: `{"n":3}`}},
	}

	done := make(chan []Message, 1)
	go func() {
		done <- registry.ExecuteBatch(context.Background(), calls, callCount)
	}()

	observed := make(map[string]bool, callCount)
	for range callCount {
		select {
		case name := <-started:
			observed[name] = true
		case <-time.After(time.Second):
			close(release)
			<-done
			t.Fatalf("only %d read-only tools started concurrently; want %d", len(observed), callCount)
		}
	}
	close(release)

	var results []Message
	select {
	case results = <-done:
	case <-time.After(time.Second):
		t.Fatal("ExecuteBatch() did not finish after releasing read-only tools")
	}

	if got := atomic.LoadInt32(&maxActive); got < 2 {
		t.Errorf("maximum concurrent executions = %d, want at least 2", got)
	}
	if len(results) != len(calls) {
		t.Fatalf("ExecuteBatch() returned %d results, want %d", len(results), len(calls))
	}
	for i, call := range calls {
		if results[i].Role != "tool" {
			t.Errorf("result %d role = %q, want %q", i, results[i].Role, "tool")
		}
		if results[i].ToolCallID != call.ID {
			t.Errorf("result %d ToolCallID = %q, want %q", i, results[i].ToolCallID, call.ID)
		}
		if want := "result-" + call.Function.Name; results[i].Content != want {
			t.Errorf("result %d content = %q, want %q", i, results[i].Content, want)
		}
	}
}

func TestRegistryExecuteBatchHandlesUnknownToolAndDefaultsLimit(t *testing.T) {
	registry := NewRegistry()
	registry.RegisterReadOnly(&registryTestTool{
		name: "known",
		execute: func(context.Context, string) (string, error) {
			return "ok", nil
		},
	})

	results := registry.ExecuteBatch(context.Background(), []openai.ToolCall{
		{ID: "known-call", Function: openai.FunctionCall{Name: "known", Arguments: `{}`}},
		{ID: "unknown-call", Function: openai.FunctionCall{Name: "missing", Arguments: `{}`}},
	}, 0)

	if len(results) != 2 {
		t.Fatalf("ExecuteBatch() returned %d results, want 2", len(results))
	}
	if results[0].ToolCallID != "known-call" || results[0].Content != "ok" {
		t.Errorf("known tool result = %+v, want call ID known-call and content ok", results[0])
	}
	if results[1].ToolCallID != "unknown-call" || !strings.Contains(results[1].Content, `unknown tool "missing"`) {
		t.Errorf("unknown tool result = %+v, want unknown tool error", results[1])
	}
}

func TestRegistryExecuteBatchRunsMutatingToolsSequentially(t *testing.T) {
	var active int32
	var maxActive int32
	registry := NewRegistry()
	for _, name := range []string{"first", "second", "third"} {
		name := name
		registry.RegisterMutating(&registryTestTool{
			name: name,
			execute: func(context.Context, string) (string, error) {
				current := atomic.AddInt32(&active, 1)
				if current > maxActive {
					maxActive = current
				}
				time.Sleep(10 * time.Millisecond)
				atomic.AddInt32(&active, -1)
				return name, nil
			},
		})
	}

	calls := []openai.ToolCall{
		{ID: "1", Function: openai.FunctionCall{Name: "first"}},
		{ID: "2", Function: openai.FunctionCall{Name: "second"}},
		{ID: "3", Function: openai.FunctionCall{Name: "third"}},
	}
	results := registry.ExecuteBatch(context.Background(), calls, 3)

	if maxActive != 1 {
		t.Errorf("maximum concurrent mutating executions = %d, want 1", maxActive)
	}
	if len(results) != len(calls) {
		t.Fatalf("ExecuteBatch() returned %d results, want %d", len(results), len(calls))
	}
	for i, call := range calls {
		if results[i].ToolCallID != call.ID || results[i].Content != call.Function.Name {
			t.Errorf("result %d = %+v, want call ID %q and content %q", i, results[i], call.ID, call.Function.Name)
		}
	}
}

func TestRegistryConcurrentRegisterAndRead(t *testing.T) {
	registry := NewRegistry()
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				name := fmt.Sprintf("tool-%d-%d", i, j)
				registry.RegisterReadOnly(&registryTestTool{
					name: name,
					execute: func(context.Context, string) (string, error) {
						return "ok", nil
					},
				})
			}
		}()
	}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				registry.Has("tool-0-0")
				registry.Rights("tool-0-0")
				registry.Schemas()
				registry.Execute(context.Background(), "tool-0-0", `{}`)
			}
		}()
	}
	wg.Wait()

	if got := len(registry.Schemas()); got != 500 {
		t.Errorf("Schemas() returned %d tools after registration, want 500", got)
	}
}

func TestRegistryExecuteBatchDoesNotOverlapMutatingAndReadOnlyTools(t *testing.T) {
	var activeReadOnly int32
	var mutatingRanDuringRead int32
	releaseReadOnly := make(chan struct{})
	startedReadOnly := make(chan struct{})

	registry := NewRegistry()
	registry.RegisterReadOnly(&registryTestTool{
		name: "read",
		execute: func(context.Context, string) (string, error) {
			atomic.AddInt32(&activeReadOnly, 1)
			close(startedReadOnly)
			<-releaseReadOnly
			atomic.AddInt32(&activeReadOnly, -1)
			return "read result", nil
		},
	})
	registry.RegisterMutating(&registryTestTool{
		name: "write",
		execute: func(context.Context, string) (string, error) {
			if atomic.LoadInt32(&activeReadOnly) != 0 {
				atomic.StoreInt32(&mutatingRanDuringRead, 1)
			}
			return "write result", nil
		},
	})

	calls := []openai.ToolCall{
		{ID: "read-call", Function: openai.FunctionCall{Name: "read"}},
		{ID: "write-call", Function: openai.FunctionCall{Name: "write"}},
	}
	done := make(chan []Message, 1)
	go func() {
		done <- registry.ExecuteBatch(context.Background(), calls, 2)
	}()

	select {
	case <-startedReadOnly:
	case <-time.After(time.Second):
		close(releaseReadOnly)
		t.Fatal("read-only tool did not start")
	}

	select {
	case results := <-done:
		close(releaseReadOnly)
		t.Fatalf("batch completed while read-only call was blocked: %+v", results)
	case <-time.After(25 * time.Millisecond):
	}
	close(releaseReadOnly)

	select {
	case results := <-done:
		if len(results) != 2 || results[0].Content != "read result" || results[1].Content != "write result" {
			t.Errorf("ExecuteBatch() results = %+v, want read then write results", results)
		}
	case <-time.After(time.Second):
		t.Fatal("ExecuteBatch() did not finish after releasing read-only tool")
	}
	if atomic.LoadInt32(&mutatingRanDuringRead) != 0 {
		t.Error("mutating tool overlapped with a read-only tool")
	}
}
