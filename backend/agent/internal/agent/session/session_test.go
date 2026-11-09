package session

import (
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/sashabaranov/go-openai"
)

func TestNewSessionStartsWithSystemPrompt(t *testing.T) {
	const prompt = "You are a helpful assistant."
	s := NewSession(prompt)

	got := s.Messages()
	want := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: prompt},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Messages() = %#v, want %#v", got, want)
	}
}

func TestAddUserInput(t *testing.T) {
	s := NewSession("system")
	s.AddUserInput("hello")

	got := s.Messages()
	want := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: "system"},
		{Role: openai.ChatMessageRoleUser, Content: "hello"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Messages() = %#v, want %#v", got, want)
	}
}

func TestAddToolResult(t *testing.T) {
	s := NewSession("system")
	s.AddToolResult("call-123", `{"ok":true}`)

	got := s.Messages()
	want := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: "system"},
		{
			Role:       openai.ChatMessageRoleTool,
			Content:    `{"ok":true}`,
			ToolCallID: "call-123",
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Messages() = %#v, want %#v", got, want)
	}
}

func TestAddMessage(t *testing.T) {
	s := NewSession("system")
	message := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: "How can I help?",
	}
	s.AddMessage(message)

	got := s.Messages()
	want := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: "system"},
		message,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Messages() = %#v, want %#v", got, want)
	}
}

func TestMessagesReturnsIndependentCopy(t *testing.T) {
	s := NewSession("system")
	s.AddUserInput("original")

	first := s.Messages()
	first[0].Content = "changed system prompt"
	first[1] = openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: "changed message",
	}
	first = append(first, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: "appended outside session",
	})

	want := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: "system"},
		{Role: openai.ChatMessageRoleUser, Content: "original"},
	}
	if got := s.Messages(); !reflect.DeepEqual(got, want) {
		t.Errorf("session history changed through returned copy: got %#v, want %#v", got, want)
	}
}

func TestConcurrentAddsAreAllPreserved(t *testing.T) {
	const count = 100
	s := NewSession("system")

	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s.AddUserInput(fmt.Sprintf("message-%d", i))
		}(i)
	}
	wg.Wait()

	got := s.Messages()
	if len(got) != count+1 {
		t.Fatalf("Messages() returned %d messages, want %d", len(got), count+1)
	}
	if got[0].Role != openai.ChatMessageRoleSystem || got[0].Content != "system" {
		t.Fatalf("first message = %#v, want system prompt", got[0])
	}

	seen := make(map[string]int, count)
	for _, message := range got[1:] {
		if message.Role != openai.ChatMessageRoleUser {
			t.Errorf("message role = %q, want user", message.Role)
		}
		seen[message.Content]++
	}
	for i := 0; i < count; i++ {
		content := fmt.Sprintf("message-%d", i)
		if seen[content] != 1 {
			t.Errorf("message %q appeared %d times, want once", content, seen[content])
		}
	}
}
