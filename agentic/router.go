package agentic

import (
	"context"
	"strings"
)

// Router dispatches incoming updates to slash commands, keyword rules,
// card action handlers, or the LLM fallback. It does NOT contain
// security/rate-limit logic — those belong in middleware.
//
// Routing priority:
//  1. CardAction → verb-based handler → card default → LLM fallback
//  2. Slash command → registered command handler
//  3. Keyword rule → first matching rule wins (registration order)
//  4. Everything else → LLM fallback
type Router struct {
	commands    map[string]Command
	keywords    []KeywordRule
	cardActions map[string]Handler // keyed by verb
	llmHandler  Handler
	cardDefault Handler // fallback for unmatched card actions
}

// KeywordRule defines a keyword-based auto-reply that bypasses LLM.
// Rules are evaluated in registration order; the first match wins.
type KeywordRule struct {
	// Match returns true if this rule should handle the given text.
	Match func(text string) bool

	// Handler processes the matched update.
	Handler Handler
}

// RouterOption configures a Router.
type RouterOption func(*Router)

// WithCommands registers slash commands.
func WithCommands(cmds ...Command) RouterOption {
	return func(r *Router) {
		for _, c := range cmds {
			r.commands[strings.ToLower(c.Name)] = c
		}
	}
}

// WithLLMHandler sets a custom LLM fallback handler.
// If not set, Engine auto-binds its RunLLM method during NewEngine.
func WithLLMHandler(h Handler) RouterOption {
	return func(r *Router) { r.llmHandler = h }
}

// WithCardAction registers a handler for a specific card action verb.
func WithCardAction(verb string, handler Handler) RouterOption {
	return func(r *Router) { r.cardActions[verb] = handler }
}

// WithCardActionDefault sets a fallback handler for unmatched card action verbs.
func WithCardActionDefault(handler Handler) RouterOption {
	return func(r *Router) { r.cardDefault = handler }
}

// WithKeyword registers a keyword rule with a custom matcher.
// Rules are evaluated in registration order; the first match wins.
func WithKeyword(match func(string) bool, handler Handler) RouterOption {
	return func(r *Router) {
		r.keywords = append(r.keywords, KeywordRule{Match: match, Handler: handler})
	}
}

// WithKeywordExact registers an exact-match keyword rule (case-insensitive, trimmed).
func WithKeywordExact(keyword string, handler Handler) RouterOption {
	lower := strings.ToLower(strings.TrimSpace(keyword))
	return WithKeyword(func(text string) bool {
		return strings.ToLower(strings.TrimSpace(text)) == lower
	}, handler)
}

// WithKeywordContains registers a contains-match keyword rule (case-insensitive).
func WithKeywordContains(keyword string, handler Handler) RouterOption {
	lower := strings.ToLower(keyword)
	return WithKeyword(func(text string) bool {
		return strings.Contains(strings.ToLower(text), lower)
	}, handler)
}

// WithKeywordPrefix registers a prefix-match keyword rule (case-insensitive, trimmed).
func WithKeywordPrefix(prefix string, handler Handler) RouterOption {
	lower := strings.ToLower(strings.TrimSpace(prefix))
	return WithKeyword(func(text string) bool {
		return strings.HasPrefix(strings.ToLower(strings.TrimSpace(text)), lower)
	}, handler)
}

// NewRouter creates a Router with the given options.
//
//	router := agentic.NewRouter(
//	    agentic.WithCommands(helpCmd, pingCmd),
//	    agentic.WithKeywordExact("你好", greetHandler),
//	    agentic.WithKeywordContains("价格", priceHandler),
//	)
func NewRouter(opts ...RouterOption) *Router {
	r := &Router{
		commands:    make(map[string]Command),
		cardActions: make(map[string]Handler),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// OnCardAction registers a handler for a specific card action verb.
// Returns the Router for chaining. Prefer WithCardAction for construction-time setup.
func (r *Router) OnCardAction(verb string, handler Handler) *Router {
	r.cardActions[verb] = handler
	return r
}

// OnCardActionDefault sets a fallback handler for card actions with unregistered verbs.
// Returns the Router for chaining. Prefer WithCardActionDefault for construction-time setup.
func (r *Router) OnCardActionDefault(handler Handler) *Router {
	r.cardDefault = handler
	return r
}

// Handle routes an update to the matching handler.
//
// Routing priority:
//  1. CardAction → verb-based handler → card default → LLM fallback
//  2. Slash command → registered command handler
//  3. Keyword rule → first matching rule (registration order)
//  4. Everything else → LLM fallback
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
		// Unknown command — fall through to keyword / LLM.
	}

	// Keyword routing.
	for i := range r.keywords {
		if r.keywords[i].Match(update.Text) {
			return r.keywords[i].Handler(ctx, update)
		}
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

// Keywords returns all registered keyword rules.
func (r *Router) Keywords() []KeywordRule {
	out := make([]KeywordRule, len(r.keywords))
	copy(out, r.keywords)
	return out
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
