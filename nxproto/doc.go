// Package nxproto provides proto-related utilities for the Nexus ecosystem.
// Shared by both server (nexus-ai) and client (nexus-x/client).
//
// Capabilities:
//   - Sensitive field redaction (proto reflection, scanned once at startup)
//   - ProtoJSON lazy zap field (zero-cost when log level is disabled)
//   - Connect RPC logging interceptor (client and server side)
//   - Business error construction and parsing (ErrorDetail)
package nxproto
