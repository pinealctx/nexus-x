package agentic

import (
	"context"

	sharedv1 "github.com/pinealctx/nexus-proto/gen/go/shared/v1"
)

// IncomingUpdate represents an inbound event from Nexus IM.
// It is a superset of the proto types — the original proto is always
// available via Envelope / CardAction, plus framework-level fields
// (Channel, Metadata) and convenience fields extracted from proto.
type IncomingUpdate struct {
	// ── Original proto (complete, never loses information) ──

	// Envelope is the full message envelope for message events.
	// Nil for non-message events (e.g. CardAction).
	Envelope *sharedv1.MessageEnvelope

	// CardAction is the card action payload for Action.Submit events.
	// Nil for message events.
	CardAction *sharedv1.CardActionPayload

	// ── Framework-level fields (not in proto) ──

	// Channel is the outbound channel for sending replies.
	Channel Channel

	// Metadata holds arbitrary data that middleware can attach.
	Metadata map[string]any

	// ── Convenience fields (extracted from proto) ──

	// UserID is the sender (from Envelope.SenderId or CardAction.SenderId).
	UserID int32

	// ConversationID is the conversation context.
	ConversationID int64

	// MessageID is the message identifier.
	MessageID int64

	// Text is the text content extracted from TEXT/MARKDOWN message bodies.
	// Empty for other message types and CardAction events.
	Text string
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
