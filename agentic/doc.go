// Package agentic provides a conversational agent framework built on top of Fantasy.
//
// It handles the conversation layer above LLM: message routing, middleware chains,
// memory management, streaming, and observability. Fantasy handles the LLM layer
// (tool calling, retries, structured output) and is exposed directly — not wrapped.
package agentic
