# agentic

Conversational agent framework for Nexus IM, built on [Fantasy](https://charm.land/fantasy).

Handles the conversation layer above LLM: message routing, keyword interception, middleware chains, memory management, and observability. Fantasy handles the LLM layer (tool calling, retries, structured output) and is exposed directly — not wrapped.

## Architecture

```
IncomingUpdate
    │
    ▼
Engine.Handle()
    │
    ├─ Set context (UserID, ConversationID, Channel, Memory)
    │
    ▼
Middleware Chain
    │  RecoveryMiddleware         ← catch panics
    │  LoggingMiddleware          ← debug logging
    │  AgentFilterMiddleware      ← skip self-messages, group @mention filter
    │  DedupMiddleware            ← skip duplicate messages
    │  RateLimitMiddleware        ← reject over-limit users
    │  CredentialGateMiddleware   ← block unauthorized users
    │
    ▼
Router.Handle()
    │
    ├─ CardAction?  → verb handler / card default / LLM
    ├─ /command?    → command handler
    ├─ Keyword?     → keyword handler (first match wins)
    └─ Otherwise    → LLM fallback
                         │
                         ▼
                    Engine.RunLLM()
                         │
                         ├─ Load memory
                         ├─ Build prompt + history
                         ├─ Call Fantasy Agent
                         ├─ (LLM uses tools to send messages)
                         └─ Save memory
```

### Core Components

| Component | Role |
|-----------|------|
| **Engine** | Central coordinator. Wires Router, middleware, Memory, and Fantasy Agent. |
| **Router** | Dispatches by priority: card actions → commands → keywords → LLM. |
| **Middleware** | Cross-cutting concerns: filtering, dedup, rate limiting, auth, logging. |
| **Memory** | Conversation history storage, scoped by agent + user + conversation. |
| **Channel** | Outbound messaging interface (text, cards, images, files, edit, recall). |

### Key Design Decisions

- **Engine only orchestrates, never sends.** LLM sends messages through tools (`send_text`, `send_card`), giving it full control over when and how to respond.
- **Fantasy is exposed, not wrapped.** `fantasy.Agent`, `fantasy.AgentTool`, and `fantasy.AgentCall` are used directly. No leaky abstractions.
- **Proto types flow through.** Channel interface uses `nexus-proto` types directly for full access to Nexus capabilities.

## Quick Start

```go
package main

import (
    "context"
    "net/http"
    "time"

    "charm.land/fantasy"
    "charm.land/fantasy/providers/anthropic"

    "github.com/pinealctx/nexus-x/agentic"
    "github.com/pinealctx/nexus-x/agentic/tools"
    "github.com/pinealctx/nexus-x/client"
)

func main() {
    // 1. Create a Nexus client.
    c, _ := client.NewWebhook(token, secret)

    // 2. Create a Fantasy agent with tools.
    provider, _ := anthropic.New()
    model, _ := provider.LanguageModel(context.Background(), "claude-sonnet-4-20250514")
    agent := fantasy.NewAgent(model,
        fantasy.WithTools(tools.BasicTools(c)...),
    )

    // 3. Build a router with commands and keywords.
    router := agentic.NewRouter(
        agentic.WithCommands(
            agentic.Command{Name: "help", Description: "Show help", Handler: helpFn},
            agentic.Command{Name: "ping", Description: "Pong!", Handler: pingFn},
        ),
        agentic.WithKeywordExact("你好", greetHandler),
        agentic.WithKeywordContains("价格", priceHandler),
    )

    // 4. Assemble the engine.
    engine, _ := agentic.NewEngine(
        agentic.WithAgent(agent),
        agentic.WithRouter(router),
        agentic.WithMemory(agentic.NewInMemoryMemory()),
        agentic.WithMiddleware(
            agentic.RecoveryMiddleware(),
            agentic.LoggingMiddleware(),
            agentic.AgentFilterMiddleware(c.SelfUserID()),
            agentic.DedupMiddleware(agentic.NewInMemoryDedup(5 * time.Minute)),
        ),
        agentic.WithSystemPrompt("You are a helpful assistant."),
    )

    // 5. Serve.
    http.Handle("/webhook", c.WebhookHandler(engine.Handle))
    http.ListenAndServe(":8080", nil)
}
```

## Router

The Router dispatches incoming updates by priority:

1. **Card actions** — `Action.Submit` events dispatched by verb.
2. **Slash commands** — Messages starting with `/` matched by command name.
3. **Keyword rules** — Custom matchers evaluated in registration order; first match wins.
4. **LLM fallback** — Everything else goes to the Fantasy Agent.

### Configuration

Router uses the `RouterOption` pattern:

```go
router := agentic.NewRouter(
    // Slash commands
    agentic.WithCommands(helpCmd, statusCmd),

    // Keyword rules (bypass LLM)
    agentic.WithKeywordExact("你好", greetHandler),       // exact match, case-insensitive
    agentic.WithKeywordContains("价格", priceHandler),     // substring match
    agentic.WithKeywordPrefix("查询", queryHandler),       // prefix match
    agentic.WithKeyword(func(text string) bool {           // custom matcher
        return len(text) < 3
    }, askMoreHandler),

    // Card action handlers
    agentic.WithCardAction("approve", approveHandler),
    agentic.WithCardAction("reject", rejectHandler),
    agentic.WithCardActionDefault(defaultCardHandler),

    // Custom LLM handler (optional, Engine auto-binds if not set)
    agentic.WithLLMHandler(customLLMHandler),
)
```

### Keyword Rules

Keyword rules are ideal for high-frequency, deterministic responses that don't need LLM reasoning:

| Use Case | Method | Example |
|----------|--------|---------|
| Greetings | `WithKeywordExact` | "你好", "hi", "hello" |
| FAQ triggers | `WithKeywordContains` | "价格", "退款", "营业时间" |
| Command-like prefixes | `WithKeywordPrefix` | "查询", "搜索" |
| Complex logic | `WithKeyword` | Message length, regex, time-based |

Rules are evaluated in registration order. The first match wins and short-circuits — no LLM call is made.

## Middleware

Built-in middleware (apply in this order):

```go
agentic.WithMiddleware(
    agentic.RecoveryMiddleware(),                          // catch panics
    agentic.LoggingMiddleware(),                           // debug logging
    agentic.AgentFilterMiddleware(selfID),                 // skip self, group @mention filter
    agentic.DedupMiddleware(dedup),                        // skip duplicates
    agentic.RateLimitMiddleware(limiter, onLimited),       // rate limiting
    agentic.CredentialGateMiddleware(checker, onMissing),  // credential check
)
```

Middleware is composed left-to-right via `Chain(A, B, C)(handler)` = `A(B(C(handler)))`.

Custom middleware follows the `func(next Handler) Handler` signature:

```go
func MyMiddleware() agentic.Middleware {
    return func(next agentic.Handler) agentic.Handler {
        return func(ctx context.Context, update *agentic.IncomingUpdate) error {
            // before
            err := next(ctx, update)
            // after
            return err
        }
    }
}
```

## Memory

Conversation history storage, scoped by `MemoryKey{AgentID, UserID, ConversationID}`.

| Implementation | Use Case | Backend |
|---------------|----------|---------|
| `InMemoryMemory` | Development, testing | `sync.Map` |
| `RedisMemory` | Production | Redis |

```go
// In-memory (dev)
mem := agentic.NewInMemoryMemory(
    agentic.WithMaxMessages(100),
)

// Redis (production)
mem := agentic.NewRedisMemory(redisClient,
    agentic.WithRedisPrefix("myagent:memory"),
    agentic.WithRedisMaxMessages(100),
    agentic.WithRedisTTL(24 * time.Hour),
    agentic.WithRedisSummarizer(mySummarizer),
)
```

## Tools (sub-package)

The `tools` sub-package provides built-in Fantasy AgentTools that expose Nexus IM capabilities to LLMs:

| Tier | Tools | Purpose |
|------|-------|---------|
| 1 Basic | `send_text`, `send_card`, `edit_message`, `reply_message` | Core messaging |
| 2 Query | `get_message_history`, `get_message`, `get_conversation`, `search_users` | Context awareness |
| 3 Group | `get_group_info`, `list_groups`, `invite_members`, `remove_member` | Group management |
| 4 Media | `send_image`, `send_file`, `get_download_url` | Media handling |

```go
// Pick what your agent needs:
tools.BasicTools(c)                    // Tier 1 only
tools.QueryTools(c)                    // Tier 2 only
tools.AllTools(c)                      // All tiers
append(tools.BasicTools(c), myTool)    // Mix built-in + custom
```

## Channel

Outbound messaging interface, implemented by the `client` package:

```go
// Convenience functions (use from middleware, command handlers, keyword handlers):
agentic.SendText(ctx, ch, convID, "Hello!")
agentic.SendCard(ctx, ch, convID, card)
agentic.SendAdaptiveCard(ctx, ch, convID, adaptiveCard)

// Get channel from context (set by Engine):
ch := agentic.ChannelFromContext(ctx)
```

## LLM Pipeline

`Engine.RunLLM` / `Engine.CallLLM` execute the Fantasy agent pipeline:

```go
// RunLLM — fire-and-forget (default LLM handler)
engine.RunLLM(ctx, update)

// CallLLM — get full result for inspection
result, err := engine.CallLLM(ctx, update,
    agentic.LLMWithTemperature(0.7),
    agentic.LLMWithMaxTokens(4096),
    agentic.LLMWithTools(extraTool),
    agentic.LLMWithActiveTools("send_text", "search_users"),
)
```

### Multi-Agent Routing

Register named agents for different scenarios (e.g. fast/cheap vs smart/expensive):

```go
provider, _ := anthropic.New()
haikuModel, _ := provider.LanguageModel(ctx, "claude-haiku-4-5-20251001")
sonnetModel, _ := provider.LanguageModel(ctx, "claude-sonnet-4-20250514")

haikuAgent := fantasy.NewAgent(haikuModel, fantasy.WithTools(tools.BasicTools(c)...))
sonnetAgent := fantasy.NewAgent(sonnetModel, fantasy.WithTools(tools.AllTools(c)...))

engine, _ := agentic.NewEngine(
    agentic.WithAgent(sonnetAgent), // default
    agentic.WithAgents(
        agentic.NamedAgent{Name: "fast", Agent: haikuAgent},
    ),
    agentic.WithRouter(agentic.NewRouter(
        agentic.WithKeywordExact("你好", greetHandler),       // no LLM
        agentic.WithLLMHandler(func(ctx context.Context, update *agentic.IncomingUpdate) error {
            if isSimpleQuery(update.Text) {
                return engine.RunLLM(ctx, update, agentic.LLMWithAgent("fast"))
            }
            return engine.RunLLM(ctx, update) // default agent
        }),
    )),
)
```

Three tiers of cost optimization:
1. **Keyword rules** — zero LLM cost, instant response
2. **Fast agent** — cheap model for simple queries
3. **Default agent** — full-capability model for complex tasks

## Events

Lifecycle hooks for observability:

```go
type MyEvents struct {
    agentic.NoopEventHandler // embed for partial implementation
}

func (e *MyEvents) OnLLMStart(ctx context.Context, userID int32, input string) {
    metrics.LLMRequestsTotal.Inc()
}

func (e *MyEvents) OnLLMEnd(ctx context.Context, userID int32, response string, tokens agentic.TokenUsage) {
    metrics.TokensUsed.Add(float64(tokens.InputTokens + tokens.OutputTokens))
}
```
