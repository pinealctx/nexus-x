# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Reusable Go packages extracted from the Nexus ecosystem. Provides configuration loading, logging, proto utilities, Adaptive Card builder, and higher-level Agent engine + IM client wrappers. Published as `github.com/pinealctx/nexus-x`. Go 1.26.2.

## Commands

```bash
go test ./...           # Test all packages
go build ./...          # Build all packages
```

No Makefile, Taskfile, or pre-commit hooks.

## Structure

```
nxutil/         Pure utility functions, zero external deps (convid, hmac, time)
nxconfig/       Configuration loading (Source abstraction, YAML/JSON, Cobra, TLS)
  awssm/        AWS Secrets Manager config source
nxlog/          Global structured logger (zap-based, level-safe)
nxproto/        Proto utilities (redact, interceptor, error re-export from nexus-proto/errors)
adaptivecard/   Adaptive Card type-safe builder (pure Go, zero external deps)
agentic/        Nexus Agent SDK — Engine, Router, Middleware, Memory, Channel, Tools
  tools/        Built-in LLM tools for Nexus IM (messaging, queries, groups, media)
  llmconfig/    LLM configuration types
client/         Nexus IM Client SDK — Channel impl, Streaming, Webhook, WebSocket, Services
```

## Dependency Graph

```
nxutil          ← zero deps (stdlib only)
nxconfig        ← yaml.v3, cobra
nxlog           ← zap
nxproto         ← nexus-proto, zap, connect
adaptivecard    ← zero deps (encoding/json only)
agentic         ← nexus-proto, fantasy, adaptivecard
client          ← nexus-proto, connect, nxutil, agentic, websocket, redis
```

## Key Dependencies

- `charm.land/fantasy` — LLM agent framework (used by agentic)
- `connectrpc.com/connect` — Connect RPC (used by client, nxproto)
- `github.com/pinealctx/nexus-proto` — Protobuf types + error system
- `github.com/coder/websocket` — WebSocket client (used by client)
- `github.com/redis/go-redis` — Redis (used by client for memory/dedup)
- `github.com/aws/aws-sdk-go-v2` — AWS SDK (used by nxconfig/awssm)

## Code Standards

- All code, comments, identifiers in English.
- Follow Effective Go + Uber Go Style Guide.
- `nxutil/` and `adaptivecard/` must have zero external dependencies.
- `nxconfig/` provides Source interface + format auto-detection. No IM-specific types.
- `nxlog/` is a thin zap wrapper. No IM-specific types.
- `nxproto/` depends on nexus-proto + proto/connect ecosystem. Re-exports nexus-proto/errors.
- `agentic/` directly imports nexus-proto types. Channel interface uses proto types. Tools expose Nexus IM capabilities to LLM.
- `client/` implements agentic.Channel. Exposes all Connect RPC service clients via `Services()`.

## Language Policy

Conversational replies and spec documents in Simplified Chinese. All code in English.
