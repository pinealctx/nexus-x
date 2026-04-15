package agentic

import "context"

// MessageType identifies the kind of incoming message.
// Mirrors the platform's message type enum without importing proto.
type MessageType int8

const (
	MessageTypeUnknown  MessageType = 0
	MessageTypeText     MessageType = 1
	MessageTypeImage    MessageType = 2
	MessageTypeAudio    MessageType = 3
	MessageTypeVideo    MessageType = 4
	MessageTypeFile     MessageType = 5
	MessageTypeMarkdown MessageType = 6
	MessageTypeCard     MessageType = 7
	MessageTypeStream   MessageType = 8
)

// IncomingUpdate represents an inbound event from a conversation platform.
// It carries the raw event data plus a reference to the Channel for replies.
type IncomingUpdate struct {
	// UserID identifies the user who triggered this update.
	UserID int32

	// ConversationID identifies the conversation context.
	ConversationID int64

	// MessageID is the platform-specific message identifier (for dedup).
	MessageID int64

	// Type is the message type (text, image, markdown, card, etc.).
	// Zero value (MessageTypeUnknown) for non-message events like card actions.
	Type MessageType

	// Text is the message text content.
	// For TEXT/MARKDOWN messages, this is the full text.
	// For other types, this may be empty or contain a text representation.
	Text string

	// Channel is the outbound channel for sending replies.
	Channel Channel

	// CardAction holds card action data if this update is a card callback.
	CardAction *CardAction

	// RawBody holds the platform-specific raw message body.
	// For Nexus, this is *sharedv1.MessageBody.
	// Allows advanced agents to access media, entities, etc.
	RawBody any

	// Metadata holds arbitrary platform-specific data.
	Metadata map[string]any
}

// CardAction represents a user interaction with a structured card.
type CardAction struct {
	// ActionID is the server-assigned action identifier.
	ActionID string
	// Verb is the action verb (from Action.Submit's id/verb property).
	Verb string
	// UserID is the user who triggered the action.
	UserID int32
	// ConversationID is the conversation containing the card.
	ConversationID int64
	// MessageID is the card message ID.
	MessageID int64
	// Data holds the parsed action_data JSON.
	Data map[string]any
}

// Handler processes an incoming update.
type Handler func(ctx context.Context, update *IncomingUpdate) error

// Middleware wraps a Handler to add cross-cutting behavior.
type Middleware func(next Handler) Handler

// Chain composes middlewares into a single Middleware, applied left to right.
// Chain(A, B, C)(handler) == A(B(C(handler))).
func Chain(mws ...Middleware) Middleware {
	return func(next Handler) Handler {
		for i := len(mws) - 1; i >= 0; i-- {
			next = mws[i](next)
		}
		return next
	}
}
