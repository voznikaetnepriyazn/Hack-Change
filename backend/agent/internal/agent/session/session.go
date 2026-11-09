package session

import (
	"sync"

	openai "github.com/sashabaranov/go-openai"
)

// Session is a storage of one dialogue
type Session struct {
	mu       sync.Mutex
	messages []openai.ChatCompletionMessage
}

// NewSession creates a new session with system promt
func NewSession(systemPrompt string) *Session {
	return &Session{
		messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem,
				Content: systemPrompt},
		},
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

// AddToolResult adds tool result to the session
func (s *Session) AddToolResult(toolCallId, msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, openai.ChatCompletionMessage{
		Role:       openai.ChatMessageRoleTool,
		Content:    msg,
		ToolCallID: toolCallId,
	})
}

// AddMessage adds message from llm to the session
func (s *Session) AddMessage(msg openai.ChatCompletionMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, msg)
}

// Messages returns copy of history from the session
func (s *Session) Messages() []openai.ChatCompletionMessage {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]openai.ChatCompletionMessage, len(s.messages))
	copy(result, s.messages)
	return result
}
