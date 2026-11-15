package run

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	config "github.com/voznikaetnepriyazn/Hack-Change/agent/internal/config"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/event"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/lib/validate"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/stream"

	openai "github.com/sashabaranov/go-openai"
)

// RunResult - result of agent's step.
type RunResult struct {
	Output     string
	Err        error
	Iterations int
	ToolsUsed  []string
}

// Run is a ReAct-cycle with wrappers. Turns sharing a session are serialized.
func Run(ctx context.Context, cfg config.AgentConfig, userMessage string) RunResult {
	res := RunResult{}
	cfg.Limits = cfg.Limits.WithDefaults()

	if cfg.Limits.Deadline > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.Limits.Deadline)
		defer cancel()
	}

	unlockTurn, err := cfg.Session.LockTurn(ctx)
	if err != nil {
		slog.Error("Failed to acquire session lock", "error", err)
		res.Err = fmt.Errorf("failed to acquire session turn: %w", err)
		return res
	}
	defer unlockTurn()

	if cfg.Loop != nil {
		cfg.Loop.Reset()
	}

	cfg.Session.AddUserInput(userMessage)

	for step := 0; step < cfg.Limits.MaxIterations; step++ {
		if ctx.Err() != nil {
			res.Err = fmt.Errorf("deadline exceeded: %w", ctx.Err())
			return res
		}
		res.Iterations = step + 1

		event.Emit(cfg.On, event.LLMRequestEvent{Step: step + 1})

		var msg openai.ChatCompletionMessage

		if cfg.Stream {
			var err error
			msg, err = stream.StreamLLM(ctx, cfg, step+1)
			if err != nil {
				res.Err = err
				return res
			}
		} else {
			llmCtx, cancel := context.WithTimeout(ctx, cfg.Limits.LLMTimeout)
			resp, err := cfg.Client.CreateChatCompletion(llmCtx, openai.ChatCompletionRequest{
				Model:       cfg.Model,
				Messages:    cfg.Session.Messages(),
				Tools:       cfg.Registry.Schemas(),
				Temperature: 0,
			})
			cancel()

			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					res.Err = fmt.Errorf("LLM timeout on step %d", step+1)
					return res
				}
				res.Err = fmt.Errorf("LLM error on step %d: %w", step+1, err)
				return res
			}

			if len(resp.Choices) == 0 {
				res.Err = fmt.Errorf("empty response from LLM on step %d", step+1)
				return res
			}

			msg = resp.Choices[0].Message
			event.Emit(cfg.On, event.LLMResponseEvent{
				Step:         step + 1,
				ToolCalls:    len(msg.ToolCalls),
				FinishReason: string(resp.Choices[0].FinishReason),
			})
		}

		cfg.Session.AddMessage(msg)

		if len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				slog.Debug("LLM Response: Tool Call",
					"id", tc.ID,
					"name", tc.Function.Name,
					"args_raw", tc.Function.Arguments,
				)
			}
		} else {
			slog.Debug("LLM Response: Text",
				"content_len", len(msg.Content),
				"content_preview", msg.Content,
			)
		}

		if len(msg.ToolCalls) == 0 {
			res.Output = msg.Content
			return res
		}

		var validCalls []openai.ToolCall

		for i, tc := range msg.ToolCalls {
			slog.Debug("raw tool_call",
				"i", i,
				"id", tc.ID,
				"name", tc.Function.Name,
				"args_type", fmt.Sprintf("%T", tc.Function.Arguments),
				"args", tc.Function.Arguments,
			)

			if err := validate.ValidateToolCall(tc, cfg.Registry); err != nil {
				slog.Warn("invalid tool_call", "error", err)
				cfg.Session.AddToolResult(tc.ID, fmt.Sprintf("Error: %v", err))
				continue
			}

			validCalls = append(validCalls, tc)

			event.Emit(cfg.On, event.ToolStartEvent{
				CallID: tc.ID,
				Name:   tc.Function.Name,
				Args:   tc.Function.Arguments,
			})
		}

		if len(validCalls) == 0 {
			continue
		}

		results := cfg.Registry.ExecuteBatch(ctx, validCalls, cfg.Limits.ToolConcurrency)

		for i, m := range results {
			call := validCalls[i]
			originalContent := m.Content

			if cfg.Loop != nil {
				if err := cfg.Loop.Record(call.Function.Name, call.Function.Arguments, originalContent); err != nil {
					slog.Warn("loop detected", "tool", call.Function.Name, "error", err)
					event.Emit(cfg.On, event.LoopDetectedEvent{Name: call.Function.Name})
					m.Content = fmt.Sprintf("Error: %v. Please try a different approach.", err)
				}
			}

			event.Emit(cfg.On, event.ToolEndEvent{
				CallID: call.ID,
				Name:   call.Function.Name,
				Result: m.Content,
			})

			res.ToolsUsed = append(res.ToolsUsed, call.Function.Name)
			cfg.Session.AddToolResult(m.ToolCallID, m.Content)
		}
	}

	res.Err = fmt.Errorf("max iterations exceeded (%d)", cfg.Limits.MaxIterations)
	return res
}
