package agentic

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"charm.land/fantasy"
)

// Engine is the central coordinator of the conversational agent framework.
// It wires together the Router, middleware chain, Memory, and Fantasy Agent.
type Engine struct {
	agentID  string
	agent    fantasy.Agent
	router   *Router
	memory   Memory
	events   EventHandler
	mws      []Middleware
	handler  Handler // compiled handler (middleware + router)
	systemFn func(ctx context.Context, update *IncomingUpdate) string
	errorFn  func(ctx context.Context, ch Channel, convID int64, err error) error

	// Fantasy defaults — applied to every AgentCall unless overridden per-request.
	temperature *float64
	maxTokens   *int64
	stopWhen    []fantasy.StopCondition
	prepareStep fantasy.PrepareStepFunction
}

// EngineOption configures an Engine.
type EngineOption func(*Engine)

// WithAgentID sets the agent identifier used for memory key namespacing.
func WithAgentID(id string) EngineOption {
	return func(e *Engine) { e.agentID = id }
}

// WithAgent sets the Fantasy agent. Required.
func WithAgent(agent fantasy.Agent) EngineOption {
	return func(e *Engine) { e.agent = agent }
}

// WithRouter sets the command router. Required.
func WithRouter(router *Router) EngineOption {
	return func(e *Engine) { e.router = router }
}

// WithMemory sets the conversation memory store.
func WithMemory(m Memory) EngineOption {
	return func(e *Engine) { e.memory = m }
}

// WithEventHandler sets the lifecycle event handler.
func WithEventHandler(h EventHandler) EngineOption {
	return func(e *Engine) { e.events = h }
}

// WithMiddleware appends middleware to the chain.
func WithMiddleware(mws ...Middleware) EngineOption {
	return func(e *Engine) { e.mws = append(e.mws, mws...) }
}

// WithSystemPrompt sets a static system prompt for LLM calls.
func WithSystemPrompt(prompt string) EngineOption {
	return func(e *Engine) {
		e.systemFn = func(context.Context, *IncomingUpdate) string { return prompt }
	}
}

// WithDynamicSystemPrompt sets a function that generates the system prompt per request.
func WithDynamicSystemPrompt(fn func(ctx context.Context, update *IncomingUpdate) string) EngineOption {
	return func(e *Engine) { e.systemFn = fn }
}

// WithErrorHandler sets a custom error handler for LLM failures.
// If nil, the default behavior sends the error message to the user.
func WithErrorHandler(fn func(ctx context.Context, ch Channel, convID int64, err error) error) EngineOption {
	return func(e *Engine) { e.errorFn = fn }
}

// WithTemperature sets the default temperature for LLM calls.
func WithTemperature(t float64) EngineOption {
	return func(e *Engine) { e.temperature = &t }
}

// WithMaxTokens sets the default max output tokens for LLM calls.
func WithMaxTokens(n int64) EngineOption {
	return func(e *Engine) { e.maxTokens = &n }
}

// WithStopConditions sets the default stop conditions for LLM calls.
func WithStopConditions(conditions ...fantasy.StopCondition) EngineOption {
	return func(e *Engine) { e.stopWhen = conditions }
}

// WithPrepareStep sets the default prepare step function for LLM calls.
func WithPrepareStep(fn fantasy.PrepareStepFunction) EngineOption {
	return func(e *Engine) { e.prepareStep = fn }
}

// NewEngine creates and compiles an Engine with the given options.
func NewEngine(opts ...EngineOption) (*Engine, error) {
	e := &Engine{
		events: NoopEventHandler{},
	}
	for _, opt := range opts {
		opt(e)
	}
	if e.agent == nil {
		return nil, fmt.Errorf("agentic: WithAgent is required")
	}
	if e.router == nil {
		return nil, fmt.Errorf("agentic: WithRouter is required")
	}

	// Compile the middleware chain around the router.
	// If the router has no LLM handler, bind Engine.RunLLM as the default.
	if e.router.llmHandler == nil {
		e.router.llmHandler = func(ctx context.Context, update *IncomingUpdate) error {
			return e.RunLLM(ctx, update)
		}
	}
	var handler Handler = e.router.Handle
	if len(e.mws) > 0 {
		handler = Chain(e.mws...)(handler)
	}
	e.handler = handler

	return e, nil
}

// Agent returns the underlying Fantasy agent for direct access.
func (e *Engine) Agent() fantasy.Agent { return e.agent }

// Memory returns the configured Memory, or nil if not set.
func (e *Engine) Memory() Memory { return e.memory }

