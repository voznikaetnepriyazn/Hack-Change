package agent

import (
	"context"
	_ "embed"
	"sync"

	"log/slog"

	"github.com/sashabaranov/go-openai"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/agent/run"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/agent/session"
	config "github.com/voznikaetnepriyazn/Hack-Change/agent/internal/config"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/event"
	direction "github.com/voznikaetnepriyazn/Hack-Change/agent/internal/lib/loop"
	tool "github.com/voznikaetnepriyazn/Hack-Change/agent/tools"
)

const (
	loopWindow    = 20
	loopThreshold = 5
)

//go:embed system-prompt.md
var SystemPromt string

// Agent runs the assistant and keeps a separate conversation history per session.
type Agent struct {
	baseCfg  config.AgentConfig
	sessions map[string]*session.Session
	mu       sync.Mutex
}

// NewAgent creates an agent with default limits and event logging.
func NewAgent(client *openai.Client, registry *tool.Registry, model string, stream bool) *Agent {
	return &Agent{
		baseCfg: config.AgentConfig{
			Client:   client,
			Registry: registry,
			Model:    model,
			Stream:   stream,
			Limits:   config.DefaultLimits(),
			On:       defaultEventHandler,
		},
		sessions: make(map[string]*session.Session),
	}
}

// Run sends a user message and returns the assistant's final response.
func (a *Agent) Run(ctx context.Context, sessionID string, userMessage string) (string, error) {
	result := run.Run(ctx, a.configForSession(sessionID), userMessage)
	return result.Output, result.Err
}

// RunWithResult sends a user message and returns the full run result.
func (a *Agent) RunWithResult(ctx context.Context, sessionID string, userMessage string) run.RunResult {
	return run.Run(ctx, a.configForSession(sessionID), userMessage)
}

// WithLimits overrides the limits used by subsequent runs.
// Configure limits before starting runs; concurrent configuration changes are not supported.
func (a *Agent) WithLimits(limits config.Limits) *Agent {
	a.baseCfg.Limits = limits
	return a
}

// WithEventHandler sets the event handler used by subsequent runs.
// Configure the handler before starting runs; concurrent configuration changes are not supported.
func (a *Agent) WithEventHandler(on func(event.Event)) *Agent {
	a.baseCfg.On = on
	return a
}

// WithStream enables or disables streaming for subsequent runs.
// Configure streaming before starting runs; concurrent configuration changes are not supported.
func (a *Agent) WithStream(enabled bool) *Agent {
	a.baseCfg.Stream = enabled
	return a
}

// RunWithChannel runs one turn and sends streaming text deltas to outChan.
// If the channel is full or has no receiver ready, that delta is dropped.
// The caller must keep the channel open until RunWithChannel returns.
func (a *Agent) RunWithChannel(ctx context.Context, sessionID string, userMessage string, outChan chan<- string) (string, error) {
	cfg := a.configForSession(sessionID)

	originalOn := cfg.On

	cfg.On = func(ev event.Event) {
		if originalOn != nil {
			originalOn(ev)
		}

		if delta, ok := ev.(event.StreamDeltaEvent); ok {
			select {
			case outChan <- delta.Text:
			default:
				// channel is full, skip this delta to avoid blocking
				slog.Warn("Channel overflow, skipping delta")
			}
		}
	}

	result := run.Run(ctx, cfg, userMessage)
	return result.Output, result.Err
}

// configForSession returns a run-specific config and the persistent history for sessionID.
func (a *Agent) configForSession(sessionID string) config.AgentConfig {
	a.mu.Lock()
	sess, exists := a.sessions[sessionID]
	if !exists {
		sess = session.NewSession(SystemPromt)
		a.sessions[sessionID] = sess
	}
	a.mu.Unlock()

	cfg := a.baseCfg
	cfg.Session = sess
	cfg.Loop = direction.NewLoopDetector(loopWindow, loopThreshold)
	return cfg
}

// defaultEventHandler - logging handler.
func defaultEventHandler(ev event.Event) {
	switch e := ev.(type) {
	case event.LLMRequestEvent:
		slog.Debug("LLM request was sent", "step", e.Step, "tool_calls_planned", e.ToolCalls)
	case event.LLMResponseEvent:
		slog.Debug("LLM response", "step", e.Step, "tool_calls", e.ToolCalls)
	case event.ToolStartEvent:
		slog.Debug("tool start", "name", e.Name, "args", e.Args)
	case event.ToolEndEvent:
		slog.Debug("tool end", "name", e.Name)
	case event.LoopDetectedEvent:
		slog.Warn("loop detected", "name", e.Name)
	case event.RepetitionDetectedEvent:
		slog.Warn("repetition detected", "text", e.Text)
	case event.StreamDeltaEvent:
		slog.Debug("stream delta", "text", e.Text)
	case event.StreamEndEvent:
		slog.Debug("stream end")
	}
}
