// Package nxconfig provides a two-layer configuration loading system.
//
// Layer 1 — Source: abstracts where raw config bytes come from (file,
// AWS Secrets Manager, env var, static bytes, etc.).
//
// Layer 2 — Load: parses bytes into a typed struct (YAML or JSON,
// auto-detected or explicit), applies defaults, and validates.
//
// Usage:
//
//	var cfg MyConfig
//	err := nxconfig.Load(ctx, &cfg, nxconfig.NewFileSource("config.yaml"))
//
//	// JSON file:
//	err := nxconfig.Load(ctx, &cfg, nxconfig.NewFileSource("config.json"))
//
//	// Explicit format:
//	err := nxconfig.LoadAs(ctx, &cfg, source, nxconfig.FormatJSON)
package nxconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"dario.cat/mergo"
	"gopkg.in/yaml.v3"
)

// Config is implemented by all configuration struct pointers.
// It provides zero-value fallback logic and required-field validation.
type Config interface {
	SetDefaults()
	Validate() error
}

// Format specifies the configuration file format.
type Format int

const (
	// FormatAuto detects format from the Source (file extension, content type, etc.).
	FormatAuto Format = iota
	// FormatYAML parses as YAML.
	FormatYAML
	// FormatJSON parses as JSON.
	FormatJSON
)

// Load fetches raw bytes from source, auto-detects format, unmarshals
// into cfg, applies defaults, and validates.
func Load(ctx context.Context, cfg Config, source Source) error {
	return LoadAs(ctx, cfg, source, FormatAuto)
}

// OverlayLoader is optionally implemented by Source to provide overlay config bytes.
// When a source implements OverlayLoader, LoadAs merges the overlay on top of the
// base configuration — fields set in the overlay override the base.
type OverlayLoader interface {
	LoadOverlay(ctx context.Context) ([]byte, error)
}

// LoadAs is like Load but with an explicit format override.
func LoadAs(ctx context.Context, cfg Config, source Source, format Format) error {
	data, err := source.Load(ctx)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if format == FormatAuto {
		format = detectFormat(source, data)
	}

	if err = unmarshal(data, cfg, format); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	// Apply overlay if available.
	if ol, ok := source.(OverlayLoader); ok {
		overlay, oerr := ol.LoadOverlay(ctx)
		if oerr != nil {
			return fmt.Errorf("load overlay config: %w", oerr)
		}
		if overlay != nil {
			overlayCfg := reflect.New(reflect.ValueOf(cfg).Elem().Type()).Interface()
			if err = unmarshal(overlay, overlayCfg, format); err != nil {
				return fmt.Errorf("parse overlay config: %w", err)
			}
			if err = mergo.Merge(cfg, overlayCfg, mergo.WithOverride); err != nil {
				return fmt.Errorf("merge overlay config: %w", err)
			}
		}
	}

	cfg.SetDefaults()
	if err = cfg.Validate(); err != nil {
		return fmt.Errorf("validate config: %w", err)
	}
	return nil
}

func unmarshal(data []byte, v any, format Format) error {
	switch format {
	case FormatJSON:
		return json.Unmarshal(data, v)
	default:
		return yaml.Unmarshal(data, v)
	}
}

// detectFormat guesses the format from the source and data content.
func detectFormat(source Source, data []byte) Format {
	// Check if source provides a hint.
	if h, ok := source.(FormatHinter); ok {
		if f := h.FormatHint(); f != FormatAuto {
			return f
		}
	}

	// Fallback: sniff content.
	trimmed := strings.TrimSpace(string(data))
	if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
		return FormatJSON
	}
	return FormatYAML
}

// FormatHinter is optionally implemented by Source to hint at the data format.
type FormatHinter interface {
	FormatHint() Format
}
