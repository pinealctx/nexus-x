package agentic

import "context"

// Channel is the outbound interface for sending messages to a conversation platform.
// Implementations are platform-specific (e.g., Nexus IM, Slack, Discord).
// Channel is intentionally minimal — it only covers outbound operations.
type Channel interface {
	// SendText sends a plain text message to the given conversation.
	SendText(ctx context.Context, conversationID int64, text string) error

	// SendCard sends a structured card (e.g., Adaptive Card JSON) to the given conversation.
	SendCard(ctx context.Context, conversationID int64, card any) error

	// AnswerCardAction responds to a card action callback.
	AnswerCardAction(ctx context.Context, conversationID int64, actionID string, card any) error
}

// StreamingChannel extends Channel with streaming support.
// If a Channel implementation also implements StreamingChannel,
// the Engine will use streaming mode for LLM responses.
type StreamingChannel interface {
	Channel

	// StartStream begins a streaming message in the given conversation.
	// The returned StreamWriter must be used to push deltas and finalize.
	StartStream(ctx context.Context, conversationID int64) (StreamWriter, error)
}

// StreamWriter writes incremental LLM output to a streaming message.
type StreamWriter interface {
	// Push sends a text delta to the stream.
	Push(ctx context.Context, delta string) error

	// End finalizes the stream with the complete text.
	End(ctx context.Context, fullText string) error

	// Error terminates the stream with an error message.
	Error(ctx context.Context, errMsg string) error
}