// Close performs graceful shutdown. Currently a no-op placeholder
// for future resource cleanup (e.g., flushing memory, closing connections).
func (e *Engine) Close() error { return nil }

// Handle processes an incoming update through the full middleware + routing pipeline.
func (e *Engine) Handle(ctx context.Context, update *IncomingUpdate) error {
	ctx = WithUserID(ctx, update.UserID)
	ctx = WithConversationID(ctx, update.ConversationID)
	ctx = WithChannel(ctx, update.Channel)
	if e.memory != nil {
		ctx = ContextWithMemory(ctx, e.memory)
	}
	return e.handler(ctx, update)
}

// --- LLM Pipeline ---

// LLMResult holds the outcome of an LLM call, exposing Fantasy's full result.
type LLMResult struct {
	// Text is the final text response from the LLM.
	Text string
	// Result is the full Fantasy AgentResult (steps, usage, warnings, provider metadata).
	Result *fantasy.AgentResult
}

// LLMOption configures a single LLM call, overriding Engine defaults.
type LLMOption func(*llmConfig)

type llmConfig struct {
	temperature *float64
	maxTokens   *int64
	extraTools  []fantasy.AgentTool   // additional tools for this call
	stopWhen    []fantasy.StopCondition
	prepareStep fantasy.PrepareStepFunction
	activeTools []string              // filter which tools are active
	extraMsgs   []fantasy.Message     // prepended before user message
}

// LLMWithTemperature overrides temperature for this call.
func LLMWithTemperature(t float64) LLMOption {
	return func(c *llmConfig) { c.temperature = &t }
}

// LLMWithMaxTokens overrides max output tokens for this call.
func LLMWithMaxTokens(n int64) LLMOption {
	return func(c *llmConfig) { c.maxTokens = &n }
}

// LLMWithTools adds extra tools for this specific call (appended to agent's tools).
func LLMWithTools(tools ...fantasy.AgentTool) LLMOption {
	return func(c *llmConfig) { c.extraTools = tools }
}

// LLMWithStopConditions overrides stop conditions for this call.
func LLMWithStopConditions(conditions ...fantasy.StopCondition) LLMOption {
	return func(c *llmConfig) { c.stopWhen = conditions }
}

// LLMWithPrepareStep overrides the prepare step function for this call.
func LLMWithPrepareStep(fn fantasy.PrepareStepFunction) LLMOption {
	return func(c *llmConfig) { c.prepareStep = fn }
}

// LLMWithExtraMessages prepends additional messages before the user message.
func LLMWithExtraMessages(msgs ...fantasy.Message) LLMOption {
	return func(c *llmConfig) { c.extraMsgs = msgs }
}

// LLMWithActiveTools filters which tools are active for this call (by name).
func LLMWithActiveTools(names ...string) LLMOption {
	return func(c *llmConfig) { c.activeTools = names }
}

// RunLLM executes the Fantasy agent with conversation memory and streaming support.
// This is the default LLM handler — pass it to NewRouter as the fallback.
//
// Pipeline: load memory → build prompt → call Fantasy → save memory → send reply.
// Each step respects Engine defaults, which can be overridden per-call via LLMOption.
func (e *Engine) RunLLM(ctx context.Context, update *IncomingUpdate, opts ...LLMOption) error {
	result, err := e.CallLLM(ctx, update, opts...)
	if err != nil {
		return err
	}
	_ = result // already sent via channel in CallLLM
	return nil
}

// RunLLMHandler returns a Handler that calls RunLLM with the given options.
// Use this to pass to NewRouter as the LLM fallback.
func (e *Engine) RunLLMHandler(opts ...LLMOption) Handler {
	return func(ctx context.Context, update *IncomingUpdate) error {
		return e.RunLLM(ctx, update, opts...)
	}
}

