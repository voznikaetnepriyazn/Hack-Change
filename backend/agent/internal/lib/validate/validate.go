package validate

import (
	"encoding/json"
	"fmt"
	"log/slog"

	openai "github.com/sashabaranov/go-openai"
	tool "github.com/voznikaetnepriyazn/Hack-Change/agent/tools"
)

// ValidateToolCall checks if the tool call is valid.
func ValidateToolCall(tc openai.ToolCall, registry *tool.Registry) error {
	if tc.Function.Name == "" {
		slog.Warn("tool_call without name", "id", tc.ID)
		return fmt.Errorf("tool_call has empty name")
	}

	if !registry.Has(tc.Function.Name) {
		slog.Warn("unknown tool",
			"name", tc.Function.Name,
		)
		return fmt.Errorf("unknown tool %q", tc.Function.Name)
	}

	if tc.ID == "" {
		slog.Warn("tool_call without id", "name", tc.Function.Name)
		return fmt.Errorf("tool_call has empty id")
	}

	if !json.Valid([]byte(tc.Function.Arguments)) {
		return fmt.Errorf("invalid JSON in arguments for %s", tc.Function.Name)
	}

	return nil
}
