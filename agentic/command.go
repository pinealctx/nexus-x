package agentic

// Command represents a slash command that the Router can dispatch.
type Command struct {
	// Name is the command trigger (without the leading slash), e.g. "help".
	Name string

	// Description is a short help text shown in /help listings.
	Description string

	// Handler processes the command. The update's Text contains the full
	// message including the command prefix (e.g. "/help arg1 arg2").
	Handler Handler
}
