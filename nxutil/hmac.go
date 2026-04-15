package nxutil

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	// WebhookTimestampDrift is the maximum allowed clock drift for webhook verification.
	WebhookTimestampDrift = 5 * time.Minute

	// InitDataMaxAge is the maximum age of a Mini App initData before it's considered expired.
	InitDataMaxAge = 5 * time.Minute

	// InitDataHMACKey is the HMAC key prefix for Mini App initData signing.
	InitDataHMACKey = "NexusMiniApp"
)

// --- Webhook HMAC ---

// WebhookSign computes the HMAC-SHA256 signature for a webhook payload.
// Returns the signature in "sha256={hex}" format.
func WebhookSign(secretKey, timestamp string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(body)
	return "sha256=" + fmt.Sprintf("%x", mac.Sum(nil))
}

// WebhookVerify verifies a webhook signature against the expected HMAC.
// Returns true if the signature is valid and the timestamp is within drift.
func WebhookVerify(secretKey, signature, timestamp string, body []byte) bool {
	if signature == "" || timestamp == "" {
		return false
	}

	tsInt, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	drift := time.Since(time.Unix(tsInt, 0))
	if drift < 0 {
		drift = -drift
	}
	if drift > WebhookTimestampDrift {
		return false
	}

	expected := WebhookSign(secretKey, timestamp, body)
	return hmac.Equal([]byte(signature), []byte(expected))
}

// --- Mini App initData ---

// InitDataSign signs Mini App initData parameters using the agent's secret_key.
// Returns the hex-encoded HMAC-SHA256 hash.
//
// Algorithm (D14 §5.4):
//  1. Sort params by key alphabetically.
//  2. Join as "key=value" lines separated by "\n".
//  3. HMAC-SHA256(key=HMAC-SHA256("NexusMiniApp", secret_key), data=sorted_params).
func InitDataSign(secretKey string, params url.Values) string {
	var keys []string
	for k := range params {
		if k != "hash" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var lines []string
	for _, k := range keys {
		lines = append(lines, k+"="+params.Get(k))
	}
	dataStr := strings.Join(lines, "\n")

	secretMAC := hmac.New(sha256.New, []byte(InitDataHMACKey))
	secretMAC.Write([]byte(secretKey))
	secretHash := secretMAC.Sum(nil)

	dataMAC := hmac.New(sha256.New, secretHash)
	dataMAC.Write([]byte(dataStr))
	return fmt.Sprintf("%x", dataMAC.Sum(nil))
}

// InitDataVerify verifies a Mini App initData string using the agent's secret_key.
// Returns the parsed parameters (excluding hash) on success.
func InitDataVerify(secretKey, initData string) (url.Values, error) {
	params, err := url.ParseQuery(initData)
	if err != nil {
		return nil, fmt.Errorf("parse initData: %w", err)
	}

	hash := params.Get("hash")
	if hash == "" {
		return nil, fmt.Errorf("missing hash in initData")
	}

	expected := InitDataSign(secretKey, params)
	if !hmac.Equal([]byte(hash), []byte(expected)) {
		return nil, fmt.Errorf("invalid initData signature")
	}

	authDateStr := params.Get("auth_date")
	if authDateStr == "" {
		return nil, fmt.Errorf("missing auth_date in initData")
	}
	authDate, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid auth_date: %w", err)
	}
	drift := time.Since(time.Unix(authDate, 0))
	if drift < 0 {
		drift = -drift
	}
	if drift > InitDataMaxAge {
		return nil, fmt.Errorf("initData expired (auth_date drift: %s)", drift)
	}

	params.Del("hash")
	return params, nil
}
