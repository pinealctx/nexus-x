package agentic

import "context"

// EventHandler receives lifecycle events from the Engine.
// All methods are optional — embed NoopEventHandler for a partial implementation.
type EventHandler interface {
	OnLLMStart(ctx context.Context, userID int32, input string)
	OnLLMEnd(ctx context.Context, userID int32, response string, tokens TokenUsage)
	OnToolCall(ctx context.Context, toolName string, input string)
	OnToolResult(ctx context.Context, toolName string, output string, err error)
	OnError(ctx context.Context, err error)
}

// TokenUsage tracks LLM token consumption for a single request.
type TokenUsage struct {
	InputTokens  int
	OutputTokens int
}

// NoopEventHandler is a no-op implementation of EventHandler.
// Embed it to implement only the methods you care about.
type NoopEventHandler struct{}

func (NoopEventHandler) OnLLMStart(context.Context, int32, string)           {}
func (NoopEventHandler) OnLLMEnd(context.Context, int32, string, TokenUsage) {}
func (NoopEventHandler) OnToolCall(context.Context, string, string)          {}
func (NoopEventHandler) OnToolResult(context.Context, string, string, error) {}
func (NoopEventHandler) OnError(context.Context, error)                      {}
