package agentic

import (
	"context"
	"strings"
)

// Router dispatches incoming updates to slash commands, card action handlers,
// or the LLM fallback. It does NOT contain security/rate-limit logic — those
// belong in middleware.
type Router struct {
	commands    map[string]Command
	cardActions map[string]Handler // keyed by verb
	llmHandler  Handler
	cardDefault Handler // fallback for unmatched card actions
}

// NewRouter creates a Router with the given LLM fallback handler and commands.
// If llmHandler is nil, the Engine will auto-bind its RunLLM method during NewEngine.
func NewRouter(llmHandler Handler, cmds ...Command) *Router {
	m := make(map[string]Command, len(cmds))
	for _, c := range cmds {
		m[strings.ToLower(c.Name)] = c
	}
	return &Router{
		commands:    m,
		cardActions: make(map[string]Handler),
		llmHandler:  llmHandler,
	}
}

// OnCardAction registers a handler for a specific card action verb.
// When a CardAction with this verb arrives, the handler is called instead of LLM.
func (r *Router) OnCardAction(verb string, handler Handler) *Router {
	r.cardActions[verb] = handler
	return r
}

// OnCardActionDefault sets a fallback handler for card actions with unregistered verbs.
// If not set, unmatched card actions go to the LLM handler.
func (r *Router) OnCardActionDefault(handler Handler) *Router {
	r.cardDefault = handler
	return r
}

// Handle routes an update to the matching handler.
//
// Routing priority:
//  1. CardAction → verb-based handler → card default → LLM fallback
//  2. Slash command → registered command handler
//  3. Everything else → LLM fallback
func (r *Router) Handle(ctx context.Context, update *IncomingUpdate) error {
	// Card action routing.
	if update.CardAction != nil {
		verb := update.CardAction.Verb
		if h, ok := r.cardActions[verb]; ok {
			return h(ctx, update)
		}
		if r.cardDefault != nil {
			return r.cardDefault(ctx, update)
		}
		return r.llmHandler(ctx, update)
	}

	// Slash command routing.
	if name, ok := parseCommand(update.Text); ok {
		if cmd, found := r.commands[name]; found {
			return cmd.Handler(ctx, update)
		}
		// Unknown command — fall through to LLM.
	}

	return r.llmHandler(ctx, update)
}

// Commands returns all registered commands (for /help generation).
func (r *Router) Commands() []Command {
	cmds := make([]Command, 0, len(r.commands))
	for _, c := range r.commands {
		cmds = append(cmds, c)
	}
	return cmds
}

// parseCommand extracts the command name from a message like "/help arg1".
func parseCommand(text string) (string, bool) {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, "/") {
		return "", false
	}
	cmd := strings.TrimPrefix(text, "/")
	if i := strings.IndexByte(cmd, ' '); i > 0 {
		cmd = cmd[:i]
	}
	cmd = strings.ToLower(cmd)
	if cmd == "" {
		return "", false
	}
	return cmd, true
}
