// Package agentic provides a conversational agent framework for Nexus IM,
// built on top of [Fantasy] (charm.land/fantasy).
//
// It handles the conversation layer above LLM: message routing, keyword
// interception, middleware chains, memory management, and observability.
// Fantasy handles the LLM layer (tool calling, retries, structured output)
// and is exposed directly — not wrapped.
//
// # Architecture
//
// The framework is organized around five core components:
//
//   - [Engine] — Central coordinator. Wires Router, middleware, Memory, and
//     Fantasy Agent. Manages the full pipeline: receive update → middleware →
//     route → (keyword reply | command | LLM call) → save memory.
//
//   - [Router] — Dispatches incoming updates by priority: card actions →
//     slash commands → keyword rules → LLM fallback. Configured via
//     [RouterOption] functions.
//
//   - Middleware — Cross-cutting concerns (filtering, dedup, rate limiting,
//     credential gating, logging, recovery). Composed via [Chain] and applied
//     around the Router in [Engine].
//
//   - [Memory] — Conversation history storage. Scoped by agent + user +
//     conversation. Built-in implementations: [InMemoryMemory] (dev/test)
//     and [RedisMemory] (production).
//
//   - [Channel] — Outbound messaging interface. Sends text, cards, images,
//     files; edits and recalls messages. Implemented by the client package.
//     StreamingChannel extends Channel with real-time text streaming via
//     [StreamWriter] (Push/End/Error).
//
// # Message Flow
//
//	IncomingUpdate
//	    │
//	    ▼
//	Engine.Handle()
//	    │
//	    ├─ Set context (UserID, ConversationID, Channel, Memory)
//	    │
//	    ▼
//	Middleware Chain (left to right)
//	    │  RecoveryMiddleware
//	    │  LoggingMiddleware
//	    │  AgentFilterMiddleware  (skip self-messages, group @mention filter)
//	    │  DedupMiddleware        (skip duplicate messages)
//	    │  RateLimitMiddleware    (reject over-limit users)
//	    │  CredentialGateMiddleware
//	    │
//	    ▼
//	Router.Handle()
//	    │
//	    ├─ CardAction?  → verb handler / card default / LLM
//	    ├─ /command?    → command handler
//	    ├─ Keyword?     → keyword handler (first match wins)
//	    └─ Otherwise    → LLM fallback
//	                         │
//	                         ▼
//	                    Engine.RunLLM()
//	                         │
//	                         ├─ Load memory
//	                         ├─ Build prompt + history
//	                         ├─ Stream mode?
//	                         │   ├─ Yes: StartStream → Stream() → Push deltas → End
//	                         │   └─ No:  Generate() → SendText
//	                         ├─ (LLM uses tools to send messages)
//	                         └─ Save memory
//
// # Routing Priority
//
// The Router evaluates handlers in this order:
//
//  1. Card actions — Action.Submit events dispatched by verb.
//  2. Slash commands — Messages starting with "/" matched by command name.
//  3. Keyword rules — Custom matchers evaluated in registration order; first match wins.
//  4. LLM fallback — Everything else goes to the Fantasy Agent.
//
// Keyword rules are ideal for high-frequency, deterministic responses that
// don't need LLM reasoning (greetings, FAQ, price lookups, status checks).
//
// # Quick Start
//
//	// 1. Create a Nexus client.
//	c, _ := client.NewWebhook(token, secret)
//
//	// 2. Create a Fantasy agent with tools.
//	provider, _ := anthropic.New()
//	model, _ := provider.LanguageModel(ctx, "claude-sonnet-4-20250514")
//	agent := fantasy.NewAgent(model,
//	    fantasy.WithTools(tools.BasicTools(c)...),
//	)
//
//	// 3. Build a router with commands and keywords.
//	router := agentic.NewRouter(
//	    agentic.WithCommands(
//	        agentic.Command{Name: "help", Description: "Show help", Handler: helpFn},
//	    ),
//	    agentic.WithKeywordExact("你好", greetHandler),
//	    agentic.WithKeywordContains("价格", priceHandler),
//	)
//
//	// 4. Assemble the engine.
//	engine, _ := agentic.NewEngine(
//	    agentic.WithAgent(agent),
//	    agentic.WithRouter(router),
//	    agentic.WithMemory(agentic.NewInMemoryMemory()),
//	    agentic.WithStreamMode(), // enable real-time text streaming (optional)
//	    agentic.WithMiddleware(
//	        agentic.RecoveryMiddleware(),
//	        agentic.LoggingMiddleware(),
//	        agentic.AgentFilterMiddleware(c.SelfUserID()),
//	        agentic.DedupMiddleware(agentic.NewInMemoryDedup(5 * time.Minute)),
//	    ),
//	    agentic.WithSystemPrompt("You are a helpful assistant."),
//	)
//
//	// 5. Serve.
//	http.Handle("/webhook", c.WebhookHandler(engine.Handle))
//
// # Tools
//
// The [tools] sub-package provides built-in Fantasy AgentTools that expose
// Nexus IM capabilities to LLMs. Tools are organized in tiers:
//
//   - Tier 1 (Basic): send_text, send_card, edit_message, reply_message
//   - Tier 2 (Query): get_message_history, get_message, get_conversation, search_users
//   - Tier 3 (Group): get_group_info, list_groups, invite_members, remove_member
//   - Tier 4 (Media): send_image, send_file, get_download_url
//
// Engine supports two execution modes:
//
//   - Generate mode (default): calls fantasy.Generate(), sends the final text
//     response as a single message after all steps complete.
//   - Stream mode: calls fantasy.Stream(), pushes text deltas in real time via
//     StreamingChannel. Text arrives before tool execution, so natural ordering
//     is preserved. Enable with [WithStreamMode].
//
// [Fantasy]: https://charm.land/fantasy
package agentic
