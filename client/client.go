// Package client provides the Nexus IM client SDK for Agent developers.
// This is the ONLY package in nexus-x that imports nexus-proto.
// It implements agentic.Channel and agentic.StreamingChannel.
package client

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

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

	messages      apiv1connect.MessageServiceClient
	auth          apiv1connect.AuthServiceClient
	users         apiv1connect.UserServiceClient
	conversations apiv1connect.ConversationServiceClient
	contacts      apiv1connect.ContactServiceClient
	groups        apiv1connect.GroupServiceClient
	media         apiv1connect.MediaServiceClient

	selfMu       sync.Mutex
	selfID       int32
	selfResolved bool
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
//
//	c := client.New("nxa_xxx", "https://nexus.example.com")
//	c := client.New("nxa_xxx", "https://nexus.example.com",
//	    client.WithSecretKey("sk_xxx"),
//	)
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
		token:         token,
		secretKey:     o.secretKey,
		serverAddr:    serverAddr,
		messages:      apiv1connect.NewMessageServiceClient(o.httpClient, serverAddr, connOpts...),
		auth:          apiv1connect.NewAuthServiceClient(o.httpClient, serverAddr, connOpts...),
		users:         apiv1connect.NewUserServiceClient(o.httpClient, serverAddr, connOpts...),
		conversations: apiv1connect.NewConversationServiceClient(o.httpClient, serverAddr, connOpts...),
		contacts:      apiv1connect.NewContactServiceClient(o.httpClient, serverAddr, connOpts...),
		groups:        apiv1connect.NewGroupServiceClient(o.httpClient, serverAddr, connOpts...),
		media:         apiv1connect.NewMediaServiceClient(o.httpClient, serverAddr, connOpts...),
	}
}

// SelfUserID returns the agent's own user ID, fetching it lazily via
// GetProfile on first call. Concurrent-safe.
func (c *Client) SelfUserID(ctx context.Context) (int32, error) {
	c.selfMu.Lock()
	defer c.selfMu.Unlock()

	if c.selfResolved {
		return c.selfID, nil
	}

	resp, err := c.users.GetProfile(ctx, connect.NewRequest(&apiv1.GetProfileRequest{}))
	if err != nil {
		return 0, fmt.Errorf("GetProfile: %w", err)
	}
	c.selfID = resp.Msg.GetProfile().GetUserId()
	c.selfResolved = true
	slog.Info("resolved self user ID", "user_id", c.selfID)
	return c.selfID, nil
}

func (c *Client) mustSelfID() int32 {
	return c.selfID
}

// --- Outbound (agentic.Channel) ---

// SendText sends a Markdown message to a conversation.
func (c *Client) SendText(ctx context.Context, conversationID int64, text string) error {
	_, err := c.sendMessage(ctx, conversationID, &sharedv1.MessageBody{
		Type: sharedv1.MessageType_MESSAGE_TYPE_MARKDOWN,
		Content: &sharedv1.MessageBody_Markdown{
			Markdown: &sharedv1.MarkdownContent{RawMarkdown: text},
		},
	}, nil)
	return err
}

// SendCard sends a structured card to a conversation.
// Accepts *adaptivecard.Card, []byte (raw JSON), string (raw JSON), or any JSON-serializable value.
func (c *Client) SendCard(ctx context.Context, conversationID int64, card any) error {
	cardJSON, err := marshalCard(card)
	if err != nil {
		return fmt.Errorf("marshal card: %w", err)
	}
	_, err = c.sendMessage(ctx, conversationID, &sharedv1.MessageBody{
		Type: sharedv1.MessageType_MESSAGE_TYPE_CARD,
		Content: &sharedv1.MessageBody_Card{
			Card: &sharedv1.CardContent{
				CardJson: string(cardJSON),
			},
		},
	}, nil)
	return err
}

// AnswerCardAction responds to a card action callback with a toast or alert.
func (c *Client) AnswerCardAction(ctx context.Context, _ int64, actionID string, card any) error {
	text := ""
	if s, ok := card.(string); ok {
		text = s
	}
	req := connect.NewRequest(&apiv1.AnswerCardActionRequest{
		ActionId:  actionID,
		Text:      &text,
		ShowAlert: false,
	})
	_, err := c.messages.AnswerCardAction(ctx, req)
	return err
}

// AnswerCardActionAlert responds to a card action with an alert dialog.
func (c *Client) AnswerCardActionAlert(ctx context.Context, actionID string, text string) error {
	req := connect.NewRequest(&apiv1.AnswerCardActionRequest{
		ActionId:  actionID,
		Text:      &text,
		ShowAlert: true,
	})
	_, err := c.messages.AnswerCardAction(ctx, req)
	return err
}

// SendMessageResult contains the result of a sent message.
type SendMessageResult struct {
	MessageID int64
	CreatedAt int64
}

// SendOption configures a SendMessage call.
type SendOption func(*sendOptions)

