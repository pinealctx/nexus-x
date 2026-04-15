// Package llmconfig provides config-driven initialization of Fantasy providers
// and language models. It bridges YAML/JSON configuration with Fantasy's
// provider system, eliminating boilerplate when constructing agents.
//
// Supported providers: anthropic, openai, azure, google, bedrock, openrouter, openaicompat.
//
// Usage:
//
//	var cfg llmconfig.Config
//	nxconfig.Load(ctx, &cfg, nxconfig.NewFileSource("llm.yaml"))
//
//	registry, err := llmconfig.NewRegistry(ctx, cfg)
//	model, err := registry.Model(ctx, "default")
//	agent := fantasy.NewAgent(model, fantasy.WithTools(...))
//
// YAML example:
//
//	providers:
//	  anthropic:
//	    provider: anthropic
//	    api_key: ${ANTHROPIC_API_KEY}
//	  openai:
//	    provider: openai
//	    api_key: ${OPENAI_API_KEY}
//
//	models:
//	  default:
//	    provider: anthropic
//	    model: claude-sonnet-4-20250514
//	  fast:
//	    provider: anthropic
//	    model: claude-haiku-4-5-20251001
//	  gpt4:
//	    provider: openai
//	    model: gpt-4o
package llmconfig