// CallLLM executes the LLM pipeline and returns the full result.
// Unlike RunLLM, the caller gets access to the LLMResult for inspection.
func (e *Engine) CallLLM(ctx context.Context, update *IncomingUpdate, opts ...LLMOption) (*LLMResult, error) {
	cfg := e.buildLLMConfig(opts)
	ch := update.Channel
	convID := update.ConversationID

	// 1. Load memory.
	memKey := MemoryKey{
		AgentID:        e.agentID,
		UserID:         update.UserID,
		ConversationID: convID,
	}
	history := e.loadMemory(ctx, memKey)

	// 2. Build prompt.
	var systemPrompt string
	if e.systemFn != nil {
		systemPrompt = e.systemFn(ctx, update)
	}

	// Prepend extra messages if any.
	if len(cfg.extraMsgs) > 0 {
		history = append(history, cfg.extraMsgs...)
	}

	// Append current user message.
	history = append(history, fantasy.NewUserMessage(update.Text))

	// 3. Fire event.
	e.events.OnLLMStart(ctx, update.UserID, update.Text)

	// 4. Call Fantasy (streaming or blocking).
	var result *LLMResult
	var err error
	if streamer, ok := ch.(StreamingChannel); ok {
		result, err = e.callLLMStreaming(ctx, streamer, convID, systemPrompt, history, cfg)
	} else {
		result, err = e.callLLMBlocking(ctx, ch, convID, systemPrompt, history, cfg)
	}
	if err != nil {
		e.events.OnError(ctx, err)
		return nil, e.handleLLMError(ctx, ch, convID, err)
	}

	// 5. Fire end event.
	usage := TokenUsage{
		InputTokens:  int(result.Result.TotalUsage.InputTokens),
		OutputTokens: int(result.Result.TotalUsage.OutputTokens),
	}
	e.events.OnLLMEnd(ctx, update.UserID, result.Text, usage)

	// 6. Save memory (includes tool call steps).
	e.saveMemory(ctx, memKey, history, result.Result)

	return result, nil
}

func (e *Engine) buildLLMConfig(opts []LLMOption) *llmConfig {
	cfg := &llmConfig{
		temperature: e.temperature,
		maxTokens:   e.maxTokens,
		stopWhen:    e.stopWhen,
		prepareStep: e.prepareStep,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

func (e *Engine) loadMemory(ctx context.Context, key MemoryKey) []fantasy.Message {
	if e.memory == nil {
		return nil
	}
	msgs, err := e.memory.Load(ctx, key)
	if err != nil {
		slog.Warn("failed to load memory", "err", err, "user_id", key.UserID)
		return nil
	}
	return toFantasyMessages(msgs)
}

func (e *Engine) saveMemory(ctx context.Context, key MemoryKey, history []fantasy.Message, result *fantasy.AgentResult) {
	if e.memory == nil {
		return
	}
	// Convert history (user messages) + agent result steps into Memory messages.
	updated := fromFantasyMessages(history)

	// Append all step messages from the agent result to preserve tool call history.
	for _, step := range result.Steps {
		for _, msg := range step.Messages {
			updated = append(updated, fromFantasyMessage(msg))
		}
	}

	if err := e.memory.Save(ctx, key, updated); err != nil {
		slog.Warn("failed to save memory", "err", err, "user_id", key.UserID)
	}
}

func (e *Engine) handleLLMError(ctx context.Context, ch Channel, convID int64, err error) error {
	if e.errorFn != nil {
		return e.errorFn(ctx, ch, convID, err)
	}
	return ch.SendText(ctx, convID, fmt.Sprintf("Sorry, I encountered an error: %s", err.Error()))
}

func (e *Engine) buildAgentCall(systemPrompt string, history []fantasy.Message, cfg *llmConfig) fantasy.AgentCall {
	return fantasy.AgentCall{
		Prompt:          systemPrompt,
		Messages:        history,
		Temperature:     cfg.temperature,
		MaxOutputTokens: cfg.maxTokens,
		StopWhen:        cfg.stopWhen,
		PrepareStep:     cfg.prepareStep,
		ActiveTools:     cfg.activeTools,
	}
}

func (e *Engine) callLLMBlocking(
	ctx context.Context,
	ch Channel,
	convID int64,
	systemPrompt string,
	history []fantasy.Message,
	cfg *llmConfig,
) (*LLMResult, error) {
	call := e.buildAgentCall(systemPrompt, history, cfg)
	result, err := e.agent.Generate(ctx, call)
	if err != nil {
		return nil, err
	}

	responseText := result.Response.Content.Text()
	if err := ch.SendText(ctx, convID, responseText); err != nil {
		return nil, fmt.Errorf("send response: %w", err)
	}

	return &LLMResult{Text: responseText, Result: result}, nil
}

func (e *Engine) callLLMStreaming(
	ctx context.Context,
	ch StreamingChannel,
	convID int64,
	systemPrompt string,
	history []fantasy.Message,
	cfg *llmConfig,
) (*LLMResult, error) {
	sw, err := ch.StartStream(ctx, convID)
	if err != nil {
		// Fall back to blocking.
		return e.callLLMBlocking(ctx, ch, convID, systemPrompt, history, cfg)
	}

	var fullText string
	streamCall := fantasy.AgentStreamCall{
		Prompt:          systemPrompt,
		Messages:        history,
		Temperature:     cfg.temperature,
		MaxOutputTokens: cfg.maxTokens,
		StopWhen:        cfg.stopWhen,
		PrepareStep:     cfg.prepareStep,
		OnTextDelta: func(_ string, delta string) error {
			fullText += delta
			return sw.Push(ctx, delta)
		},
		OnToolCall: func(tc fantasy.ToolCallContent) error {
			e.events.OnToolCall(ctx, tc.ToolName, tc.Input)
			return nil
		},
		OnToolResult: func(tr fantasy.ToolResultContent) error {
			e.events.OnToolResult(ctx, tr.ToolName, fmt.Sprintf("%v", tr.Result), nil)
			return nil
		},
	}

	result, err := e.agent.Stream(ctx, streamCall)
	if err != nil {
		_ = sw.Error(ctx, fmt.Sprintf("Sorry, I encountered an error: %s", err.Error()))
		return nil, err
	}

	if err := sw.End(ctx, fullText); err != nil {
		return nil, fmt.Errorf("end stream: %w", err)
	}

	return &LLMResult{Text: fullText, Result: result}, nil
}

// --- Message conversion ---

func toFantasyMessages(msgs []Message) []fantasy.Message {
	out := make([]fantasy.Message, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, toFantasyMessage(m))
	}
	return out
}

func toFantasyMessage(m Message) fantasy.Message {
	role := fantasy.MessageRole(m.Role)

	// If Parts are populated, convert them to Fantasy MessageParts.
	if len(m.Parts) > 0 {
		parts := make([]fantasy.MessagePart, 0, len(m.Parts))
		for _, p := range m.Parts {
			switch p.Type {
			case "text":
				parts = append(parts, fantasy.TextPart{Text: p.Text})
			case "tool_call":
				parts = append(parts, fantasy.ToolCallPart{
					ToolCallID: p.ToolCallID,
					ToolName:   p.ToolName,
					Input:      p.Input,
				})
			case "tool_result":
				parts = append(parts, fantasy.ToolResultPart{
					ToolCallID: p.ToolCallID,
					Output:     toolResultOutput(p),
				})
			}
		}
		return fantasy.Message{Role: role, Content: parts}
	}

	// Fallback: simple text message.
	return fantasy.Message{
		Role: role,
		Content: []fantasy.MessagePart{
			fantasy.TextPart{Text: m.Content},
		},
	}
}

func toolResultOutput(p MessagePart) fantasy.ToolResultOutputContent {
	if p.IsError {
		return fantasy.ToolResultOutputContentError{Error: fmt.Errorf("%s", p.Output)}
	}
	return fantasy.ToolResultOutputContentText{Text: p.Output}
}

func fromFantasyMessages(msgs []fantasy.Message) []Message {
	out := make([]Message, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, fromFantasyMessage(m))
	}
	return out
}

