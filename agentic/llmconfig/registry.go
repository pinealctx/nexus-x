package llmconfig

import (
	"context"
	"fmt"

	"charm.land/fantasy"
	"charm.land/fantasy/providers/anthropic"
	"charm.land/fantasy/providers/azure"
	"charm.land/fantasy/providers/bedrock"
	"charm.land/fantasy/providers/google"
	"charm.land/fantasy/providers/openai"
	"charm.land/fantasy/providers/openaicompat"
	"charm.land/fantasy/providers/openrouter"
)

// Registry holds initialized Fantasy providers and creates language models on demand.
type Registry struct {
	providers map[string]fantasy.Provider
	models    map[string]ModelConfig
}

// NewRegistry initializes all providers from config and returns a Registry.
func NewRegistry(cfg Config) (*Registry, error) {
	providers := make(map[string]fantasy.Provider, len(cfg.Providers))
	for name, pc := range cfg.Providers {
		p, err := newProvider(pc)
		if err != nil {
			return nil, fmt.Errorf("llmconfig: provider %q: %w", name, err)
		}
		providers[name] = p
	}
	return &Registry{
		providers: providers,
		models:    cfg.Models,
	}, nil
}

// Model returns a LanguageModel by its logical name (key in Config.Models).
func (r *Registry) Model(ctx context.Context, name string) (fantasy.LanguageModel, error) {
	mc, ok := r.models[name]
	if !ok {
		return nil, fmt.Errorf("llmconfig: unknown model %q", name)
	}
	p, ok := r.providers[mc.Provider]
	if !ok {
		return nil, fmt.Errorf("llmconfig: model %q references unknown provider %q", name, mc.Provider)
	}
	return p.LanguageModel(ctx, mc.Model)
}

// MustModel is like Model but panics on error. Useful during initialization.
func (r *Registry) MustModel(ctx context.Context, name string) fantasy.LanguageModel {
	m, err := r.Model(ctx, name)
	if err != nil {
		panic(err)
	}
	return m
}

// Provider returns the underlying Fantasy provider by its logical name.
func (r *Registry) Provider(name string) (fantasy.Provider, bool) {
	p, ok := r.providers[name]
	return p, ok
}

// ModelNames returns all registered model names.
func (r *Registry) ModelNames() []string {
	names := make([]string, 0, len(r.models))
	for name := range r.models {
		names = append(names, name)
	}
	return names
}

// newProvider creates a Fantasy provider from config.
func newProvider(pc ProviderConfig) (fantasy.Provider, error) {
	apiKey := expandEnv(pc.APIKey)
	baseURL := expandEnv(pc.BaseURL)

	switch pc.Provider {
	case ProviderAnthropic:
		var opts []anthropic.Option
		if apiKey != "" {
			opts = append(opts, anthropic.WithAPIKey(apiKey))
		}
		if baseURL != "" {
			opts = append(opts, anthropic.WithBaseURL(baseURL))
		}
		return anthropic.New(opts...)

	case ProviderOpenAI:
		var opts []openai.Option
		if apiKey != "" {
			opts = append(opts, openai.WithAPIKey(apiKey))
		}
		if baseURL != "" {
			opts = append(opts, openai.WithBaseURL(baseURL))
		}
		return openai.New(opts...)

	case ProviderAzure:
		var opts []azure.Option
		if apiKey != "" {
			opts = append(opts, azure.WithAPIKey(apiKey))
		}
		if baseURL != "" {
			opts = append(opts, azure.WithBaseURL(baseURL))
		}
		if pc.APIVersion != "" {
			opts = append(opts, azure.WithAPIVersion(pc.APIVersion))
		}
		return azure.New(opts...)

	case ProviderGoogle:
		var opts []google.Option
		if apiKey != "" {
			opts = append(opts, google.WithGeminiAPIKey(apiKey))
		}
		if baseURL != "" {
			opts = append(opts, google.WithBaseURL(baseURL))
		}
		project := expandEnv(pc.Project)
		location := expandEnv(pc.Location)
		if project != "" && location != "" {
			opts = append(opts, google.WithVertex(project, location))
		}
		return google.New(opts...)

	case ProviderBedrock:
		var opts []bedrock.Option
		if apiKey != "" {
			opts = append(opts, bedrock.WithAPIKey(apiKey))
		}
		if baseURL != "" {
			opts = append(opts, bedrock.WithBaseURL(baseURL))
		}
		return bedrock.New(opts...)

	case ProviderOpenRouter:
		var opts []openrouter.Option
		if apiKey != "" {
			opts = append(opts, openrouter.WithAPIKey(apiKey))
		}
		return openrouter.New(opts...)

	case ProviderOpenAICompat:
		var opts []openaicompat.Option
		if apiKey != "" {
			opts = append(opts, openaicompat.WithAPIKey(apiKey))
		}
		if baseURL != "" {
			opts = append(opts, openaicompat.WithBaseURL(baseURL))
		}
		return openaicompat.New(opts...)

	default:
		return nil, fmt.Errorf("unsupported provider type %q", pc.Provider)
	}
}
