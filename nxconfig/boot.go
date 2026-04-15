package nxconfig

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// RegisterFlags adds the standard --config flag to a cobra command.
func RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("config", "c", "", "path to config file (YAML or JSON) (env: CONFIG_FILE)")
}

// LoadFromFlags resolves the config source from flags/env and loads
// configuration into cfg.
//
// Resolution order:
//  1. --config flag or CONFIG_FILE env → FileSource
//  2. fallback sources (in order) → first non-nil Source is used
//  3. default file locations (config.yaml, config.yml, config.json)
//
// The fallback parameter allows callers to inject additional sources
// (e.g. AWS Secrets Manager) without nxconfig depending on them:
//
//	// Simple (file only):
//	nxconfig.LoadFromFlags(ctx, cmd, &cfg)
//
//	// With AWS SM fallback:
//	nxconfig.RegisterFlags(cmd)
//	cmd.Flags().StringP("region", "r", "", "AWS region (env: REGION)")
//	cmd.Flags().StringP("secret-name", "s", "", "secret name (env: SECRET_NAME)")
//	nxconfig.LoadFromFlags(ctx, cmd, &cfg, awsSourceFromFlags(cmd))
func LoadFromFlags(ctx context.Context, cmd *cobra.Command, cfg Config, fallback ...Source) error {
	// 1. --config flag or CONFIG_FILE env.
	if path := flagOrEnv(cmd, "config", "CONFIG_FILE"); path != "" {
		return Load(ctx, cfg, NewFileSource(path))
	}

	// 2. Fallback sources.
	for _, src := range fallback {
		if src != nil {
			return Load(ctx, cfg, src)
		}
	}

	// 3. Default file locations.
	for _, p := range []string{"config.yaml", "config.yml", "config.json"} {
		if fileExists(p) {
			return Load(ctx, cfg, NewFileSource(p))
		}
	}
	return fmt.Errorf("no config source: use --config flag, CONFIG_FILE env, or place config.yaml in working directory")
}

// FlagOrEnv returns the flag value if set, otherwise the environment variable.
func FlagOrEnv(cmd *cobra.Command, flag, envKey string) string {
	return flagOrEnv(cmd, flag, envKey)
}

func flagOrEnv(cmd *cobra.Command, flag, envKey string) string {
	if v, _ := cmd.Flags().GetString(flag); v != "" {
		return v
	}
	return os.Getenv(envKey)
}