type sendOptions struct {
	replyToMessageID *int64
	clientMessageID  int64
}

// WithReplyTo sets the message to reply to.
func WithReplyTo(messageID int64) SendOption {
	return func(o *sendOptions) { o.replyToMessageID = &messageID }
}

// WithClientMessageID sets a client-generated message ID for idempotency.
func WithClientMessageID(id int64) SendOption {
	return func(o *sendOptions) { o.clientMessageID = id }
}

// SendMessage sends a message with full control over the body.
// For most cases, use SendText or SendCard instead.
func (c *Client) SendMessage(ctx context.Context, conversationID int64, body *sharedv1.MessageBody, opts ...SendOption) (*SendMessageResult, error) {
	return c.sendMessage(ctx, conversationID, body, opts)
}

func (c *Client) sendMessage(ctx context.Context, conversationID int64, body *sharedv1.MessageBody, opts []SendOption) (*SendMessageResult, error) {
	o := &sendOptions{}
	for _, opt := range opts {
		opt(o)
	}

	req := &apiv1.SendMessageRequest{
		ConversationId:  conversationID,
		Body:            body,
		ClientMessageId: o.clientMessageID,
	}
	if o.replyToMessageID != nil {
		req.ReplyToMessageId = o.replyToMessageID
	}

	resp, err := c.messages.SendMessage(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return &SendMessageResult{
		MessageID: resp.Msg.GetMessageId(),
		CreatedAt: resp.Msg.GetCreatedAt(),
	}, nil
}

// --- Message operations ---

// EditMessage edits a previously sent message.
func (c *Client) EditMessage(ctx context.Context, conversationID, messageID int64, newBody *sharedv1.MessageBody) error {
	_, err := c.messages.EditMessage(ctx, connect.NewRequest(&apiv1.EditMessageRequest{
		ConversationId: conversationID,
		MessageId:      messageID,
		NewBody:        newBody,
	}))
	return err
}

// EditMessageText edits a message to new Markdown text.
func (c *Client) EditMessageText(ctx context.Context, conversationID, messageID int64, text string) error {
	return c.EditMessage(ctx, conversationID, messageID, &sharedv1.MessageBody{
		Type: sharedv1.MessageType_MESSAGE_TYPE_MARKDOWN,
		Content: &sharedv1.MessageBody_Markdown{
			Markdown: &sharedv1.MarkdownContent{RawMarkdown: text},
		},
	})
}

// EditMessageCard edits a message to a new card.
func (c *Client) EditMessageCard(ctx context.Context, conversationID, messageID int64, card any) error {
	cardJSON, err := marshalCard(card)
	if err != nil {
		return fmt.Errorf("marshal card: %w", err)
	}
	return c.EditMessage(ctx, conversationID, messageID, &sharedv1.MessageBody{
		Type: sharedv1.MessageType_MESSAGE_TYPE_CARD,
		Content: &sharedv1.MessageBody_Card{
			Card: &sharedv1.CardContent{CardJson: string(cardJSON)},
		},
	})
}

// RecallMessage recalls a sent message (visible to all participants).
func (c *Client) RecallMessage(ctx context.Context, conversationID, messageID int64) error {
	_, err := c.messages.RecallMessage(ctx, connect.NewRequest(&apiv1.RecallMessageRequest{
		ConversationId: conversationID,
		MessageId:      messageID,
	}))
	return err
}

// --- Streaming (agentic.StreamingChannel) ---

// StartStream begins a streaming message and returns a StreamWriter.
func (c *Client) StartStream(ctx context.Context, conversationID int64) (agentic.StreamWriter, error) {
	resp, err := c.messages.SendMessage(ctx, connect.NewRequest(&apiv1.SendMessageRequest{
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
	_, err := w.client.messages.PushStreamDelta(ctx, connect.NewRequest(&apiv1.PushStreamDeltaRequest{
		ConversationId: w.conversationID,
		MessageId:      w.messageID,
		Seq:            w.seq,
		Delta:          delta,
	}))
	return err
}

func (w *streamWriter) End(ctx context.Context, accumulatedText string) error {
	_, err := w.client.messages.EndStream(ctx, connect.NewRequest(&apiv1.EndStreamRequest{
		ConversationId:  w.conversationID,
		MessageId:       w.messageID,
		AccumulatedText: accumulatedText,
	}))
	return err
}

func (w *streamWriter) Error(ctx context.Context, errMsg string) error {
	_, err := w.client.messages.ErrorStream(ctx, connect.NewRequest(&apiv1.ErrorStreamRequest{
		ConversationId: w.conversationID,
		MessageId:      w.messageID,
		ErrorMessage:   errMsg,
	}))
	return err
}

// --- Gateway URL discovery ---

func (c *Client) fetchGatewayURL(ctx context.Context) (string, error) {
	resp, err := c.auth.GetClientConfig(ctx, connect.NewRequest(&apiv1.GetClientConfigRequest{}))
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
