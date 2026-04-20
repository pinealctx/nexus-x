package agentic

import "context"

// contextKey is an unexported type for context keys in this package.
type contextKey int

const (
	keyUserID contextKey = iota
	keyConversationID
	keyChannel
	keyMemory
	keyAgentID
	keyClearFlag
)

// WithUserID stores the user ID in the context.
func WithUserID(ctx context.Context, id int32) context.Context {
	return context.WithValue(ctx, keyUserID, id)
}

// UserIDFromContext retrieves the user ID from the context.
func UserIDFromContext(ctx context.Context) int32 {
	v, _ := ctx.Value(keyUserID).(int32)
	return v
}

// WithConversationID stores the conversation ID in the context.
func WithConversationID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, keyConversationID, id)
}

// ConversationIDFromContext retrieves the conversation ID from the context.
func ConversationIDFromContext(ctx context.Context) int64 {
	v, _ := ctx.Value(keyConversationID).(int64)
	return v
}

// WithChannel stores the Channel in the context.
func WithChannel(ctx context.Context, ch Channel) context.Context {
	return context.WithValue(ctx, keyChannel, ch)
}

// ChannelFromContext retrieves the Channel from the context.
func ChannelFromContext(ctx context.Context) Channel {
	v, _ := ctx.Value(keyChannel).(Channel)
	return v
}

// ContextWithMemory stores the Memory in the context.
func ContextWithMemory(ctx context.Context, m Memory) context.Context {
	return context.WithValue(ctx, keyMemory, m)
}

// MemoryFromContext retrieves the Memory from the context.
func MemoryFromContext(ctx context.Context) Memory {
	v, _ := ctx.Value(keyMemory).(Memory)
	return v
}

// ContextWithAgentID stores the agent ID in the context.
func ContextWithAgentID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyAgentID, id)
}

// AgentIDFromContext retrieves the agent ID from the context.
func AgentIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(keyAgentID).(string)
	return v
}

// clearFlag is a mutable pointer stored in context to signal that
// conversation memory should be cleared after the current LLM turn.
// A pointer is needed because context.Context is immutable — mutations
// via the shared pointer are visible to all holders of the context.
type clearFlag struct {
	cleared bool
}

// ContextWithClearFlag injects a new clear flag into the context.
func ContextWithClearFlag(ctx context.Context) context.Context {
	return context.WithValue(ctx, keyClearFlag, &clearFlag{})
}

// MarkContextCleared signals that memory should be cleared after this turn.
func MarkContextCleared(ctx context.Context) {
	if f, _ := ctx.Value(keyClearFlag).(*clearFlag); f != nil {
		f.cleared = true
	}
}

// IsContextCleared reports whether MarkContextCleared was called in this context.
func IsContextCleared(ctx context.Context) bool {
	if f, _ := ctx.Value(keyClearFlag).(*clearFlag); f != nil {
		return f.cleared
	}
	return false
}
