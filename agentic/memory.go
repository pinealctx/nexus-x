package agentic

import "context"

// Memory stores and retrieves conversation history.
// Implementations can scope by user, conversation, or both.
type Memory interface {
	// Load retrieves stored messages for the given key.
	Load(ctx context.Context, key MemoryKey) ([]Message, error)

	// Save persists messages for the given key.
	Save(ctx context.Context, key MemoryKey, msgs []Message) error

	// Clear removes all stored messages for the given key.
	Clear(ctx context.Context, key MemoryKey) error
}

// MemoryKey identifies a memory scope.
type MemoryKey struct {
	AgentID        string
	UserID         int32
	ConversationID int64 // 0 means per-user scope (shared across conversations).
}

// Message is a conversation message stored in memory.
//
// For simple text-only agents, use Role + Content.
// For agents with tool use, populate Parts to preserve the full
// tool call/result history across turns.
type Message struct {
	Role    string `json:"role"`    // "user", "assistant", "system", "tool"
	Content string `json:"content"` // text content (convenience field)

	// Parts holds the full message parts (text, tool calls, tool results, etc.).
	// When non-nil, this takes precedence over Content for Fantasy conversion.
	// When nil, Content is used as a single TextPart.
	Parts []MessagePart `json:"parts,omitempty"`
}

// MessagePartType identifies the kind of content in a MessagePart.
type MessagePartType string

const (
	// MessagePartTypeText represents a plain text part.
	MessagePartTypeText MessagePartType = "text"

	// MessagePartTypeToolCall represents a tool invocation by the LLM.
	MessagePartTypeToolCall MessagePartType = "tool_call"

	// MessagePartTypeToolResult represents the result of a tool execution.
	MessagePartTypeToolResult MessagePartType = "tool_result"
)

// MessagePart represents a single part of a message.
// This is a simplified, JSON-serializable representation of Fantasy's MessagePart.
type MessagePart struct {
	Type MessagePartType `json:"type"`

	// Text content (type=MessagePartTypeText).
	Text string `json:"text,omitempty"`

	// Tool call fields (type=MessagePartTypeToolCall).
	ToolCallID string `json:"tool_call_id,omitempty"`
	ToolName   string `json:"tool_name,omitempty"`
	Input      string `json:"input,omitempty"` // JSON string

	// Tool result fields (type=MessagePartTypeToolResult).
	Output  string `json:"output,omitempty"`
	IsError bool   `json:"is_error,omitempty"`
}

// Summarizer compresses conversation history to stay within token limits.
// If nil, the Engine truncates by message count.
type Summarizer interface {
	Summarize(ctx context.Context, msgs []Message) ([]Message, error)
}
