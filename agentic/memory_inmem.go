package agentic

import (
	"context"
	"fmt"
	"sync"
)

// InMemoryMemory is a simple in-memory Memory implementation for development and testing.
// It is not persistent and not safe for production use with multiple processes.
type InMemoryMemory struct {
	mu       sync.RWMutex
	store    map[string][]Message
	maxMsgs  int
	summarizer Summarizer
}

// InMemoryMemoryOption configures InMemoryMemory.
type InMemoryMemoryOption func(*InMemoryMemory)

// WithMaxMessages sets the maximum number of messages to retain per key.
// When exceeded, older messages are truncated (or summarized if a Summarizer is set).
func WithMaxMessages(n int) InMemoryMemoryOption {
	return func(m *InMemoryMemory) { m.maxMsgs = n }
}

// WithSummarizer sets a Summarizer for automatic history compression.
func WithSummarizer(s Summarizer) InMemoryMemoryOption {
	return func(m *InMemoryMemory) { m.summarizer = s }
}

// NewInMemoryMemory creates a new in-memory Memory.
func NewInMemoryMemory(opts ...InMemoryMemoryOption) *InMemoryMemory {
	m := &InMemoryMemory{
		store:   make(map[string][]Message),
		maxMsgs: 50,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func memoryKeyString(key MemoryKey) string {
	return fmt.Sprintf("%s:%d:%d", key.AgentID, key.UserID, key.ConversationID)
}

// Load retrieves stored messages.
func (m *InMemoryMemory) Load(_ context.Context, key MemoryKey) ([]Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	msgs := m.store[memoryKeyString(key)]
	if msgs == nil {
		return nil, nil
	}
	out := make([]Message, len(msgs))
	copy(out, msgs)
	return out, nil
}

// Save persists messages, applying truncation or summarization if needed.
func (m *InMemoryMemory) Save(ctx context.Context, key MemoryKey, msgs []Message) error {
	if m.maxMsgs > 0 && len(msgs) > m.maxMsgs {
		if m.summarizer != nil {
			summarized, err := m.summarizer.Summarize(ctx, msgs)
			if err == nil {
				msgs = summarized
			}
		}
		// If still over limit after summarization (or no summarizer), truncate.
		if len(msgs) > m.maxMsgs {
			msgs = msgs[len(msgs)-m.maxMsgs:]
		}
	}

	stored := make([]Message, len(msgs))
	copy(stored, msgs)

	m.mu.Lock()
	m.store[memoryKeyString(key)] = stored
	m.mu.Unlock()
	return nil
}

// Clear removes all stored messages for the given key.
func (m *InMemoryMemory) Clear(_ context.Context, key MemoryKey) error {
	m.mu.Lock()
	delete(m.store, memoryKeyString(key))
	m.mu.Unlock()
	return nil
}
