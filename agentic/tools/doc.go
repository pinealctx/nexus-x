// Package tools provides built-in Fantasy AgentTools that expose Nexus IM
// capabilities to LLMs via tool calling. Each tool is individually exported
// and can be composed freely. Group functions (BasicTools, QueryTools, etc.)
// provide convenient bundles.
//
// Tools read conversation context (ConversationID, UserID, Channel) from
// the context set by Engine.Handle.
package tools
