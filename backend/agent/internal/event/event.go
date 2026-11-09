package event

// Event - marker interface for agent events.
type Event interface {
	Type() string
}

// ToolStartEvent - before tool executing.
type ToolStartEvent struct {
	CallID string
	Name   string
	Args   string
}

func (e ToolStartEvent) Type() string { return "tool_start" }

// ToolEndEvent - after tool executing.
type ToolEndEvent struct {
	CallID string
	Name   string
	Result string
	Err    error
}

func (e ToolEndEvent) Type() string { return "tool_end" }

// LLMRequestEvent - before request to LLM.
type LLMRequestEvent struct {
	Step      int
	ToolCalls int
}

func (e LLMRequestEvent) Type() string { return "llm_request" }

// LLMResponseEvent - after LLM response.
type LLMResponseEvent struct {
	Step         int
	ToolCalls    int
	FinishReason string
}

func (e LLMResponseEvent) Type() string { return "llm_response" }

// LoopDetectedEvent - loop detection event.
type LoopDetectedEvent struct {
	Name  string
	Count int
}

func (e LoopDetectedEvent) Type() string { return "loop_detected" }

// StreamDeltaEvent - for every stream token.
type StreamDeltaEvent struct {
	Step int
	Text string
}

func (e StreamDeltaEvent) Type() string { return "stream_delta" }

// StreamEndEvent - end of stream.
type StreamEndEvent struct {
	Step int
	Text string // полный текст
}

func (e StreamEndEvent) Type() string { return "stream_end" }

// RepetitionDetectedEvent - repetition detected event.
type RepetitionDetectedEvent struct {
	Step int
	Text string
}

func (e RepetitionDetectedEvent) Type() string { return "repetition_detected" }

// emit calling event handler.
func Emit(on func(Event), ev Event) {
	if on != nil {
		on(ev)
	}
}
