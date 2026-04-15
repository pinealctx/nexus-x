# nexus-x

Go SDK for building Nexus IM Agents.

## Install

```bash
go get github.com/pinealctx/nexus-x
```

## Quick Start

```go
package main

import (
    "context"
    "net/http"

    "charm.land/fantasy"
    "charm.land/fantasy/providers/anthropic"
    "github.com/pinealctx/nexus-x/agentic"
    "github.com/pinealctx/nexus-x/agentic/tools"
    "github.com/pinealctx/nexus-x/client"
)

func main() {
    // 1. Create Nexus client.
    c := client.New("nxa_your_token", "https://nexus.example.com",
        client.WithSecretKey("your_secret_key"),
    )

    // 2. Create Fantasy LLM agent with Nexus tools.
    provider, _ := anthropic.New()
    model, _ := provider.LanguageModel(context.Background(), "claude-sonnet-4-20250514")
    agent := fantasy.NewAgent(model,
        fantasy.WithTools(tools.BasicTools(c)...),
    )

    // 3. Build engine.
    engine, _ := agentic.NewEngine(
        agentic.WithAgent(agent),
        agentic.WithRouter(agentic.NewRouter()),
        agentic.WithSystemPrompt("You are a helpful assistant."),
    )

    // 4. Start webhook server.
    http.Handle("/webhook", c.WebhookHandler(engine.Handle))
    http.ListenAndServe(":8080", nil)
}
```

## Packages

| Package | Description |
|---------|-------------|
| `client` | Nexus IM client — Channel implementation, WebSocket, Webhook, all Connect RPC services via `Services()` |
| `agentic` | Agent engine — Router, Middleware, Memory, Channel interface, convenience send functions |
| `agentic/tools` | Built-in LLM tools — messaging, queries, groups, media (Fantasy AgentTool) |
| `adaptivecard` | Adaptive Card type-safe builder (zero deps) |
| `nxconfig` | Config loading — Source abstraction, YAML/JSON auto-detect, Cobra flags, TLS |
| `nxconfig/awssm` | AWS Secrets Manager config source |
| `nxlog` | Global structured logger (zap) |
| `nxproto` | Proto utilities — sensitive field redaction, Connect RPC interceptor, error re-export |
| `nxutil` | Pure utility functions — conversation ID encoding, HMAC, time (zero deps) |

## Agent Tools

LLM tools let the model interact with Nexus IM through tool calling. Import individually or by tier:

```go
// Individual tools
fantasy.WithTools(
    tools.SendText(c),
    tools.GetMessageHistory(c),
    tools.GetGroupInfo(c),
)

// By tier
fantasy.WithTools(tools.BasicTools(c)...)    // send, edit, reply, card
fantasy.WithTools(tools.QueryTools(c)...)    // history, conversation, search
fantasy.WithTools(tools.GroupTools(c)...)    // group info, members
fantasy.WithTools(tools.MediaTools(c)...)    // image, file, download
fantasy.WithTools(tools.AllTools(c)...)      // everything
```

## Middleware

```go
engine, _ := agentic.NewEngine(
    agentic.WithMiddleware(
        agentic.AgentFilterMiddleware(selfID),  // private: all, group: @mention only
        agentic.DedupMiddleware(dedup),
        agentic.RateLimitMiddleware(limiter, nil),
        agentic.LoggingMiddleware(),
        agentic.RecoveryMiddleware(),
    ),
    // ...
)
```

## Config Loading

```go
// File only
nxconfig.RegisterFlags(cmd)
nxconfig.LoadFromFlags(ctx, cmd, &cfg)

// With AWS Secrets Manager fallback
awssm.RegisterFlags(cmd)
nxconfig.LoadFromFlags(ctx, cmd, &cfg, awssm.SourceFromFlags(cmd))

// Programmatic
nxconfig.Load(ctx, &cfg, nxconfig.NewFileSource("config.yaml"))
nxconfig.Load(ctx, &cfg, nxconfig.NewEnvSource("APP_CONFIG"))
```

## Direct RPC Access

For operations not covered by high-level methods, use the underlying Connect RPC clients:

```go
resp, err := c.Services().Groups.GetGroupInfo(ctx, connect.NewRequest(&apiv1.GetGroupInfoRequest{
    GroupId: 42,
}))
```

## License

MIT
