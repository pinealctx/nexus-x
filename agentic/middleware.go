package agentic

import (
	"context"

	sharedv1 "github.com/pinealctx/nexus-proto/gen/go/shared/v1"

	"github.com/pinealctx/nexus-x/nxlog"
	"github.com/pinealctx/nexus-x/nxutil"
	"go.uber.org/zap"
)

// AgentFilterMiddleware filters messages based on Agent conversation rules:
//   - Self-messages (sent by the agent itself) are always skipped.
//   - Private chat: process all non-self messages.
//   - Group chat: only process messages that @mention the agent (by user ID or @all).
//   - CardAction events: always pass through (they are explicit user interactions).
//
// selfID is the agent's own user ID (from client.SelfUserID).
func AgentFilterMiddleware(selfID int32) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, update *IncomingUpdate) error {
			// CardAction events always pass through.
			if update.CardAction != nil {
				return next(ctx, update)
			}

			// Non-message events pass through.
			if update.Envelope == nil {
				return next(ctx, update)
			}

			// Skip self-messages.
			if update.UserID == selfID {
				return nil
			}

			// Private chat: process all non-self messages.
			if nxutil.IsPrivateConversation(update.ConversationID) {
				return next(ctx, update)
			}

			// Group chat: only process if agent is mentioned.
			if isMentioned(update.Envelope, selfID) {
				return next(ctx, update)
			}

			// Group message without @mention — skip silently.
			return nil
		}
	}
}

// isMentioned checks if the message mentions the given user ID or uses @all.
func isMentioned(env *sharedv1.MessageEnvelope, userID int32) bool {
	if env == nil || env.Body == nil {
		return false
	}

	var entities []*sharedv1.MessageEntity
	switch body := env.Body.Content.(type) {
	case *sharedv1.MessageBody_Text:
		if body.Text != nil {
			entities = body.Text.Entities
		}
	case *sharedv1.MessageBody_Markdown:
		if body.Markdown != nil {
			entities = body.Markdown.Entities
		}
	}

	for _, e := range entities {
		if e.Type != sharedv1.MessageEntityType_MESSAGE_ENTITY_TYPE_MENTION {
			continue
		}
		m := e.GetMention()
		if m == nil {
			continue
		}
		if m.IsAll || m.UserId == userID {
			return true
		}
	}
	return false
}

// DedupMiddleware filters duplicate messages using the provided Deduplicator.
func DedupMiddleware(dedup Deduplicator) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, update *IncomingUpdate) error {
			if update.MessageID != 0 && dedup.IsDuplicate(update.ConversationID, update.MessageID) {
				nxlog.Debug("dedup: skipping duplicate message",
					zap.Int64("conversation_id", update.ConversationID),
					zap.Int64("message_id", update.MessageID),
				)
				return nil
			}
			return next(ctx, update)
		}
	}
}

// RateLimitMiddleware rejects requests from users who exceed rate limits.
// onLimited is called when a user is rate-limited; if nil, the update is silently dropped.
func RateLimitMiddleware(limiter RateLimiter, onLimited Handler) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, update *IncomingUpdate) error {
			allowed, err := limiter.Allow(ctx, update.UserID)
			if err != nil {
				nxlog.Warn("rate limiter error, allowing request", zap.Error(err), zap.Int32("user_id", update.UserID))
				return next(ctx, update)
			}
			if !allowed {
				if onLimited != nil {
					return onLimited(ctx, update)
				}
				return nil
			}
			return next(ctx, update)
		}
	}
}

// CredentialGateMiddleware blocks users without valid credentials.
// onMissing is called when credentials are missing (e.g., send a "connect" card).
func CredentialGateMiddleware(checker CredentialChecker, onMissing Handler) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, update *IncomingUpdate) error {
			ok, err := checker.Check(ctx, update.UserID)
			if err != nil {
				nxlog.Warn("credential check error", zap.Error(err), zap.Int32("user_id", update.UserID))
				return next(ctx, update) // fail open
			}
			if !ok {
				if onMissing != nil {
					return onMissing(ctx, update)
				}
				return nil
			}
			return next(ctx, update)
		}
	}
}

// LoggingMiddleware logs each incoming update at debug level.
func LoggingMiddleware() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, update *IncomingUpdate) error {
			nxlog.Debug("incoming update",
				zap.Int32("user_id", update.UserID),
				zap.Int64("conversation_id", update.ConversationID),
				zap.Int64("message_id", update.MessageID),
				zap.String("text", update.Text),
			)
			err := next(ctx, update)
			if err != nil {
				nxlog.Error("handler error",
					zap.Int32("user_id", update.UserID),
					zap.Int64("conversation_id", update.ConversationID),
					zap.Error(err),
				)
			}
			return err
		}
	}
}

// RecoveryMiddleware catches panics in downstream handlers and converts them to errors.
func RecoveryMiddleware() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, update *IncomingUpdate) (err error) {
			defer func() {
				if r := recover(); r != nil {
					nxlog.Error("panic recovered in handler",
						zap.Any("panic", r),
						zap.Int32("user_id", update.UserID),
						zap.Int64("conversation_id", update.ConversationID),
					)
					if e, ok := r.(error); ok {
						err = e
					}
				}
			}()
			return next(ctx, update)
		}
	}
}
