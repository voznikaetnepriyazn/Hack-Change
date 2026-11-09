package stream

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/config"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/event"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/lib/repeat"
)

// StreamLLM is a wrapper around openai.ChatCompletionStream
func StreamLLM(ctx context.Context, cfg config.AgentConfig, step int) (openai.ChatCompletionMessage, error) {
	var finishReason string

	llmCtx, cancel := context.WithTimeout(ctx, cfg.Limits.LLMTimeout)
	defer cancel()

	req := openai.ChatCompletionRequest{
		Model:       cfg.Model,
		Messages:    cfg.Session.Messages(),
		Tools:       cfg.Registry.Schemas(),
		Temperature: 0,
		Stream:      true,
	}

	for i, msg := range req.Messages {
		slog.Debug("Request history",
			"index", i,
			"role", msg.Role,
			"content", msg.Content,
			"tool_calls", len(msg.ToolCalls),
		)
	}

	stream, err := cfg.Client.CreateChatCompletionStream(llmCtx, req)
	if err != nil {
		return openai.ChatCompletionMessage{}, fmt.Errorf("stream error: %w", err)
	}
	defer stream.Close()

	slog.Debug("Tools in request", "tools", req.Tools)

	var (
		fullContent      strings.Builder
		repetition       = repeat.NewRepetitionDetector(256)
		toolCallsByIndex = make(map[int]*openai.ToolCall)
	)

	for {
		delta, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return openai.ChatCompletionMessage{}, fmt.Errorf("stream recv error: %w", err)
		}

		if len(delta.Choices) == 0 {
			continue
		}

		choice := delta.Choices[0]
		if choice.FinishReason != "" {
			finishReason = string(choice.FinishReason)
		}

		chunk := choice.Delta.Content

		if chunk != "" {
			fullContent.WriteString(chunk)

			event.Emit(cfg.On, event.StreamDeltaEvent{
				Step: step,
				Text: chunk,
			})

			if repetition.Feed(chunk) {
				slog.Warn("repetition detected", "step", step)

				event.Emit(cfg.On, event.RepetitionDetectedEvent{
					Step: step,
					Text: chunk,
				})

				cancel()
				return openai.ChatCompletionMessage{},
					fmt.Errorf("LLM stuck in repetition loop")
			}
		}

		for _, tc := range choice.Delta.ToolCalls {
			idx := 0
			if tc.Index != nil {
				idx = *tc.Index
			}

			existing, ok := toolCallsByIndex[idx]
			if !ok {
				existing = &openai.ToolCall{
					ID:   tc.ID,
					Type: tc.Type,
					Function: openai.FunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
				toolCallsByIndex[idx] = existing
			} else {
				existing.Function.Arguments += tc.Function.Arguments

				if existing.Function.Name == "" {
					existing.Function.Name = tc.Function.Name
				}

				if existing.ID == "" {
					existing.ID = tc.ID
				}
			}
		}
	}

	var toolCalls []openai.ToolCall

	if len(toolCallsByIndex) > 0 {
		indices := make([]int, 0, len(toolCallsByIndex))
		for idx := range toolCallsByIndex {
			indices = append(indices, idx)
		}
		sort.Ints(indices)

		for _, idx := range indices {
			toolCalls = append(toolCalls, *toolCallsByIndex[idx])
		}
	}

	event.Emit(cfg.On, event.StreamEndEvent{
		Step: step,
		Text: fullContent.String(),
	})

	if finishReason != "stop" && finishReason != "tool_calls" {
		slog.Warn("unexpected finish_reason", "reason", finishReason)
	}

	return openai.ChatCompletionMessage{
		Role:      openai.ChatMessageRoleAssistant,
		Content:   fullContent.String(),
		ToolCalls: toolCalls,
	}, nil
}
