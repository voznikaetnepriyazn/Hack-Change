package session

import (
	"context"
	"sync"

	openai "github.com/sashabaranov/go-openai"
)

// Session stores one dialogue and serializes turns that use its history.
type Session struct {
	turnMu   chan struct{}
	mu       sync.Mutex
	messages []openai.ChatCompletionMessage
}

// NewSession creates a new session with a system prompt.
func NewSession(systemPrompt string) *Session {
	return &Session{
		turnMu: make(chan struct{}, 1),
		messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem,
				Content: systemPrompt},
		},
	}
}

// LockTurn serializes complete agent turns that share this session.
// The returned function releases the lock and should be deferred.
// If ctx is canceled while waiting, LockTurn returns its error without acquiring the lock.
func (s *Session) LockTurn(ctx context.Context) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	select {
	case s.turnMu <- struct{}{}:
		if err := ctx.Err(); err != nil {
			<-s.turnMu
			return nil, err
		}
		return func() { <-s.turnMu }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// AddUserInput adds user input to the session
func (s *Session) AddUserInput(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: msg,
	})
}

// AddToolResult adds a tool result to the session.
func (s *Session) AddToolResult(toolCallId, msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, openai.ChatCompletionMessage{
		Role:       openai.ChatMessageRoleTool,
		Content:    msg,
		ToolCallID: toolCallId,
	})
}

// AddMessage adds a message from the LLM to the session.
func (s *Session) AddMessage(msg openai.ChatCompletionMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, cloneMessage(msg))
}

// Messages returns a copy of the session history.
func (s *Session) Messages() []openai.ChatCompletionMessage {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]openai.ChatCompletionMessage, len(s.messages))
	for i, message := range s.messages {
		result[i] = cloneMessage(message)
	}
	return result
}

func cloneMessage(message openai.ChatCompletionMessage) openai.ChatCompletionMessage {
	if message.FunctionCall != nil {
		functionCall := *message.FunctionCall
		message.FunctionCall = &functionCall
	}

	if message.ToolCalls != nil {
		message.ToolCalls = append([]openai.ToolCall(nil), message.ToolCalls...)
		for i := range message.ToolCalls {
			if message.ToolCalls[i].Index != nil {
				index := *message.ToolCalls[i].Index
				message.ToolCalls[i].Index = &index
			}
		}
	}

	if message.MultiContent != nil {
		message.MultiContent = append([]openai.ChatMessagePart(nil), message.MultiContent...)
		for i := range message.MultiContent {
			if message.MultiContent[i].ImageURL != nil {
				imageURL := *message.MultiContent[i].ImageURL
				message.MultiContent[i].ImageURL = &imageURL
			}
		}
	}

	return message
}
