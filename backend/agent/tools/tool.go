package tool

import (
	"context"
	"encoding/json"
	"errors"
)

type Tool interface {
	// Name returns name of tool for LLM and registry
	Name() string

	// Description returns description for LLM from .md file
	Description() string

	// Parameters returns JSON Schema
	Parameters() json.RawMessage

	// Execute executes tool with given arguments and returns result as JSON string
	Execute(ctx context.Context, argsJSON string) (string, error)
}

// Message represents a message in the session
type Message struct {
	Role       string
	ToolCallID string
	Content    string
}

// Rights represents the rights of a tool
type Rights struct {
	ReadOnly bool
	Parallel bool
	Confirm  bool
}

// Entry represents a tool entry in the registry
type Entry struct {
	Tool   Tool
	Rights Rights
}

var (
	// ErrInvalidArgs — arguments do not match JSON Schema
	ErrInvalidArgs = errors.New("invalid arguments")

	// ErrNotFound — tool not found in registry
	ErrNotFound = errors.New("not found")

	// ErrPermission — tool does not have permission to execute
	ErrPermission = errors.New("permission denied")
)

// Errorf creates a new custom error with formatted message
func Errorf(format string, args ...any) error {
	return errors.New(format)
}
