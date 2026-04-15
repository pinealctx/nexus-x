package client

import (
	"fmt"
	"net/http"

	apiv1 "github.com/pinealctx/nexus-proto/gen/go/api/v1"
	sharedv1 "github.com/pinealctx/nexus-proto/gen/go/shared/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/pinealctx/nexus-x/agentic"
	"github.com/pinealctx/nexus-x/nxutil"
)

// verifyWebhook verifies the HMAC-SHA256 webhook signature.
func (c *Client) verifyWebhook(r *http.Request, body []byte) bool {
	return nxutil.WebhookVerify(
		c.secretKey,
		r.Header.Get("X-Nexus-Signature"),
		r.Header.Get("X-Nexus-Timestamp"),
		body,
	)
}

// parseWebhook parses a Nexus webhook HTTP request body into an IncomingUpdate.
func (c *Client) parseWebhook(body []byte) (*agentic.IncomingUpdate, error) {
	var update apiv1.Update
	if err := protojson.Unmarshal(body, &update); err != nil {
		return nil, fmt.Errorf("unmarshal update: %w", err)
	}
	return c.convertUpdate(&update), nil
}

// parseWSFrame parses a Nexus WebSocket ServerFrame proto.
func (c *Client) parseWSFrame(data []byte) (*agentic.IncomingUpdate, error) {
	var frame apiv1.ServerFrame
	if err := proto.Unmarshal(data, &frame); err != nil {
		return nil, fmt.Errorf("unmarshal server frame: %w", err)
	}

	updateFrame, ok := frame.Payload.(*apiv1.ServerFrame_Update)
	if !ok {
		return nil, nil // heartbeat, auth response, etc.
	}

	return c.convertUpdate(updateFrame.Update), nil
}

func (c *Client) convertUpdate(update *apiv1.Update) *agentic.IncomingUpdate {
	if update == nil {
		return nil
	}

	switch u := update.Update.(type) {
	case *apiv1.Update_SnUpdate:
		return c.convertSnUpdate(u.SnUpdate)
	case *apiv1.Update_NonSnUpdate:
		return c.convertNonSnUpdate(u.NonSnUpdate)
	}
	return nil
}

func (c *Client) convertSnUpdate(sn *sharedv1.SnUpdate) *agentic.IncomingUpdate {
	if sn == nil {
		return nil
	}
	env, ok := sn.Update.(*sharedv1.SnUpdate_MessageEnvelope)
	if !ok {
		return nil
	}
	msg := env.MessageEnvelope
	if msg == nil {
		return nil
	}

	return &agentic.IncomingUpdate{
		Envelope:       msg,
		UserID:         msg.GetSenderId(),
		ConversationID: msg.GetConversationId(),
		MessageID:      msg.GetMessageId(),
		Text:           extractText(msg.GetBody()),
		Channel:        c,
	}
}

func (c *Client) convertNonSnUpdate(nsn *sharedv1.NonSnUpdate) *agentic.IncomingUpdate {
	if nsn == nil {
		return nil
	}

	ca, ok := nsn.Update.(*sharedv1.NonSnUpdate_CardAction)
	if !ok {
		return nil
	}
	payload := ca.CardAction
	if payload == nil {
		return nil
	}

	return &agentic.IncomingUpdate{
		CardAction:     payload,
		UserID:         payload.GetSenderId(),
		ConversationID: payload.GetConversationId(),
		MessageID:      payload.GetMessageId(),
		Channel:        c,
	}
}

func extractText(body *sharedv1.MessageBody) string {
	if body == nil {
		return ""
	}
	switch body.Type {
	case sharedv1.MessageType_MESSAGE_TYPE_TEXT:
		if t := body.GetText(); t != nil {
			return t.GetText()
		}
	case sharedv1.MessageType_MESSAGE_TYPE_MARKDOWN:
		if m := body.GetMarkdown(); m != nil {
			return m.GetRawMarkdown()
		}
	}
	return ""
}
