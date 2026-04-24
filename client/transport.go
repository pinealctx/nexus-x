package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"

	apiv1 "github.com/pinealctx/nexus-proto/gen/go/api/v1"
	"github.com/pinealctx/nexus-x/agentic"
	"github.com/pinealctx/nexus-x/nxlog"
	"github.com/pinealctx/nexus-x/nxproto"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const (
	heartbeatInterval = 30 * time.Second
	authTimeout       = 30 * time.Second
	writeTimeout      = 10 * time.Second
	missedPongLimit   = 3
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
			nxlog.Error("failed to resolve self user ID", zap.Error(err))
		} else {
			h.selfResolved = true
		}
	}
	h.mu.Unlock()

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		nxlog.Warn("webhook read body failed", zap.Error(err))
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if !h.client.verifyWebhook(r, body) {
		nxlog.Warn("webhook signature verification failed", zap.String("remote_addr", r.RemoteAddr))
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	r.Body = io.NopCloser(bytes.NewReader(body))

	update, err := h.client.parseWebhook(body)
	if err != nil {
		nxlog.Error("webhook parse failed", zap.Error(err))
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
			nxlog.Error("handler failed", zap.Error(err))
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

	reqID       atomic.Int64
	missedPongs int
}

func (w *wsClient) connect(ctx context.Context) error {
	backoff := time.Second
	for {
		nxlog.Info("ws connecting", zap.String("gateway_url", w.cfg.gatewayURL))
		if err := w.dial(ctx); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			nxlog.Warn("ws connect failed, retrying", zap.Error(err), zap.Duration("retry_in", backoff))
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
	conn, _, err := websocket.Dial(ctx, w.cfg.gatewayURL, nil)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	// Per-connection context: cancelled when this connection closes,
	// ensuring the heartbeat goroutine exits promptly on reconnect.
	connCtx, connCancel := context.WithCancel(ctx)

	w.mu.Lock()
	w.conn = conn
	w.mu.Unlock()

	nxlog.Info("ws connected")

	defer func() {
		connCancel()
		_ = conn.CloseNow()
		w.mu.Lock()
		w.conn = nil
		w.mu.Unlock()
		nxlog.Info("ws disconnected")
	}()

	// Authenticate before entering read loop.
	if err := w.authenticate(connCtx, conn); err != nil {
		return fmt.Errorf("auth: %w", err)
	}

	// Start heartbeat goroutine.
	go w.heartbeatLoop(connCtx, conn)

	return w.readLoop(connCtx, conn)
}

// authenticate sends an AUTH_REQUEST frame and waits for AUTH_RESPONSE.
func (w *wsClient) authenticate(ctx context.Context, conn *websocket.Conn) error {
	authCtx, cancel := context.WithTimeout(ctx, authTimeout)
	defer cancel()

	reqID := w.reqID.Add(1)
	frame := &apiv1.ClientFrame{
		RequestId: reqID,
		Type:      apiv1.ClientFrameType_CLIENT_FRAME_TYPE_AUTH_REQUEST,
		Payload: &apiv1.ClientFrame_AuthRequest{
			AuthRequest: &apiv1.AuthRequest{
				Token: w.client.token,
			},
		},
	}

	if err := w.writeFrame(authCtx, conn, frame); err != nil {
		return fmt.Errorf("write auth: %w", err)
	}

	_, respData, err := conn.Read(authCtx)
	if err != nil {
		nxlog.Warn("ws auth read failed", zap.Error(err))
		return fmt.Errorf("read auth response: %w", err)
	}

	var resp apiv1.ServerFrame
	if err := proto.Unmarshal(respData, &resp); err != nil {
		return fmt.Errorf("unmarshal auth response: %w", err)
	}

	authResp := resp.GetAuthResponse()
	if authResp == nil || !authResp.GetSuccess() {
		msg := "auth failed"
		if authResp != nil && authResp.GetError() != nil {
			if m, ok := authResp.GetError().GetMetadata()["message"]; ok && m != "" {
				msg = m
			} else if authResp.GetError().GetErrorName() != "" {
				msg = authResp.GetError().GetErrorName()
			}
		}
		nxlog.Warn("ws auth failed", zap.String("message", msg))
		return fmt.Errorf("%s", msg)
	}

	nxlog.Info("ws auth success", zap.Int32("user_id", authResp.GetUserId()))

	w.mu.Lock()
	w.missedPongs = 0
	w.mu.Unlock()

	return nil
}

// heartbeatLoop sends periodic HEARTBEAT_PING frames and monitors pong responses.
func (w *wsClient) heartbeatLoop(ctx context.Context, conn *websocket.Conn) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.mu.Lock()
			w.missedPongs++
			missed := w.missedPongs
			w.mu.Unlock()

			if missed > missedPongLimit {
				nxlog.Warn("ws missed pong limit exceeded, closing connection", zap.Int("missed", missed))
				_ = conn.Close(websocket.StatusGoingAway, "pong timeout")
				return
			}

			reqID := w.reqID.Add(1)
			frame := &apiv1.ClientFrame{
				RequestId: reqID,
				Type:      apiv1.ClientFrameType_CLIENT_FRAME_TYPE_HEARTBEAT_PING,
				Payload: &apiv1.ClientFrame_HeartbeatPing{
					HeartbeatPing: &apiv1.HeartbeatPing{},
				},
			}

			writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := w.writeFrame(writeCtx, conn, frame)
			cancel()
			if err != nil {
				nxlog.Debug("ws heartbeat send failed", zap.Error(err))
			}
		}
	}
}