func fromFantasyMessage(m fantasy.Message) Message {
	msg := Message{Role: string(m.Role)}

	var textBuilder strings.Builder
	var parts []MessagePart

	for _, part := range m.Content {
		switch p := part.(type) {
		case fantasy.TextPart:
			textBuilder.WriteString(p.Text)
			parts = append(parts, MessagePart{Type: "text", Text: p.Text})
		case fantasy.ToolCallPart:
			parts = append(parts, MessagePart{
				Type:       "tool_call",
				ToolCallID: p.ToolCallID,
				ToolName:   p.ToolName,
				Input:      p.Input,
			})
		case fantasy.ToolResultPart:
			mp := MessagePart{
				Type:       "tool_result",
				ToolCallID: p.ToolCallID,
			}
			if p.Output != nil {
				switch o := p.Output.(type) {
				case fantasy.ToolResultOutputContentText:
					mp.Output = o.Text
				case fantasy.ToolResultOutputContentError:
					mp.Output = o.Error.Error()
					mp.IsError = true
				}
			}
			parts = append(parts, mp)
		default:
			// AsMessagePart fallback for TextPart via interface
			if tp, ok := fantasy.AsMessagePart[fantasy.TextPart](part); ok {
				textBuilder.WriteString(tp.Text)
				parts = append(parts, MessagePart{Type: "text", Text: tp.Text})
			}
		}
	}

	msg.Content = textBuilder.String()

	// Only store Parts if there's more than just text.
	hasNonText := false
	for _, p := range parts {
		if p.Type != "text" {
			hasNonText = true
			break
		}
	}
	if hasNonText {
		msg.Parts = parts
	}

	return msg
}
