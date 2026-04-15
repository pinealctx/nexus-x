# nexus-x

Reusable Go packages extracted from the Nexus ecosystem.

## Structure

```
nxutil/         Pure utility functions, zero external deps (convid, hmac, time)
nxproto/        Proto-related tools, depends on nexus-proto (redact, interceptor, error)
agentic/        Conversational Agent Framework (Engine, Router, Middleware, Memory, Events)
client/         Nexus IM Client SDK (Channel, Streaming, Query, MiniApp)
adaptivecard/   Adaptive Card type-safe builder (pure Go, zero external deps)
```

## Dependency Graph

```
nxutil          ← zero deps (stdlib only)
nxproto         ← nexus-proto, zap, connect
agentic         ← fantasy
client          ← nexus-proto, connect, nxutil, nxproto, agentic
adaptivecard    ← zero deps (encoding/json only)
```

## Code Standards

- All code, comments, identifiers in English.
- Follow Effective Go + Uber Go Style Guide.
- `nxutil/` must have zero external dependencies.
- `nxproto/` depends only on nexus-proto + standard proto/connect ecosystem.
- `agentic/` must NOT import any IM-specific types. It defines the `Channel` interface.
- `client/` is the ONLY package that imports nexus-proto service clients.
- `adaptivecard/` is pure Go with zero external dependencies.
