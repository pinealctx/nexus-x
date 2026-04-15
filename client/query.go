package client

import (
	"context"

	"connectrpc.com/connect"
	apiv1 "github.com/pinealctx/nexus-proto/gen/go/api/v1"
	sharedv1 "github.com/pinealctx/nexus-proto/gen/go/shared/v1"
)

// --- Conversations ---

// GetConversation returns a single conversation by ID.
func (c *Client) GetConversation(ctx context.Context, conversationID int64) (*sharedv1.ConversationInfo, error) {
	resp, err := c.conversations.GetConversation(ctx, connect.NewRequest(&apiv1.GetConversationRequest{
		ConversationId: conversationID,
	}))
	if err != nil {
		return nil, err
	}
	return resp.Msg.GetConversation(), nil
}

// ConversationPage holds a page of conversations.
type ConversationPage struct {
	Conversations []*sharedv1.ConversationInfo
	RelatedUsers  []*sharedv1.UserInfo
	RelatedGroups []*sharedv1.GroupInfo
	HasMore       bool
}

// ListConversations returns conversations ordered by last_message_time descending.
// Pass beforeTime=0 for the first page.
func (c *Client) ListConversations(ctx context.Context, beforeTime int64, limit int32) (*ConversationPage, error) {
	req := &apiv1.ListConversationsRequest{Limit: limit}
	if beforeTime > 0 {
		req.BeforeTime = &beforeTime
	}
	resp, err := c.conversations.ListConversations(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return &ConversationPage{
		Conversations: resp.Msg.GetConversations(),
		RelatedUsers:  resp.Msg.GetRelatedUsers(),
		RelatedGroups: resp.Msg.GetRelatedGroups(),
		HasMore:       resp.Msg.GetHasMore(),
	}, nil
}

// MarkAsRead marks messages as read up to the given message ID.
func (c *Client) MarkAsRead(ctx context.Context, conversationID int64, upToMessageID int64) error {
	_, err := c.conversations.MarkAsRead(ctx, connect.NewRequest(&apiv1.MarkAsReadRequest{
		ConversationId: conversationID,
		UpToMessageId:  upToMessageID,
	}))
	return err
}

// --- Messages ---

// GetMessage returns a single message by ID.
func (c *Client) GetMessage(ctx context.Context, conversationID, messageID int64) (*sharedv1.MessageEnvelope, error) {
	resp, err := c.messages.GetMessage(ctx, connect.NewRequest(&apiv1.GetMessageRequest{
		ConversationId: conversationID,
		MessageId:      messageID,
	}))
	if err != nil {
		return nil, err
	}
	return resp.Msg.GetMessage(), nil
}

// MessagePage holds a page of messages.
type MessagePage struct {
	Messages     []*sharedv1.MessageEnvelope
	RelatedUsers []*sharedv1.UserInfo
	HasMore      bool
}

// GetMessageHistory returns message history with cursor-based pagination.
// For backward (older): set beforeMessageID > 0.
// For forward (newer): set afterMessageID > 0.
// Omit both for the latest messages.
func (c *Client) GetMessageHistory(ctx context.Context, conversationID int64, beforeMessageID, afterMessageID int64, limit int32) (*MessagePage, error) {
	req := &apiv1.GetMessageHistoryRequest{
		ConversationId: conversationID,
		Limit:          limit,
	}
	if beforeMessageID > 0 {
		req.BeforeMessageId = &beforeMessageID
	}
	if afterMessageID > 0 {
		req.AfterMessageId = &afterMessageID
	}
	resp, err := c.messages.GetMessageHistory(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return &MessagePage{
		Messages:     resp.Msg.GetMessages(),
		RelatedUsers: resp.Msg.GetRelatedUsers(),
		HasMore:      resp.Msg.GetHasMore(),
	}, nil
}

// --- Contacts ---

// ListContacts returns the agent's contact list.
func (c *Client) ListContacts(ctx context.Context, afterID int32, limit int32) ([]*sharedv1.ContactItem, bool, error) {
	req := &apiv1.ListContactsRequest{Limit: limit}
	if afterID > 0 {
		req.AfterId = &afterID
	}
	resp, err := c.contacts.ListContacts(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, false, err
	}
	return resp.Msg.GetContacts(), resp.Msg.GetHasMore(), nil
}

// BlockUser blocks a user, preventing them from sending messages to the agent.
func (c *Client) BlockUser(ctx context.Context, userID int32) error {
	_, err := c.contacts.BlockUser(ctx, connect.NewRequest(&apiv1.BlockUserRequest{
		TargetUserId: userID,
	}))
	return err
}

// UnblockUser unblocks a previously blocked user.
func (c *Client) UnblockUser(ctx context.Context, userID int32) error {
	_, err := c.contacts.UnblockUser(ctx, connect.NewRequest(&apiv1.UnblockUserRequest{
		TargetUserId: userID,
	}))
	return err
}

// --- Groups ---

// ListGroups returns the groups the agent is a member of.
func (c *Client) ListGroups(ctx context.Context) ([]*sharedv1.GroupInfo, error) {
	resp, err := c.groups.ListGroups(ctx, connect.NewRequest(&apiv1.ListGroupsRequest{}))
	if err != nil {
		return nil, err
	}
	return resp.Msg.GetGroups(), nil
}

// GroupDetail holds group info with members.
type GroupDetail struct {
	Group   *sharedv1.GroupInfo
	Members []*sharedv1.MemberInfo
	Users   []*sharedv1.UserInfo
}

// GetGroupInfo returns group details with all members.
func (c *Client) GetGroupInfo(ctx context.Context, groupID int32) (*GroupDetail, error) {
	resp, err := c.groups.GetGroupInfo(ctx, connect.NewRequest(&apiv1.GetGroupInfoRequest{
		GroupId: groupID,
	}))
	if err != nil {
		return nil, err
	}
	return &GroupDetail{
		Group:   resp.Msg.GetGroup(),
		Members: resp.Msg.GetMembers(),
		Users:   resp.Msg.GetUsers(),
	}, nil
}
