package tool

import (
	"context"
	"fmt"
	"sync"

	"github.com/sashabaranov/go-openai"
)

type Registry struct {
	tools map[string]Entry
}

// Constructor
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Entry),
	}
}

// Registry is a registry of tools
func (r *Registry) Register(t Tool, rights Rights) {
	r.tools[t.Name()] = Entry{Tool: t, Rights: rights}
}

// Rights returns the rights for a tool
func (r *Registry) Rights(name string) (Rights, bool) {
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
	_, ok := r.tools[name]
	return ok
}

// Schemas get the schemas for all registered tools
func (r *Registry) Schemas() []openai.Tool {
	result := make([]openai.Tool, 0, len(r.tools))
	for _, t := range r.tools {
		result = append(result, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        t.Tool.Name(),
				Description: t.Tool.Description(),
				Parameters:  t.Tool.Parameters(),
			},
		})
	}
	return result
}

// Execute execute a tool
func (r *Registry) Execute(ctx context.Context, name, argsJSON string) string {
	// check if tool is registered
	currentTool, ok := r.tools[name]
	if !ok {
		return fmt.Sprintf("Error: unknown tool %q", name)
	}

	result, err := currentTool.Tool.Execute(ctx, argsJSON)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return result
}

// ExecuteBatch executes tool_calls with rights.
func (r *Registry) ExecuteBatch(ctx context.Context, calls []openai.ToolCall, limit int) []Message {
	if limit <= 0 {
		limit = 5
	}

	results := make([]Message, len(calls))
	var wg sync.WaitGroup
	sem := make(chan struct{}, limit)

	for i, call := range calls {
		rights, ok := r.Rights(call.Function.Name)
		if !ok {
			results[i] = Message{
				Role:       "tool",
				ToolCallID: call.ID,
				Content:    fmt.Sprintf("Error: unknown tool %q", call.Function.Name),
			}
			continue
		}

		// confirm
		if !rights.ReadOnly || !rights.Parallel {
			results[i] = Message{
				Role:       "tool",
				ToolCallID: call.ID,
				Content:    r.Execute(ctx, call.Function.Name, call.Function.Arguments),
			}
			continue
		}

		// read-only
		wg.Add(1)
		go func(idx int, c openai.ToolCall) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			results[idx] = Message{
				Role:       "tool",
				ToolCallID: c.ID,
				Content:    r.Execute(ctx, c.Function.Name, c.Function.Arguments),
			}
		}(i, call)
	}

	wg.Wait()
	return results
}
