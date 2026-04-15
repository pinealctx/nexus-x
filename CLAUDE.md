# nexus-x

Reusable Go packages extracted from the Nexus ecosystem.

## Structure

```
nxutil/         Pure utility functions, zero external deps (convid, hmac, time)
nxconfig/       Configuration loading (Source abstraction, YAML/JSON, Cobra, TLS)
nxlog/          Global structured logger (zap-based, level-safe)
nxproto/        Proto-related tools (redact, interceptor, error re-export)
adaptivecard/   Adaptive Card type-safe builder (pure Go, zero external deps)
agentic/        Nexus Agent SDK — Engine, Router, Middleware, Memory, Channel, Tools
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
client          ← nexus-proto, connect, nxutil, agentic
```

## Code Standards

- All code, comments, identifiers in English.
- Follow Effective Go + Uber Go Style Guide.
- `nxutil/` must have zero external dependencies.
- `nxconfig/` provides Source interface + format auto-detection. No IM-specific types.
- `nxlog/` is a thin zap wrapper. No IM-specific types.
- `nxproto/` depends on nexus-proto + proto/connect ecosystem. Re-exports nexus-proto/errors.
- `agentic/` directly imports nexus-proto types. Channel interface uses proto types. Tools expose Nexus IM capabilities to LLM.
- `client/` implements agentic.Channel. Exposes all Connect RPC service clients via `Services()`.
- `adaptivecard/` is pure Go with zero external dependencies.
