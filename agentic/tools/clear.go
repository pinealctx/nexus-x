package tools

import (
	"context"

	"charm.land/fantasy"

	"github.com/pinealctx/nexus-x/agentic"
)

// ClearContext creates a tool that clears the conversation memory for the
// current user and conversation. After clearing, the next LLM turn starts
// with a fresh context (no previous conversation history).
//
// The tool requires memory to be configured on the Engine (via WithMemory).
// If memory is not configured, the tool returns an error message to the LLM.
func ClearContext() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"clear_context",
		"Clear all conversation history for the current user. "+
			"Use this when the user explicitly asks to start fresh, reset, "+
			"or forget the conversation. After clearing, the next response "+
			"will have no memory of previous messages.",
		func(ctx context.Context, _ struct{}, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
			mem := agentic.MemoryFromContext(ctx)
			if mem == nil {
				return fantasy.NewTextErrorResponse("memory is not configured, cannot clear context"), nil
			}

			convID := agentic.ConversationIDFromContext(ctx)
			if convID == 0 {
				return fantasy.NewTextErrorResponse("no conversation context"), nil
			}

			key := agentic.MemoryKey{
				AgentID:        agentic.AgentIDFromContext(ctx),
				UserID:         agentic.UserIDFromContext(ctx),
				ConversationID: convID,
			}

			if err := mem.Clear(ctx, key); err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}

			// Signal saveMemory to discard old history for this turn.
			agentic.MarkContextCleared(ctx)

			return fantasy.NewTextResponse("Conversation context cleared successfully."), nil
		},
	)
}
