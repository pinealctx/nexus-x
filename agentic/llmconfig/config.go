package llmconfig

import (
	"fmt"
	"os"
	"strings"
)

// Config is the top-level LLM configuration.
type Config struct {
	// Providers maps a logical name to provider connection settings.
	Providers map[string]ProviderConfig `yaml:"providers" json:"providers"`

	// Models maps a logical name to a model reference (provider + model ID).
	Models map[string]ModelConfig `yaml:"models" json:"models"`
}

// SetDefaults implements nxconfig.Config.
func (c *Config) SetDefaults() {
	for name, p := range c.Providers {
		if p.Provider == "" {
			p.Provider = ProviderType(name)
			c.Providers[name] = p
		}
	}
}

// Validate implements nxconfig.Config.
func (c *Config) Validate() error {
	if len(c.Providers) == 0 {
		return fmt.Errorf("llmconfig: at least one provider is required")
	}
	if len(c.Models) == 0 {
		return fmt.Errorf("llmconfig: at least one model is required")
	}
	for name, m := range c.Models {
		if m.Provider == "" {
			return fmt.Errorf("llmconfig: model %q missing provider", name)
		}
		if m.Model == "" {
			return fmt.Errorf("llmconfig: model %q missing model ID", name)
		}
		if _, ok := c.Providers[m.Provider]; !ok {
			return fmt.Errorf("llmconfig: model %q references unknown provider %q", name, m.Provider)
		}
	}
	return nil
}

// ProviderType identifies a Fantasy provider backend.
type ProviderType string

const (
	ProviderAnthropic   ProviderType = "anthropic"
	ProviderOpenAI      ProviderType = "openai"
	ProviderAzure       ProviderType = "azure"
	ProviderGoogle      ProviderType = "google"
	ProviderBedrock     ProviderType = "bedrock"
	ProviderOpenRouter  ProviderType = "openrouter"
	ProviderOpenAICompat ProviderType = "openaicompat"
)

// ProviderConfig holds connection settings for a single provider.
type ProviderConfig struct {
	// Provider is the backend type. If omitted, defaults to the map key name.
	Provider ProviderType `yaml:"provider" json:"provider"`

	// APIKey is the authentication key. Supports ${ENV_VAR} expansion.
	APIKey string `yaml:"api_key" json:"api_key"`

	// BaseURL overrides the default API endpoint.
	BaseURL string `yaml:"base_url" json:"base_url"`

	// --- Azure-specific ---

	// APIVersion is the Azure API version (default: "2025-01-01-preview").
	APIVersion string `yaml:"api_version,omitempty" json:"api_version,omitempty"`

	// --- Google / Vertex-specific ---

	// Project is the GCP project ID (for Vertex AI).
	Project string `yaml:"project,omitempty" json:"project,omitempty"`

	// Location is the GCP region (for Vertex AI).
	Location string `yaml:"location,omitempty" json:"location,omitempty"`
}

// ModelConfig references a provider and model ID.
type ModelConfig struct {
	// Provider is the logical name of the provider (key in Config.Providers).
	Provider string `yaml:"provider" json:"provider"`

	// Model is the model identifier (e.g. "claude-sonnet-4-20250514", "gpt-4o").
	Model string `yaml:"model" json:"model"`
}

// expandEnv replaces ${VAR} or $VAR in s with the corresponding environment variable.
func expandEnv(s string) string {
	if strings.Contains(s, "$") {
		return os.ExpandEnv(s)
	}
	return s
}
