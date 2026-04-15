package nxconfig

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
)

// LogConfig holds logging parameters.
type LogConfig struct {
	Level string `yaml:"level" json:"level"` // debug, info, warn, error
}

// TLSConfig holds optional mTLS certificate material.
// All fields are PEM-encoded text (not file paths) so that
// configuration can be loaded uniformly from YAML, env vars, or secret
// managers without touching the filesystem.
type TLSConfig struct {
	CertPEM string `yaml:"cert_pem" json:"cert_pem"`
	KeyPEM  string `yaml:"key_pem" json:"key_pem"`
	CAPEM   string `yaml:"ca_pem" json:"ca_pem"`
}

// Enabled reports whether TLS material is configured.
func (c *TLSConfig) Enabled() bool {
	return c.CertPEM != "" && c.KeyPEM != ""
}

// ServerTLSConfig builds a *tls.Config suitable for a gRPC/HTTP server.
// When CAPEM is provided, client certificate verification is enabled (mTLS).
func (c *TLSConfig) ServerTLSConfig() (*tls.Config, error) {
	cert, err := tls.X509KeyPair([]byte(c.CertPEM), []byte(c.KeyPEM))
	if err != nil {
		return nil, fmt.Errorf("parse server certificate: %w", err)
	}

	tc := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
	}

	if c.CAPEM != "" {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM([]byte(c.CAPEM)) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tc.ClientCAs = pool
		tc.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return tc, nil
}

// ClientTLSConfig builds a *tls.Config suitable for a gRPC/HTTP client.
func (c *TLSConfig) ClientTLSConfig() (*tls.Config, error) {
	tc := &tls.Config{
		MinVersion: tls.VersionTLS13,
	}

	if c.CAPEM != "" {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM([]byte(c.CAPEM)) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tc.RootCAs = pool
	}

	if c.CertPEM != "" && c.KeyPEM != "" {
		cert, err := tls.X509KeyPair([]byte(c.CertPEM), []byte(c.KeyPEM))
		if err != nil {
			return nil, fmt.Errorf("parse client certificate: %w", err)
		}
		tc.Certificates = []tls.Certificate{cert}
	}

	return tc, nil
}

// ListenConfig holds HTTP/gRPC listener parameters.
type ListenConfig struct {
	Host string `yaml:"host" json:"host"`
	Port int    `yaml:"port" json:"port"`
}

// Addr returns "host:port" string.
func (c *ListenConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// CORSConfig holds Cross-Origin Resource Sharing parameters.
type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins" json:"allowed_origins"`
}