func (w *wsClient) readLoop(ctx context.Context, conn *websocket.Conn) error {
	selfID := w.client.mustSelfID()
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return err
		}

		var frame apiv1.ServerFrame
		if err := proto.Unmarshal(data, &frame); err != nil {
			nxlog.Warn("ws unmarshal failed", zap.Error(err))
			continue
		}

		nxlog.Debug("ws recv", nxproto.ProtoJSON("frame", &frame))

		switch frame.Type {
		case apiv1.ServerFrameType_SERVER_FRAME_TYPE_UPDATE:
			updateFrame, ok := frame.Payload.(*apiv1.ServerFrame_Update)
			if !ok || updateFrame == nil {
				continue
			}
			update := w.client.convertUpdate(updateFrame.Update)
			if update == nil || update.UserID == selfID {
				continue
			}
			w.dispatchUpdate(update)

		case apiv1.ServerFrameType_SERVER_FRAME_TYPE_HEARTBEAT_PONG:
			w.mu.Lock()
			w.missedPongs = 0
			w.mu.Unlock()

		case apiv1.ServerFrameType_SERVER_FRAME_TYPE_ERROR:
			ef := frame.GetError()
			if ef != nil && ef.GetFatal() {
				nxlog.Warn("ws recv fatal error", zap.Int64("req_id", frame.RequestId))
			}
		}
	}
}

// dispatchUpdate dispatches a parsed update to the handler in a goroutine.
func (w *wsClient) dispatchUpdate(update *agentic.IncomingUpdate) {
	go func() {
		dispatchCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		if err := w.handler(dispatchCtx, update); err != nil {
			nxlog.Error("handler failed (ws)", zap.Error(err))
		}
	}()
}

// writeFrame marshals and writes a ClientFrame to the connection.
// The caller must handle any needed timeout via the ctx.
func (w *wsClient) writeFrame(ctx context.Context, conn *websocket.Conn, frame *apiv1.ClientFrame) error {
	nxlog.Debug("ws send", nxproto.ProtoJSON("frame", frame))
	data, err := proto.Marshal(frame)
	if err != nil {
		return fmt.Errorf("marshal frame: %w", err)
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	return conn.Write(ctx, websocket.MessageBinary, data)
}
