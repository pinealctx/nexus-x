package nxconfig

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// testConfig implements Config for testing.
type testConfig struct {
	Host    string        `yaml:"host" json:"host"`
	Port    int           `yaml:"port" json:"port"`
	Verbose bool          `yaml:"verbose" json:"verbose"`
	DB      testDBConfig  `yaml:"db" json:"db"`
	Tags    []string      `yaml:"tags" json:"tags"`
}

type testDBConfig struct {
	Driver string `yaml:"driver" json:"driver"`
	DSN    string `yaml:"dsn" json:"dsn"`
}

func (c *testConfig) SetDefaults() {
	if c.Host == "" {
		c.Host = "localhost"
	}
	if c.Port == 0 {
		c.Port = 8080
	}
	if c.DB.Driver == "" {
		c.DB.Driver = "postgres"
	}
}

func (c *testConfig) Validate() error { return nil }

func TestLoad_FileSourceNoOverlay(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "config.yaml")

	content := "host: example.com\nport: 3000\nverbose: true\n"
	if err := os.WriteFile(base, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	var cfg testConfig
	if err := Load(context.Background(), &cfg, NewFileSource(base)); err != nil {
		t.Fatal(err)
	}

	if cfg.Host != "example.com" {
		t.Errorf("Host = %q, want %q", cfg.Host, "example.com")
	}
	if cfg.Port != 3000 {
		t.Errorf("Port = %d, want %d", cfg.Port, 3000)
	}
	if !cfg.Verbose {
		t.Error("Verbose = false, want true")
	}
}

func TestLoad_FileSourceOverlayMerge(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "config.yaml")
	local := filepath.Join(dir, "config.local.yaml")

	baseContent := "host: example.com\nport: 3000\nverbose: true\ndb:\n  driver: mysql\n  dsn: base_dsn\n"
	if err := os.WriteFile(base, []byte(baseContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Overlay only overrides port and db.dsn.
	localContent := "port: 9090\ndb:\n  dsn: local_dsn\n"
	if err := os.WriteFile(local, []byte(localContent), 0o644); err != nil {
		t.Fatal(err)
	}

	var cfg testConfig
	if err := Load(context.Background(), &cfg, NewFileSource(base)); err != nil {
		t.Fatal(err)
	}

	// Overridden by overlay.
	if cfg.Port != 9090 {
		t.Errorf("Port = %d, want %d (from overlay)", cfg.Port, 9090)
	}
	if cfg.DB.DSN != "local_dsn" {
		t.Errorf("DB.DSN = %q, want %q (from overlay)", cfg.DB.DSN, "local_dsn")
	}

	// Preserved from base.
	if cfg.Host != "example.com" {
		t.Errorf("Host = %q, want %q (from base)", cfg.Host, "example.com")
	}
	if !cfg.Verbose {
		t.Error("Verbose = false, want true (from base)")
	}
	if cfg.DB.Driver != "mysql" {
		t.Errorf("DB.Driver = %q, want %q (from base)", cfg.DB.Driver, "mysql")
	}
}

func TestLoad_FileSourceOverlayNewField(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "config.yaml")
	local := filepath.Join(dir, "config.local.yaml")

	baseContent := "host: example.com\n"
	if err := os.WriteFile(base, []byte(baseContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Overlay adds a field not in base.
	localContent := "tags:\n  - alpha\n  - beta\n"
	if err := os.WriteFile(local, []byte(localContent), 0o644); err != nil {
		t.Fatal(err)
	}

	var cfg testConfig
	if err := Load(context.Background(), &cfg, NewFileSource(base)); err != nil {
		t.Fatal(err)
	}

	if len(cfg.Tags) != 2 || cfg.Tags[0] != "alpha" || cfg.Tags[1] != "beta" {
		t.Errorf("Tags = %v, want [alpha beta]", cfg.Tags)
	}
	if cfg.Host != "example.com" {
		t.Errorf("Host = %q, want %q (from base)", cfg.Host, "example.com")
	}
}

func TestLoad_FileSourceDefaultsStillApply(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "config.yaml")
	local := filepath.Join(dir, "config.local.yaml")

	// Base sets only host.
	baseContent := "host: example.com\n"
	if err := os.WriteFile(base, []byte(baseContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Overlay sets only verbose.
	localContent := "verbose: true\n"
	if err := os.WriteFile(local, []byte(localContent), 0o644); err != nil {
		t.Fatal(err)
	}

	var cfg testConfig
	if err := Load(context.Background(), &cfg, NewFileSource(base)); err != nil {
		t.Fatal(err)
	}

	// Port and DB.Driver should come from SetDefaults.
	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want %d (default)", cfg.Port, 8080)
	}
	if cfg.DB.Driver != "postgres" {
		t.Errorf("DB.Driver = %q, want %q (default)", cfg.DB.Driver, "postgres")
	}
}

func TestLoad_StaticSourceNoOverlay(t *testing.T) {
	// StaticSource does not implement OverlayLoader, so no overlay logic.
	data := "host: static.com\nport: 5555\n"
	var cfg testConfig
	if err := Load(context.Background(), &cfg, NewStaticSource([]byte(data), FormatYAML)); err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "static.com" {
		t.Errorf("Host = %q, want %q", cfg.Host, "static.com")
	}
	if cfg.Port != 5555 {
		t.Errorf("Port = %d, want %d", cfg.Port, 5555)
	}
}

func TestLoad_JSONOverlayMerge(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "config.json")
	local := filepath.Join(dir, "config.local.json")

	baseContent := `{"host":"example.com","port":3000,"verbose":true}`
	if err := os.WriteFile(base, []byte(baseContent), 0o644); err != nil {
		t.Fatal(err)
	}

	localContent := `{"port":9090}`
	if err := os.WriteFile(local, []byte(localContent), 0o644); err != nil {
		t.Fatal(err)
	}

	var cfg testConfig
	if err := Load(context.Background(), &cfg, NewFileSource(base)); err != nil {
		t.Fatal(err)
	}

	if cfg.Port != 9090 {
		t.Errorf("Port = %d, want %d (from overlay)", cfg.Port, 9090)
	}
	if cfg.Host != "example.com" {
		t.Errorf("Host = %q, want %q (from base)", cfg.Host, "example.com")
	}
	if !cfg.Verbose {
		t.Error("Verbose = false, want true (from base)")
	}
}
