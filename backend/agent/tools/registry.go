package tool

import (
	"context"
	"fmt"
	"sync"

	"github.com/sashabaranov/go-openai"
)

type Registry struct {
	mu    sync.RWMutex
	tools map[string]Entry
}

// DefaultToolConcurrency is the maximum number of parallel read-only tool calls.
const DefaultToolConcurrency = 3

// Constructor
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Entry),
	}
}

// Registry is a registry of tools
func (r *Registry) Register(t Tool, rights Rights) {
	name := t.Name()
	entry := Entry{Tool: t, Rights: rights}
	r.mu.Lock()
	if r.tools == nil {
		r.tools = make(map[string]Entry)
	}
	r.tools[name] = entry
	r.mu.Unlock()
}

// Rights returns the rights for a tool
func (r *Registry) Rights(name string) (Rights, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entry, ok := r.tools[name]
	if !ok {
		return Rights{}, false
	}
	return entry.Rights, true
}

// RegisterReadOnly register read-only tools and they execute parellel.
func (r *Registry) RegisterReadOnly(t Tool) {
	r.Register(t, Rights{ReadOnly: true, Parallel: true})
}

// RegisterMutating register mutating tools and they execute sequentially.
func (r *Registry) RegisterMutating(t Tool) {
	r.Register(t, Rights{ReadOnly: false, Parallel: false})
}

func (r *Registry) RegisterAll() {

}

// Has returns true if the tool is registered
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.tools[name]
	return ok
}

// Schemas get the schemas for all registered tools
func (r *Registry) Schemas() []openai.Tool {
	r.mu.RLock()
	entries := make([]Entry, 0, len(r.tools))
	for _, entry := range r.tools {
		entries = append(entries, entry)
	}
	r.mu.RUnlock()

	result := make([]openai.Tool, 0, len(entries))
	for _, entry := range entries {
		result = append(result, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        entry.Tool.Name(),
				Description: entry.Tool.Description(),
				Parameters:  entry.Tool.Parameters(),
			},
		})
	}
	return result
}

// Execute execute a tool
func (r *Registry) Execute(ctx context.Context, name, argsJSON string) string {
	entry, ok := r.entry(name)
	if !ok {
		return fmt.Sprintf("Error: unknown tool %q", name)
	}

	result, err := entry.Tool.Execute(ctx, argsJSON)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return result
}

// ExecuteBatch executes tool_calls with rights.
func (r *Registry) ExecuteBatch(ctx context.Context, calls []openai.ToolCall, limit int) []Message {
	if limit <= 0 {
		limit = DefaultToolConcurrency
	}

	results := make([]Message, len(calls))
	entries := make([]Entry, len(calls))
	registered := make([]bool, len(calls))
	r.mu.RLock()
	for i, call := range calls {
		entries[i], registered[i] = r.tools[call.Function.Name]
	}
	r.mu.RUnlock()

	for i := 0; i < len(calls); {
		call := calls[i]
		entry := entries[i]
		if !registered[i] {
			results[i] = Message{
				Role:       "tool",
				ToolCallID: call.ID,
				Content:    fmt.Sprintf("Error: unknown tool %q", call.Function.Name),
			}
			i++
			continue
		}

		if !entry.Rights.ReadOnly || !entry.Rights.Parallel {
			results[i] = executeEntry(ctx, call, entry)
			i++
			continue
		}

		end := i
		for end < len(calls) && registered[end] &&
			entries[end].Rights.ReadOnly && entries[end].Rights.Parallel {
			end++
		}

		var wg sync.WaitGroup
		sem := make(chan struct{}, limit)
		for idx := i; idx < end; idx++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				results[idx] = executeEntry(ctx, calls[idx], entries[idx])
			}(idx)
		}
		wg.Wait()
		i = end
	}

	return results
}

func (r *Registry) entry(name string) (Entry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entry, ok := r.tools[name]
	return entry, ok
}

func executeEntry(ctx context.Context, call openai.ToolCall, entry Entry) Message {
	result, err := entry.Tool.Execute(ctx, call.Function.Arguments)
	if err != nil {
		result = fmt.Sprintf("Error: %v", err)
	}
	return Message{
		Role:       "tool",
		ToolCallID: call.ID,
		Content:    result,
	}
}
