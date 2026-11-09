package config

import (
	"time"

	openai "github.com/sashabaranov/go-openai"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/agent/session"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/event"
	detection "github.com/voznikaetnepriyazn/Hack-Change/agent/internal/lib/loop"
	tool "github.com/voznikaetnepriyazn/Hack-Change/agent/tools"
)

// Limits of one agent step.
type Limits struct {
	MaxIterations   int
	Deadline        time.Duration
	LLMTimeout      time.Duration
	ToolConcurrency int
}

// DefaultLimits limit by default.
func DefaultLimits() Limits {
	return Limits{
		MaxIterations:   20,
		Deadline:        5 * time.Minute,
		LLMTimeout:      240 * time.Second,
		ToolConcurrency: 3,
	}
}

// AgentConfig - one agent configuration.
type AgentConfig struct {
	// dependencies
	Client       *openai.Client
	Registry     *tool.Registry
	Model        string
	SystemPrompt string
	Stream       bool

	// limits
	Limits Limits
	Loop   *detection.LoopDetector

	// for events
	On func(event.Event)

	// Session do not serialize, only for runtime usage
	Session *session.Session `json:"-"`
}
