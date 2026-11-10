package config

import (
	"time"

	openai "github.com/sashabaranov/go-openai"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/agent/session"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/event"
	detection "github.com/voznikaetnepriyazn/Hack-Change/agent/internal/lib/loop"
	tool "github.com/voznikaetnepriyazn/Hack-Change/agent/tools"
)

// Limits bounds one complete agent run and the work within each iteration.
type Limits struct {
	MaxIterations int
	// Deadline bounds the complete run, including waiting for the session and tool execution.
	// A zero value disables the overall deadline; a negative value uses the default.
	Deadline time.Duration
	// LLMTimeout bounds each individual LLM request.
	LLMTimeout time.Duration
	// ToolConcurrency is the maximum number of parallel read-only tools.
	ToolConcurrency int
}

// DefaultLimits returns the default run and request limits.
func DefaultLimits() Limits {
	return Limits{
		MaxIterations:   20,
		Deadline:        5 * time.Minute,
		LLMTimeout:      240 * time.Second,
		ToolConcurrency: tool.DefaultToolConcurrency,
	}
}

// WithDefaults replaces unset iteration, LLM timeout, and tool concurrency values
// with defaults. A zero Deadline intentionally means no overall deadline.
func (l Limits) WithDefaults() Limits {
	defaults := DefaultLimits()
	if l.MaxIterations <= 0 {
		l.MaxIterations = defaults.MaxIterations
	}
	if l.Deadline < 0 {
		l.Deadline = defaults.Deadline
	}
	if l.LLMTimeout <= 0 {
		l.LLMTimeout = defaults.LLMTimeout
	}
	if l.ToolConcurrency <= 0 {
		l.ToolConcurrency = defaults.ToolConcurrency
	}
	return l
}

// AgentConfig contains dependencies and per-run settings for the agent.
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
