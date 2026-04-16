// Package client provides the Nexus IM client SDK for Agent developers.
// This is the ONLY package in nexus-x that imports nexus-proto service clients.
// It implements agentic.Channel and agentic.StreamingChannel.
//
// For advanced use cases, access the underlying Connect RPC service clients
// directly via the Services() method.
package client

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/pinealctx/nexus-x/nxlog"
	"go.uber.org/zap"

	"connectrpc.com/connect"
	apiv1 "github.com/pinealctx/nexus-proto/gen/go/api/v1"
	"github.com/pinealctx/nexus-proto/gen/go/api/v1/apiv1connect"
	sharedv1 "github.com/pinealctx/nexus-proto/gen/go/shared/v1"

	"github.com/pinealctx/nexus-x/agentic"
)

// Client is the Nexus IM client. It implements agentic.Channel and
// agentic.StreamingChannel for outbound messaging, and provides methods
// for inbound message reception via webhook or WebSocket.
type Client struct {
	token      string
	secretKey  string
	serverAddr string
	services   Services

	selfMu       sync.Mutex
	selfID       int32
	selfResolved bool
}

// Services holds all Connect RPC service clients.
type Services struct {
	Messages      apiv1connect.MessageServiceClient
	Auth          apiv1connect.AuthServiceClient
	Users         apiv1connect.UserServiceClient
	Conversations apiv1connect.ConversationServiceClient
	Contacts      apiv1connect.ContactServiceClient
	Groups        apiv1connect.GroupServiceClient
	Media         apiv1connect.MediaServiceClient
	Agents        apiv1connect.AgentServiceClient
	Push          apiv1connect.PushServiceClient
	Sync          apiv1connect.SyncServiceClient
}

var (
	_ agentic.Channel          = (*Client)(nil)
	_ agentic.StreamingChannel = (*Client)(nil)
)

// Option configures the Client.
type Option func(*options)

type options struct {
	secretKey  string
	httpClient *http.Client
}

// WithSecretKey sets the HMAC secret for webhook verification and Mini App initData.
func WithSecretKey(key string) Option {
	return func(o *options) { o.secretKey = key }
}

// WithHTTPClient sets a custom HTTP client for Connect RPC calls.
func WithHTTPClient(c *http.Client) Option {
	return func(o *options) { o.httpClient = c }
}

// New creates a Nexus IM client.
func New(token, serverAddr string, opts ...Option) *Client {
	o := &options{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(o)
	}

	interceptor := bearerInterceptor(token)
	connOpts := []connect.ClientOption{
		connect.WithInterceptors(interceptor),
	}

	return &Client{
		token:      token,
		secretKey:  o.secretKey,
		serverAddr: serverAddr,
		services: Services{
			Messages:      apiv1connect.NewMessageServiceClient(o.httpClient, serverAddr, connOpts...),
			Auth:          apiv1connect.NewAuthServiceClient(o.httpClient, serverAddr, connOpts...),
			Users:         apiv1connect.NewUserServiceClient(o.httpClient, serverAddr, connOpts...),
			Conversations: apiv1connect.NewConversationServiceClient(o.httpClient, serverAddr, connOpts...),
			Contacts:      apiv1connect.NewContactServiceClient(o.httpClient, serverAddr, connOpts...),
			Groups:        apiv1connect.NewGroupServiceClient(o.httpClient, serverAddr, connOpts...),
			Media:         apiv1connect.NewMediaServiceClient(o.httpClient, serverAddr, connOpts...),
			Agents:        apiv1connect.NewAgentServiceClient(o.httpClient, serverAddr, connOpts...),
			Push:          apiv1connect.NewPushServiceClient(o.httpClient, serverAddr, connOpts...),
			Sync:          apiv1connect.NewSyncServiceClient(o.httpClient, serverAddr, connOpts...),
		},
	}
}

// Services returns the underlying Connect RPC service clients.
func (c *Client) Services() *Services {
	return &c.services
}

// SelfUserID returns the agent's own user ID, fetching it lazily via
// GetProfile on first call. Concurrent-safe.
func (c *Client) SelfUserID(ctx context.Context) (int32, error) {
	c.selfMu.Lock()
	defer c.selfMu.Unlock()

	if c.selfResolved {
		return c.selfID, nil
	}

	resp, err := c.services.Users.GetProfile(ctx, connect.NewRequest(&apiv1.GetProfileRequest{}))
	if err != nil {
		return 0, fmt.Errorf("GetProfile: %w", err)
	}
	c.selfID = resp.Msg.GetProfile().GetUserId()
	c.selfResolved = true
	nxlog.Info("resolved self user ID", zap.Int32("user_id", c.selfID))
	return c.selfID, nil
}

func (c *Client) mustSelfID() int32 {
	return c.selfID
}

