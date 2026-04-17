package agentic

import (
	"context"

	"github.com/pinealctx/nexus-x/nxlog"
	"go.uber.org/zap"
)

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

// LoggingEventHandler logs all lifecycle events at Info level.
// Use it as the default EventHandler for production observability.
type LoggingEventHandler struct{}

func (LoggingEventHandler) OnLLMStart(ctx context.Context, userID int32, input string) {
	nxlog.Info("llm start",
		zap.Int32("user_id", userID),
		zap.Int64("conversation_id", ConversationIDFromContext(ctx)),
		zap.String("input", input),
	)
}

func (LoggingEventHandler) OnLLMEnd(ctx context.Context, userID int32, response string, tokens TokenUsage) {
	nxlog.Info("llm end",
		zap.Int32("user_id", userID),
		zap.Int64("conversation_id", ConversationIDFromContext(ctx)),
		zap.String("response", response),
		zap.Int("input_tokens", tokens.InputTokens),
		zap.Int("output_tokens", tokens.OutputTokens),
	)
}

func (LoggingEventHandler) OnToolCall(ctx context.Context, toolName string, input string) {
	nxlog.Info("llm tool call",
		zap.Int64("conversation_id", ConversationIDFromContext(ctx)),
		zap.String("tool", toolName),
		zap.String("input", input),
	)
}

func (LoggingEventHandler) OnToolResult(ctx context.Context, toolName string, output string, err error) {
	if err != nil {
		nxlog.Warn("llm tool error",
			zap.Int64("conversation_id", ConversationIDFromContext(ctx)),
			zap.String("tool", toolName),
			zap.Error(err),
		)
		return
	}
	nxlog.Info("llm tool result",
		zap.Int64("conversation_id", ConversationIDFromContext(ctx)),
		zap.String("tool", toolName),
		zap.String("output", output),
	)
}

func (LoggingEventHandler) OnError(ctx context.Context, err error) {
	nxlog.Error("llm error",
		zap.Int64("conversation_id", ConversationIDFromContext(ctx)),
		zap.Error(err),
	)
}
