package agentic

import (
	"context"
	"fmt"

	sharedv1 "github.com/pinealctx/nexus-proto/gen/go/shared/v1"

	"github.com/pinealctx/nexus-x/adaptivecard"
)

// Channel is the outbound interface for Nexus IM messaging.
// It uses proto types directly for full access to Nexus capabilities.
type Channel interface {
	// SendMessage sends a message to a conversation.
	SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResult, error)

	// EditMessage edits a previously sent message.
	EditMessage(ctx context.Context, conversationID, messageID int64, newBody *sharedv1.MessageBody) error

	// RecallMessage recalls a sent message.
	RecallMessage(ctx context.Context, conversationID, messageID int64) error

	// AnswerCardAction responds to a card action callback.
	AnswerCardAction(ctx context.Context, conversationID, messageID int64, actionID string, text string, showAlert bool) error
}

// StreamingChannel extends Channel with streaming message support.
type StreamingChannel interface {
	Channel

	// StartStream begins a streaming message in the given conversation.
	StartStream(ctx context.Context, conversationID int64) (StreamWriter, error)
}

// StreamWriter writes incremental output to a streaming message.
type StreamWriter interface {
	// Push sends a text delta to the stream.
	Push(ctx context.Context, delta string) error

	// End finalizes the stream with the complete text.
	End(ctx context.Context, fullText string) error

	// Error terminates the stream with an error message.
	Error(ctx context.Context, errMsg string) error
}

// SendMessageRequest holds the full parameters for sending a message.
type SendMessageRequest struct {
	ConversationID   int64
	Body             *sharedv1.MessageBody
	ReplyToMessageID *int64
	ClientMessageID  int64
}

// SendMessageResult holds the result of a sent message.
type SendMessageResult struct {
	MessageID int64
	CreatedAt int64
}

// --- Send options ---

// SendOption configures a SendMessageRequest.
type SendOption func(*SendMessageRequest)

// WithReplyTo sets the message to reply to.
func WithReplyTo(messageID int64) SendOption {
	return func(r *SendMessageRequest) { r.ReplyToMessageID = &messageID }
}

// WithClientMessageID sets a client-generated message ID for idempotency.
func WithClientMessageID(id int64) SendOption {
	return func(r *SendMessageRequest) { r.ClientMessageID = id }
}

// WithMentions appends @mention entities to the message body.
// Only applies to TEXT and MARKDOWN message types.
func WithMentions(mentions ...*sharedv1.MessageEntity) SendOption {
	return func(r *SendMessageRequest) {
		if r.Body == nil {
			return
		}
		switch c := r.Body.Content.(type) {
		case *sharedv1.MessageBody_Text:
			if c.Text != nil {
				c.Text.Entities = append(c.Text.Entities, mentions...)
			}
		case *sharedv1.MessageBody_Markdown:
			if c.Markdown != nil {
				c.Markdown.Entities = append(c.Markdown.Entities, mentions...)
			}
		}
	}
}

// --- Convenience send functions ---

// SendText sends a Markdown message to a conversation.
func SendText(ctx context.Context, ch Channel, convID int64, text string, opts ...SendOption) error {
	req := &SendMessageRequest{
		ConversationID: convID,
		Body: &sharedv1.MessageBody{
			Type: sharedv1.MessageType_MESSAGE_TYPE_MARKDOWN,
			Content: &sharedv1.MessageBody_Markdown{
				Markdown: &sharedv1.MarkdownContent{RawMarkdown: text},
			},
		},
	}
	for _, opt := range opts {
		opt(req)
	}
	_, err := ch.SendMessage(ctx, req)
	return err
}

// SendCard sends an Adaptive Card message to a conversation.
// Accepts *adaptivecard.Card, string (raw JSON), or []byte (raw JSON).
func SendCard(ctx context.Context, ch Channel, convID int64, card any, opts ...SendOption) error {
	cardJSON, fallback, err := marshalCardAny(card)
	if err != nil {
		return fmt.Errorf("marshal card: %w", err)
	}
	return sendCardBody(ctx, ch, convID, cardJSON, fallback, opts)
}

// SendAdaptiveCard sends a strongly-typed Adaptive Card message.
// Preferred over SendCard for compile-time type safety.
func SendAdaptiveCard(ctx context.Context, ch Channel, convID int64, card *adaptivecard.Card, opts ...SendOption) error {
	cardJSON, err := card.JSON()
	if err != nil {
		return fmt.Errorf("marshal card: %w", err)
	}
	return sendCardBody(ctx, ch, convID, cardJSON, card.FallbackText, opts)
}

func sendCardBody(ctx context.Context, ch Channel, convID int64, cardJSON, fallback string, opts []SendOption) error {
	req := &SendMessageRequest{
		ConversationID: convID,
		Body: &sharedv1.MessageBody{
			Type: sharedv1.MessageType_MESSAGE_TYPE_CARD,
			Content: &sharedv1.MessageBody_Card{
				Card: &sharedv1.CardContent{
					CardJson:     cardJSON,
					FallbackText: fallback,
				},
			},
		},
	}
	for _, opt := range opts {
		opt(req)
	}
	_, err := ch.SendMessage(ctx, req)
	return err
}

// EditCard edits a message to a new Adaptive Card.
// Accepts *adaptivecard.Card, string (raw JSON), or []byte (raw JSON).
func EditCard(ctx context.Context, ch Channel, convID, msgID int64, card any) error {
	cardJSON, fallback, err := marshalCardAny(card)
	if err != nil {
		return fmt.Errorf("marshal card: %w", err)
	}
	return editCardBody(ctx, ch, convID, msgID, cardJSON, fallback)
}

// EditAdaptiveCard edits a message to a strongly-typed Adaptive Card.
func EditAdaptiveCard(ctx context.Context, ch Channel, convID, msgID int64, card *adaptivecard.Card) error {
	cardJSON, err := card.JSON()
	if err != nil {
		return fmt.Errorf("marshal card: %w", err)
	}
	return editCardBody(ctx, ch, convID, msgID, cardJSON, card.FallbackText)
}

func editCardBody(ctx context.Context, ch Channel, convID, msgID int64, cardJSON, fallback string) error {
	return ch.EditMessage(ctx, convID, msgID, &sharedv1.MessageBody{
		Type: sharedv1.MessageType_MESSAGE_TYPE_CARD,
		Content: &sharedv1.MessageBody_Card{
			Card: &sharedv1.CardContent{
				CardJson:     cardJSON,
				FallbackText: fallback,
			},
		},
	})
}

// marshalCardAny converts a card value to JSON string + fallback text.
func marshalCardAny(card any) (cardJSON string, fallback string, err error) {
	switch v := card.(type) {
	case *adaptivecard.Card:
		s, e := v.JSON()
		if e != nil {
			return "", "", e
		}
		return s, v.FallbackText, nil
	case string:
		return v, "", nil
	case []byte:
		return string(v), "", nil
	default:
		return "", "", fmt.Errorf("unsupported card type %T; use *adaptivecard.Card, string, or []byte", card)
	}
}