// --- agentic.Channel implementation ---

// SendMessage implements agentic.Channel.
func (c *Client) SendMessage(ctx context.Context, req *agentic.SendMessageRequest) (*agentic.SendMessageResult, error) {
	protoReq := &apiv1.SendMessageRequest{
		ConversationId:  req.ConversationID,
		Body:            req.Body,
		ClientMessageId: req.ClientMessageID,
	}
	if req.ReplyToMessageID != nil {
		protoReq.ReplyToMessageId = req.ReplyToMessageID
	}

	resp, err := c.services.Messages.SendMessage(ctx, connect.NewRequest(protoReq))
	if err != nil {
		return nil, err
	}
	return &agentic.SendMessageResult{
		MessageID: resp.Msg.GetMessageId(),
		CreatedAt: resp.Msg.GetCreatedAt(),
	}, nil
}

// EditMessage implements agentic.Channel.
func (c *Client) EditMessage(ctx context.Context, conversationID, messageID int64, newBody *sharedv1.MessageBody) error {
	_, err := c.services.Messages.EditMessage(ctx, connect.NewRequest(&apiv1.EditMessageRequest{
		ConversationId: conversationID,
		MessageId:      messageID,
		NewBody:        newBody,
	}))
	return err
}

// RecallMessage implements agentic.Channel.
func (c *Client) RecallMessage(ctx context.Context, conversationID, messageID int64) error {
	_, err := c.services.Messages.RecallMessage(ctx, connect.NewRequest(&apiv1.RecallMessageRequest{
		ConversationId: conversationID,
		MessageId:      messageID,
	}))
	return err
}

// AnswerCardAction implements agentic.Channel.
func (c *Client) AnswerCardAction(ctx context.Context, conversationID, messageID int64, actionID string, text string, showAlert bool) error {
	req := connect.NewRequest(&apiv1.AnswerCardActionRequest{
		ConversationId: conversationID,
		MessageId:      messageID,
		ActionId:       actionID,
		Text:           &text,
		ShowAlert:      showAlert,
	})
	_, err := c.services.Messages.AnswerCardAction(ctx, req)
	return err
}

// --- agentic.StreamingChannel implementation ---

// StartStream implements agentic.StreamingChannel.
func (c *Client) StartStream(ctx context.Context, conversationID int64) (agentic.StreamWriter, error) {
	resp, err := c.services.Messages.SendMessage(ctx, connect.NewRequest(&apiv1.SendMessageRequest{
		ConversationId: conversationID,
		Body: &sharedv1.MessageBody{
			Type: sharedv1.MessageType_MESSAGE_TYPE_STREAM,
		},
	}))
	if err != nil {
		return nil, err
	}
	return &streamWriter{
		client:         c,
		conversationID: conversationID,
		messageID:      resp.Msg.GetMessageId(),
	}, nil
}

type streamWriter struct {
	client         *Client
	conversationID int64
	messageID      int64
	seq            int32
}

func (w *streamWriter) Push(ctx context.Context, delta string) error {
	w.seq++
	_, err := w.client.services.Messages.PushStreamDelta(ctx, connect.NewRequest(&apiv1.PushStreamDeltaRequest{
		ConversationId: w.conversationID,
		MessageId:      w.messageID,
		Seq:            w.seq,
		Delta:          delta,
	}))
	return err
}

func (w *streamWriter) End(ctx context.Context, accumulatedText string) error {
	_, err := w.client.services.Messages.EndStream(ctx, connect.NewRequest(&apiv1.EndStreamRequest{
		ConversationId:  w.conversationID,
		MessageId:       w.messageID,
		AccumulatedText: accumulatedText,
	}))
	return err
}

func (w *streamWriter) Error(ctx context.Context, errMsg string) error {
	_, err := w.client.services.Messages.ErrorStream(ctx, connect.NewRequest(&apiv1.ErrorStreamRequest{
		ConversationId: w.conversationID,
		MessageId:      w.messageID,
		ErrorMessage:   errMsg,
	}))
	return err
}

// --- Gateway URL discovery ---

func (c *Client) fetchGatewayURL(ctx context.Context) (string, error) {
	resp, err := c.services.Auth.GetClientConfig(ctx, connect.NewRequest(&apiv1.GetClientConfigRequest{}))
	if err != nil {
		return "", fmt.Errorf("GetClientConfig: %w", err)
	}
	gw := resp.Msg.GetGateway()
	if gw == nil || gw.GetWsUrl() == "" {
		return "", fmt.Errorf("server returned empty gateway URL")
	}
	return gw.GetWsUrl(), nil
}

// bearerInterceptor adds Bearer token to all Connect RPC requests.
func bearerInterceptor(token string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+token)
			return next(ctx, req)
		}
	}
}
