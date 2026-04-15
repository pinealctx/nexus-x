package agentic

import "context"

// RateLimiter controls request throughput per user.
type RateLimiter interface {
	// Allow returns true if the user is within rate limits.
	Allow(ctx context.Context, userID int32) (bool, error)
}

// Deduplicator detects and filters duplicate messages.
type Deduplicator interface {
	// IsDuplicate returns true if this message has been seen before.
	IsDuplicate(conversationID int64, messageID int64) bool
}

// CredentialChecker verifies whether a user has valid credentials for the agent.
type CredentialChecker interface {
	// Check returns true if the user has valid credentials.
	Check(ctx context.Context, userID int32) (bool, error)
}
