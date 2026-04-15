package nxconfig

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Source provides raw configuration bytes from an external store.
// Implementations must be safe for concurrent use.
type Source interface {
	// Load returns the raw configuration bytes.
	Load(ctx context.Context) ([]byte, error)
}

// --- FileSource ---

// FileSource reads configuration from a local file (YAML or JSON).
// When a ".local" variant exists (e.g. server.local.yaml for server.yaml),
// it is used instead — this allows per-developer overrides without
// touching the committed config file.
type FileSource struct {
	Path string
}

// NewFileSource creates a FileSource for the given path.
func NewFileSource(path string) *FileSource {
	return &FileSource{Path: path}
}

// Load reads the config file, preferring the .local variant if it exists.
func (s *FileSource) Load(_ context.Context) ([]byte, error) {
	path := s.Path
	if lp := localPath(path); fileExists(lp) {
		path = lp
	}
	data, err := os.ReadFile(path) //nolint:gosec // config path from CLI flag
	if err != nil {
		return nil, fmt.Errorf("read config file %s: %w", path, err)
	}
	return data, nil
}

// FormatHint implements FormatHinter by checking the file extension.
func (s *FileSource) FormatHint() Format {
	return formatFromExt(s.Path)
}

// --- EnvSource ---

// EnvSource reads configuration from an environment variable.
type EnvSource struct {
	Key string
}

// NewEnvSource creates an EnvSource for the given environment variable name.
func NewEnvSource(key string) *EnvSource {
	return &EnvSource{Key: key}
}

// Load reads the environment variable value.
func (s *EnvSource) Load(_ context.Context) ([]byte, error) {
	v := os.Getenv(s.Key)
	if v == "" {
		return nil, fmt.Errorf("environment variable %s is not set", s.Key)
	}
	return []byte(v), nil
}

// --- StaticSource ---

// StaticSource returns fixed bytes. Useful for testing and embedding.
type StaticSource struct {
	Data   []byte
	Format Format
}

// NewStaticSource creates a StaticSource with the given data and format.
func NewStaticSource(data []byte, format Format) *StaticSource {
	return &StaticSource{Data: data, Format: format}
}

// Load returns the static data.
func (s *StaticSource) Load(_ context.Context) ([]byte, error) {
	return s.Data, nil
}

// FormatHint implements FormatHinter.
func (s *StaticSource) FormatHint() Format {
	return s.Format
}

// --- Helpers ---

// localPath derives the ".local" variant of a config file path.
// Example: "configs/server.yaml" -> "configs/server.local.yaml"
func localPath(p string) string {
	ext := filepath.Ext(p)
	base := strings.TrimSuffix(p, ext)
	return base + ".local" + ext
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func formatFromExt(path string) Format {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		return FormatJSON
	case ".yaml", ".yml":
		return FormatYAML
	default:
		return FormatAuto
	}
}
