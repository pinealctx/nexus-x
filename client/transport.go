package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"

	"github.com/pinealctx/nexus-x/agentic"
)

// --- Webhook ---

// WebhookHandler returns an http.Handler that verifies and parses Nexus
// webhook callbacks. Parsed messages are delivered to the handler asynchronously.
// Self-messages (from the agent itself) are filtered out automatically.
func (c *Client) WebhookHandler(handler agentic.Handler) http.Handler {
	return &webhookHandler{client: c, handler: handler}
}

type webhookHandler struct {
	client       *Client
	handler      agentic.Handler
	mu           sync.Mutex
	selfResolved bool
}

func (h *webhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Lazily resolve selfID on first request.
	h.mu.Lock()
	if !h.selfResolved {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		if _, err := h.client.SelfUserID(ctx); err != nil {
			slog.Error("failed to resolve self user ID", "err", err)
		} else {
			h.selfResolved = true
		}
	}
	h.mu.Unlock()

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		slog.Warn("webhook read body failed", "err", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if !h.client.verifyWebhook(r, body) {
		slog.Warn("webhook signature verification failed", "remote_addr", r.RemoteAddr)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	r.Body = io.NopCloser(bytes.NewReader(body))

	update, err := h.client.parseWebhook(body)
	if err != nil {
		slog.Error("webhook parse failed", "err", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

	if update == nil {
		return
	}

	// Filter self-messages.
	if update.UserID == h.client.mustSelfID() {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		if err := h.handler(ctx, update); err != nil {
			slog.Error("handler failed", "err", err)
		}
	}()
}

// --- WebSocket ---

// WSOption configures the WebSocket connection.
type WSOption func(*wsConfig)

type wsConfig struct {
	gatewayURL  string
	maxInterval time.Duration
}

// WithGatewayURL overrides the auto-discovered WebSocket gateway URL.
func WithGatewayURL(url string) WSOption {
	return func(c *wsConfig) { c.gatewayURL = url }
}

// WithReconnectMaxInterval sets the maximum backoff interval for reconnection.
func WithReconnectMaxInterval(d time.Duration) WSOption {
	return func(c *wsConfig) { c.maxInterval = d }
}

// ConnectWebSocket establishes a persistent WebSocket connection to the
// Nexus gateway and delivers parsed messages to the handler.
// Reconnects with exponential backoff on disconnect.
// Blocks until ctx is cancelled.
func (c *Client) ConnectWebSocket(ctx context.Context, handler agentic.Handler, opts ...WSOption) error {
	if _, err := c.SelfUserID(ctx); err != nil {
		return fmt.Errorf("resolve self user ID: %w", err)
	}

	wsCfg := wsConfig{maxInterval: 60 * time.Second}
	for _, opt := range opts {
		opt(&wsCfg)
	}

	if wsCfg.gatewayURL == "" {
		url, err := c.fetchGatewayURL(ctx)
		if err != nil {
			return fmt.Errorf("discover gateway URL: %w", err)
		}
		wsCfg.gatewayURL = url
	}

	ws := &wsClient{client: c, handler: handler, cfg: wsCfg}
	return ws.connect(ctx)
}

type wsClient struct {
	client  *Client
	handler agentic.Handler
	cfg     wsConfig
	mu      sync.Mutex
	conn    *websocket.Conn
}

func (w *wsClient) connect(ctx context.Context) error {
	backoff := time.Second
	for {
		slog.Info("ws connecting", "gateway_url", w.cfg.gatewayURL)
		if err := w.dial(ctx); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			slog.Warn("ws connect failed, retrying", "err", err, "retry_in", backoff)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
			backoff = min(backoff*2, w.cfg.maxInterval)
			continue
		}
		backoff = time.Second
	}
}

func (w *wsClient) dial(ctx context.Context) error {
	opts := &websocket.DialOptions{
		HTTPHeader: http.Header{
			"Authorization": []string{"Bearer " + w.client.token},
		},
	}
	conn, _, err := websocket.Dial(ctx, w.cfg.gatewayURL, opts)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	w.mu.Lock()
	w.conn = conn
	w.mu.Unlock()
	slog.Info("ws connected")
	defer func() {
		_ = conn.CloseNow()
		w.mu.Lock()
		w.conn = nil
		w.mu.Unlock()
		slog.Info("ws disconnected")
	}()

	return w.readLoop(ctx, conn)
}

func (w *wsClient) readLoop(ctx context.Context, conn *websocket.Conn) error {
	selfID := w.client.mustSelfID()
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return err
		}

		update, err := w.client.parseWSFrame(data)
		if err != nil {
			slog.Error("ws parse frame failed", "err", err)
			continue
		}
		if update == nil {
			continue
		}

		if update.UserID == selfID {
			continue
		}

		go func() {
			dispatchCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			if err := w.handler(dispatchCtx, update); err != nil {
				slog.Error("handler failed (ws)", "err", err)
			}
		}()
	}
}
