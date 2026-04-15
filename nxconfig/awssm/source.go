// Package awssm provides an nxconfig.Source that fetches configuration
// from AWS Secrets Manager. Import this package only when you need AWS
// Secrets Manager support — it pulls in the AWS SDK v2 dependency.
package awssm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/spf13/cobra"

	"github.com/pinealctx/nexus-x/nxconfig"
)

const defaultTimeout = 15 * time.Second

// Source reads configuration from AWS Secrets Manager.
// The secret value is expected to be a YAML or JSON string stored as
// plaintext in Secrets Manager.
type Source struct {
	Region     string
	SecretName string
	Timeout    time.Duration
}

// NewSource creates an AWS Secrets Manager source.
func NewSource(region, secretName string) *Source {
	return &Source{
		Region:     region,
		SecretName: secretName,
		Timeout:    defaultTimeout,
	}
}

// Load fetches the secret value from AWS Secrets Manager.
func (s *Source) Load(ctx context.Context) ([]byte, error) {
	timeout := s.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}
	fetchCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cfg, err := awsconfig.LoadDefaultConfig(fetchCtx, awsconfig.WithRegion(s.Region))
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	client := secretsmanager.NewFromConfig(cfg)
	out, err := client.GetSecretValue(fetchCtx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(s.SecretName),
	})
	if err != nil {
		return nil, fmt.Errorf("get secret %q: %w", s.SecretName, err)
	}

	if out.SecretString != nil {
		return []byte(*out.SecretString), nil
	}

	// SecretBinary: try to parse as JSON-encoded string.
	if out.SecretBinary != nil {
		var str string
		if jsonErr := json.Unmarshal(out.SecretBinary, &str); jsonErr != nil {
			// Not JSON-wrapped, use raw bytes.
			return out.SecretBinary, nil //nolint:nilerr
		}
		return []byte(str), nil
	}

	return nil, fmt.Errorf("secret %q has no value", s.SecretName)
}

// Ensure Source implements nxconfig.Source.
var _ nxconfig.Source = (*Source)(nil)

// RegisterFlags adds --region and --secret-name flags to a cobra command,
// along with the standard --config flag from nxconfig.
func RegisterFlags(cmd *cobra.Command) {
	nxconfig.RegisterFlags(cmd)
	cmd.Flags().StringP("region", "r", "", "AWS region for Secrets Manager (env: REGION)")
	cmd.Flags().StringP("secret-name", "s", "", "AWS Secrets Manager secret name (env: SECRET_NAME)")
}

// SourceFromFlags returns an AWS Secrets Manager source if --region and
// --secret-name are set (via flags or env), otherwise nil.
// Pass as a fallback to nxconfig.LoadFromFlags:
//
//	awssm.RegisterFlags(cmd)
//	nxconfig.LoadFromFlags(ctx, cmd, &cfg, awssm.SourceFromFlags(cmd))
func SourceFromFlags(cmd *cobra.Command) nxconfig.Source {
	region := nxconfig.FlagOrEnv(cmd, "region", "REGION")
	secret := nxconfig.FlagOrEnv(cmd, "secret-name", "SECRET_NAME")
	if region != "" && secret != "" {
		return NewSource(region, secret)
	}
	return nil
}
