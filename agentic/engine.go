package agentic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/fantasy"
	"github.com/pinealctx/nexus-x/nxlog"
	"go.uber.org/zap"
)

// Engine is the central coordinator of the conversational agent framework.
// It wires together the Router, middleware chain, Memory, and Fantasy Agent.
//
// Engine only orchestrates — it does NOT send messages. LLM sends messages
// through tools (e.g. send_text, send_card). Engine manages the pipeline:
// receive update → middleware → route → LLM call → save memory.
type Engine struct {
	agentID  string
	agent    fantasy.Agent
	agents   map[string]fantasy.Agent // named agents for multi-model routing
	router   *Router
	memory   Memory
	events   EventHandler
	mws      []Middleware
	handler  Handler // compiled handler (middleware + router)
	systemFn func(ctx context.Context, update *IncomingUpdate) string
	errorFn  func(ctx context.Context, update *IncomingUpdate, err error) error

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

// WithAgent sets the default Fantasy agent. Required.
func WithAgent(agent fantasy.Agent) EngineOption {
	return func(e *Engine) { e.agent = agent }
}

// NamedAgent pairs a name with a Fantasy Agent for multi-model routing.
type NamedAgent struct {
	Name  string
	Agent fantasy.Agent
}

// WithAgents registers named agents for multi-model routing.
// Use LLMWithAgent("name") in RunLLM/CallLLM to select a specific agent.
//
//	engine, _ := agentic.NewEngine(
//	    agentic.WithAgent(sonnetAgent),
//	    agentic.WithAgents(
//	        agentic.NamedAgent{Name: "fast", Agent: haikuAgent},
//	        agentic.NamedAgent{Name: "smart", Agent: opusAgent},
//	    ),
//	)
//
//	// In a handler:
//	engine.RunLLM(ctx, update, agentic.LLMWithAgent("fast"))
func WithAgents(agents ...NamedAgent) EngineOption {
	return func(e *Engine) {
		if e.agents == nil {
			e.agents = make(map[string]fantasy.Agent, len(agents))
		}
		for _, a := range agents {
			e.agents[a.Name] = a.Agent
		}
	}
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
// If nil, errors are logged but not sent to the user (LLM should use tools to communicate).
func WithErrorHandler(fn func(ctx context.Context, update *IncomingUpdate, err error) error) EngineOption {
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

// Agent returns the default Fantasy agent.
func (e *Engine) Agent() fantasy.Agent { return e.agent }

// AgentByName returns a named agent, or the default if not found.
func (e *Engine) AgentByName(name string) fantasy.Agent { return e.resolveAgent(name) }

// Memory returns the configured Memory, or nil if not set.
func (e *Engine) Memory() Memory { return e.memory }

// resolveAgent returns the named agent if it exists, otherwise the default.
func (e *Engine) resolveAgent(name string) fantasy.Agent {
	if name != "" && e.agents != nil {
		if a, ok := e.agents[name]; ok {
			return a
		}
	}
	return e.agent
}

// Close performs graceful shutdown.
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
	agentName   string // named agent to use (empty = default)
	temperature *float64
	maxTokens   *int64
	extraTools  []fantasy.AgentTool
	stopWhen    []fantasy.StopCondition
	prepareStep fantasy.PrepareStepFunction
	activeTools []string
	extraMsgs   []fantasy.Message
}

// LLMWithAgent selects a named agent registered via WithAgents.
// If the name is not found, the default agent is used.
func LLMWithAgent(name string) LLMOption {
	return func(c *llmConfig) { c.agentName = name }
}

// LLMWithTemperature overrides temperature for this call.
func LLMWithTemperature(t float64) LLMOption {
	return func(c *llmConfig) { c.temperature = &t }
}

// LLMWithMaxTokens overrides max output tokens for this call.
func LLMWithMaxTokens(n int64) LLMOption {
	return func(c *llmConfig) { c.maxTokens = &n }
}

// LLMWithTools adds extra tools for this specific call.
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

// LLMWithActiveTools filters which tools are active for this call.
func LLMWithActiveTools(names ...string) LLMOption {
	return func(c *llmConfig) { c.activeTools = names }
}

// RunLLM executes the Fantasy agent pipeline and auto-sends the LLM response.
// If the LLM produces a non-empty text response, it is sent via the Channel
// from context. If the LLM already sent messages through tools (e.g. send_text),
// both the tool-sent message and the final text response are delivered.
func (e *Engine) RunLLM(ctx context.Context, update *IncomingUpdate, opts ...LLMOption) error {
	result, err := e.CallLLM(ctx, update, opts...)
	if err != nil {
		return err
	}
	if result != nil && result.Text != "" {
		ch := ChannelFromContext(ctx)
		if ch != nil {
			if sendErr := SendText(ctx, ch, update.ConversationID, result.Text); sendErr != nil {
				nxlog.Warn("failed to auto-send LLM response",
					zap.Int64("conversation_id", update.ConversationID),
					zap.Error(sendErr),
				)
			}
		}
	}
	return nil
}

// RunLLMHandler returns a Handler that calls RunLLM with the given options.
func (e *Engine) RunLLMHandler(opts ...LLMOption) Handler {
	return func(ctx context.Context, update *IncomingUpdate) error {
		return e.RunLLM(ctx, update, opts...)
	}
}

// CallLLM executes the LLM pipeline and returns the full result.
// Unlike RunLLM, the caller gets access to the LLMResult for inspection.
func (e *Engine) CallLLM(ctx context.Context, update *IncomingUpdate, opts ...LLMOption) (*LLMResult, error) {
	cfg := e.buildLLMConfig(opts)
	start := time.Now()

	// 1. Load memory.
	memKey := MemoryKey{
		AgentID:        e.agentID,
		UserID:         update.UserID,
		ConversationID: update.ConversationID,
	}
	history := e.loadMemory(ctx, memKey)

	// 2. Build prompt.
	var systemPrompt string
	if e.systemFn != nil {
		systemPrompt = e.systemFn(ctx, update)
	}

	if len(cfg.extraMsgs) > 0 {
		history = append(history, cfg.extraMsgs...)
	}

	// Append current user message.
	if update.Text != "" {
		history = append(history, fantasy.NewUserMessage(update.Text))
	}

	// 3. Fire event.
	e.events.OnLLMStart(ctx, update.UserID, update.Text)

	// 4. Call Fantasy agent.
	call := e.buildAgentCall(systemPrompt, history, cfg)
	selectedAgent := e.resolveAgent(cfg.agentName)
	result, err := selectedAgent.Generate(ctx, call)
	if err != nil {
		e.events.OnError(ctx, err)
		nxlog.Debug("llm generate failed",
			zap.Int32("user_id", update.UserID),
			zap.Int64("conversation_id", update.ConversationID),
			zap.Duration("elapsed", time.Since(start)),
			zap.Error(err),
		)
		if e.errorFn != nil {
			return nil, e.errorFn(ctx, update, err)
		}
		return nil, err
	}

	responseText := result.Response.Content.Text()

	// 5. Fire end event.
	usage := TokenUsage{
		InputTokens:  int(result.TotalUsage.InputTokens),
		OutputTokens: int(result.TotalUsage.OutputTokens),
	}
	e.events.OnLLMEnd(ctx, update.UserID, responseText, usage)

	nxlog.Debug("llm generate complete",
		zap.Int32("user_id", update.UserID),
		zap.Int64("conversation_id", update.ConversationID),
		zap.Duration("elapsed", time.Since(start)),
		zap.Int("steps", len(result.Steps)),
		zap.Int("input_tokens", usage.InputTokens),
		zap.Int("output_tokens", usage.OutputTokens),
		zap.String("response", responseText),
	)

	// 6. Save memory.
	e.saveMemory(ctx, memKey, history, result)

	return &LLMResult{Text: responseText, Result: result}, nil
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
		nxlog.Warn("failed to load memory", zap.Error(err), zap.Int32("user_id", key.UserID))
		return nil
	}
	return toFantasyMessages(msgs)
}

func (e *Engine) saveMemory(ctx context.Context, key MemoryKey, history []fantasy.Message, result *fantasy.AgentResult) {
	if e.memory == nil {
		return
	}
	updated := fromFantasyMessages(history)

	for _, step := range result.Steps {
		for _, msg := range step.Messages {
			updated = append(updated, fromFantasyMessage(msg))
		}
	}

	if err := e.memory.Save(ctx, key, updated); err != nil {
		nxlog.Warn("failed to save memory", zap.Error(err), zap.Int32("user_id", key.UserID))
	}
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

	if len(m.Parts) > 0 {
		parts := make([]fantasy.MessagePart, 0, len(m.Parts))
		for _, p := range m.Parts {
			switch p.Type {
			case MessagePartTypeText:
				parts = append(parts, fantasy.TextPart{Text: p.Text})
			case MessagePartTypeToolCall:
				parts = append(parts, fantasy.ToolCallPart{
					ToolCallID: p.ToolCallID,
					ToolName:   p.ToolName,
					Input:      p.Input,
				})
			case MessagePartTypeToolResult:
				parts = append(parts, fantasy.ToolResultPart{
					ToolCallID: p.ToolCallID,
					Output:     toolResultOutput(p),
				})
			}
		}
		return fantasy.Message{Role: role, Content: parts}
	}

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
			parts = append(parts, MessagePart{Type: MessagePartTypeText, Text: p.Text})
		case fantasy.ToolCallPart:
			parts = append(parts, MessagePart{
				Type:       MessagePartTypeToolCall,
				ToolCallID: p.ToolCallID,
				ToolName:   p.ToolName,
				Input:      p.Input,
			})
		case fantasy.ToolResultPart:
			mp := MessagePart{
				Type:       MessagePartTypeToolResult,
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
			if tp, ok := fantasy.AsMessagePart[fantasy.TextPart](part); ok {
				textBuilder.WriteString(tp.Text)
				parts = append(parts, MessagePart{Type: MessagePartTypeText, Text: tp.Text})
			}
		}
	}

	msg.Content = textBuilder.String()

	hasNonText := false
	for _, p := range parts {
		if p.Type != MessagePartTypeText {
			hasNonText = true
			break
		}
	}
	if hasNonText {
		msg.Parts = parts
	}

	return msg
}
