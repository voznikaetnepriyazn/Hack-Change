package validate

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sashabaranov/go-openai"
	tool "github.com/voznikaetnepriyazn/Hack-Change/agent/tools"
)

type validationTestTool struct{}

func (validationTestTool) Name() string                { return "test_tool" }
func (validationTestTool) Description() string         { return "test tool" }
func (validationTestTool) Parameters() json.RawMessage { return json.RawMessage(`{"type":"object"}`) }
func (validationTestTool) Execute(context.Context, string) (string, error) {
	return "ok", nil
}

func TestValidateToolCall(t *testing.T) {
	registry := tool.NewRegistry()
	registry.RegisterReadOnly(validationTestTool{})

	tests := []struct {
		name    string
		call    openai.ToolCall
		wantErr string
	}{
		{
			name: "valid call",
			call: openai.ToolCall{
				ID:       "call-1",
				Function: openai.FunctionCall{Name: "test_tool", Arguments: `{"query":"hello"}`},
			},
		},
		{
			name: "empty tool name",
			call: openai.ToolCall{
				ID:       "call-1",
				Function: openai.FunctionCall{Arguments: `{}`},
			},
			wantErr: "tool_call has empty name",
		},
		{
			name: "unknown tool",
			call: openai.ToolCall{
				ID:       "call-1",
				Function: openai.FunctionCall{Name: "missing", Arguments: `{}`},
			},
			wantErr: `unknown tool "missing"`,
		},
		{
			name: "empty call id",
			call: openai.ToolCall{
				Function: openai.FunctionCall{Name: "test_tool", Arguments: `{}`},
			},
			wantErr: "tool_call has empty id",
		},
		{
			name: "invalid JSON arguments",
			call: openai.ToolCall{
				ID:       "call-1",
				Function: openai.FunctionCall{Name: "test_tool", Arguments: `{"query":`},
			},
			wantErr: "invalid JSON in arguments for test_tool",
		},
		{
			name: "empty arguments are invalid JSON",
			call: openai.ToolCall{
				ID:       "call-1",
				Function: openai.FunctionCall{Name: "test_tool"},
			},
			wantErr: "invalid JSON in arguments for test_tool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToolCall(tt.call, registry)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("ValidateToolCall() returned unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("ValidateToolCall() returned nil error, want %q", tt.wantErr)
			}
			if err.Error() != tt.wantErr {
				t.Errorf("ValidateToolCall() error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}
