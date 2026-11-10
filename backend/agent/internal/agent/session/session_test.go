package session

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

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

func TestMessagesDeepCopiesNestedFields(t *testing.T) {
	index := 1
	original := openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleAssistant,
		FunctionCall: &openai.FunctionCall{
			Name:      "original-function",
			Arguments: `{"value":1}`,
		},
		ToolCalls: []openai.ToolCall{{
			Index: &index,
			Function: openai.FunctionCall{
				Name:      "original-tool",
				Arguments: `{"value":2}`,
			},
		}},
		MultiContent: []openai.ChatMessagePart{{
			Type:     openai.ChatMessagePartTypeImageURL,
			ImageURL: &openai.ChatMessageImageURL{URL: "https://example.test/original"},
		}},
	}
	s := NewSession("system")
	s.AddMessage(original)

	original.FunctionCall.Name = "changed-after-add"
	original.ToolCalls[0].Function.Name = "changed-after-add"
	*original.ToolCalls[0].Index = 99
	original.MultiContent[0].ImageURL.URL = "https://example.test/changed-after-add"

	got := s.Messages()[1]
	if got.FunctionCall.Name != "original-function" ||
		got.ToolCalls[0].Function.Name != "original-tool" ||
		*got.ToolCalls[0].Index != 1 ||
		got.MultiContent[0].ImageURL.URL != "https://example.test/original" {
		t.Fatalf("AddMessage() retained caller-owned nested data: %#v", got)
	}

	got.FunctionCall.Name = "changed-returned-copy"
	got.ToolCalls[0].Function.Name = "changed-returned-copy"
	*got.ToolCalls[0].Index = 100
	got.MultiContent[0].ImageURL.URL = "https://example.test/changed-returned-copy"

	stored := s.Messages()[1]
	if stored.FunctionCall.Name != "original-function" ||
		stored.ToolCalls[0].Function.Name != "original-tool" ||
		*stored.ToolCalls[0].Index != 1 ||
		stored.MultiContent[0].ImageURL.URL != "https://example.test/original" {
		t.Fatalf("Messages() exposed nested session-owned data: %#v", stored)
	}
}

func TestLockTurnSerializesTurnOwnership(t *testing.T) {
	s := NewSession("system")
	unlockFirst, err := s.LockTurn(context.Background())
	if err != nil {
		t.Fatalf("LockTurn() for first turn returned unexpected error: %v", err)
	}
	defer func() { unlockFirst() }()

	type lockResult struct {
		unlock func()
		err    error
	}
	acquiredSecond := make(chan lockResult, 1)
	go func() {
		unlock, err := s.LockTurn(context.Background())
		acquiredSecond <- lockResult{unlock: unlock, err: err}
	}()

	select {
	case result := <-acquiredSecond:
		if result.unlock != nil {
			result.unlock()
		}
		t.Fatal("second turn acquired the lock while the first turn was active")
	case <-time.After(25 * time.Millisecond):
	}

	unlockFirst()
	unlockFirst = func() {}

	select {
	case result := <-acquiredSecond:
		if result.err != nil {
			t.Fatalf("LockTurn() for second turn returned unexpected error: %v", result.err)
		}
		result.unlock()
	case <-time.After(time.Second):
		t.Fatal("second turn did not acquire the lock after the first turn ended")
	}
}

func TestLockTurnHonorsCanceledContextWhileWaiting(t *testing.T) {
	s := NewSession("system")
	unlock, err := s.LockTurn(context.Background())
	if err != nil {
		t.Fatalf("LockTurn() returned unexpected error: %v", err)
	}
	defer unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	type lockResult struct {
		unlock func()
		err    error
	}
	resultCh := make(chan lockResult, 1)
	go func() {
		unlock, err := s.LockTurn(ctx)
		resultCh <- lockResult{unlock: unlock, err: err}
	}()

	cancel()
	select {
	case result := <-resultCh:
		if result.unlock != nil {
			result.unlock()
		}
		if result.err != context.Canceled {
			t.Errorf("LockTurn() error = %v, want %v", result.err, context.Canceled)
		}
	case <-time.After(time.Second):
		t.Fatal("LockTurn() did not return after its context was canceled")
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
