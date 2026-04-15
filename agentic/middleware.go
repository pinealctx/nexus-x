package agentic

import (
	"context"
	"log/slog"
)

// DedupMiddleware filters duplicate messages using the provided Deduplicator.
func DedupMiddleware(dedup Deduplicator) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, update *IncomingUpdate) error {
			if update.MessageID != 0 && dedup.IsDuplicate(update.ConversationID, update.MessageID) {
				slog.Debug("dedup: skipping duplicate message",
					"conversation_id", update.ConversationID,
					"message_id", update.MessageID,
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
				slog.Warn("rate limiter error, allowing request", "err", err, "user_id", update.UserID)
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
				slog.Warn("credential check error", "err", err, "user_id", update.UserID)
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
			slog.Debug("incoming update",
				"user_id", update.UserID,
				"conversation_id", update.ConversationID,
				"message_id", update.MessageID,
				"text_len", len(update.Text),
			)
			err := next(ctx, update)
			if err != nil {
				slog.Error("handler error",
					"user_id", update.UserID,
					"conversation_id", update.ConversationID,
					"err", err,
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
					slog.Error("panic recovered in handler",
						"panic", r,
						"user_id", update.UserID,
						"conversation_id", update.ConversationID,
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
