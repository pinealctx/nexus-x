package agentic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisMemory is a Redis-backed Memory implementation.
type RedisMemory struct {
	rdb        redis.UniversalClient
	prefix     string
	maxMsgs    int
	ttl        time.Duration
	summarizer Summarizer
}

// RedisMemoryOption configures RedisMemory.
type RedisMemoryOption func(*RedisMemory)

// WithRedisPrefix sets the key prefix for Redis keys.
func WithRedisPrefix(prefix string) RedisMemoryOption {
	return func(m *RedisMemory) { m.prefix = prefix }
}

// WithRedisMaxMessages sets the maximum number of messages to retain.
func WithRedisMaxMessages(n int) RedisMemoryOption {
	return func(m *RedisMemory) { m.maxMsgs = n }
}

// WithRedisTTL sets the TTL for memory keys.
func WithRedisTTL(ttl time.Duration) RedisMemoryOption {
	return func(m *RedisMemory) { m.ttl = ttl }
}

// WithRedisSummarizer sets a Summarizer for automatic history compression.
func WithRedisSummarizer(s Summarizer) RedisMemoryOption {
	return func(m *RedisMemory) { m.summarizer = s }
}

// NewRedisMemory creates a new Redis-backed Memory.
func NewRedisMemory(rdb redis.UniversalClient, opts ...RedisMemoryOption) *RedisMemory {
	m := &RedisMemory{
		rdb:     rdb,
		prefix:  "agentic:memory",
		maxMsgs: 50,
		ttl:     24 * time.Hour,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func (m *RedisMemory) redisKey(key MemoryKey) string {
	return fmt.Sprintf("%s:%s:%d:%d", m.prefix, key.AgentID, key.UserID, key.ConversationID)
}

// Load retrieves stored messages from Redis.
func (m *RedisMemory) Load(ctx context.Context, key MemoryKey) ([]Message, error) {
	data, err := m.rdb.Get(ctx, m.redisKey(key)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("redis memory load: %w", err)
	}
	var msgs []Message
	if err := json.Unmarshal(data, &msgs); err != nil {
		return nil, fmt.Errorf("redis memory unmarshal: %w", err)
	}
	return msgs, nil
}

// Save persists messages to Redis.
func (m *RedisMemory) Save(ctx context.Context, key MemoryKey, msgs []Message) error {
	if m.maxMsgs > 0 && len(msgs) > m.maxMsgs {
		if m.summarizer != nil {
			summarized, err := m.summarizer.Summarize(ctx, msgs)
			if err == nil {
				msgs = summarized
			}
		}
		if len(msgs) > m.maxMsgs {
			msgs = msgs[len(msgs)-m.maxMsgs:]
		}
	}

	data, err := json.Marshal(msgs)
	if err != nil {
		return fmt.Errorf("redis memory marshal: %w", err)
	}
	return m.rdb.Set(ctx, m.redisKey(key), data, m.ttl).Err()
}

// Clear removes all stored messages from Redis.
func (m *RedisMemory) Clear(ctx context.Context, key MemoryKey) error {
	return m.rdb.Del(ctx, m.redisKey(key)).Err()
}
