package session

// FunctionCall is the OpenAI-style nested function call format.
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolCall represents a single tool invocation in a message.
type ToolCall struct {
	ID       string        `json:"id"`
	Name     string        `json:"name,omitempty"`
	Args     string        `json:"arguments,omitempty"`
	Function *FunctionCall `json:"function,omitempty"`
}

// Message is a provider-agnostic chat message.
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}
