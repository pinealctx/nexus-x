package client

import (
	"fmt"
	"net/url"

	"github.com/pinealctx/nexus-x/nxutil"
)

// VerifyInitData verifies a Mini App initData string using the agent's secret_key.
// Returns the parsed parameters (excluding hash) on success.
func (c *Client) VerifyInitData(initData string) (url.Values, error) {
	if c.secretKey == "" {
		return nil, fmt.Errorf("secret_key not configured; use client.WithSecretKey()")
	}
	return nxutil.InitDataVerify(c.secretKey, initData)
}
