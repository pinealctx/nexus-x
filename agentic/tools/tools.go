package tools

import (
	"context"
	"fmt"

	"charm.land/fantasy"
	"connectrpc.com/connect"
	apiv1 "github.com/pinealctx/nexus-proto/gen/go/api/v1"
	sharedv1 "github.com/pinealctx/nexus-proto/gen/go/shared/v1"

	"github.com/pinealctx/nexus-x/agentic"
	"github.com/pinealctx/nexus-x/client"
)

// --- Tier 1: Basic messaging tools ---

// BasicTools returns all Tier 1 tools.
func BasicTools(c *client.Client) []fantasy.AgentTool {
	return []fantasy.AgentTool{
		SendText(c),
		SendCard(c),
		EditMessage(c),
		ReplyMessage(c),
	}
}

// SendText creates a tool that sends a text/markdown message.
func SendText(c *client.Client) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"send_text",
		"Send a text message (markdown) to the current Nexus conversation. "+
			"Supports @mentions by user ID. Use this to reply to the user.",
		func(ctx context.Context, input sendTextInput, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
			convID := agentic.ConversationIDFromContext(ctx)
			if convID == 0 {
				return fantasy.NewTextErrorResponse("no conversation context"), nil
			}

			entities := buildMentionEntities(input.Text, input.Mentions)
			body := &sharedv1.MessageBody{
				Type: sharedv1.MessageType_MESSAGE_TYPE_MARKDOWN,
				Content: &sharedv1.MessageBody_Markdown{
					Markdown: &sharedv1.MarkdownContent{
						RawMarkdown: input.Text,
						Entities:    entities,
					},
				},
			}

			ch := agentic.ChannelFromContext(ctx)
			result, err := ch.SendMessage(ctx, &agentic.SendMessageRequest{
				ConversationID: convID,
				Body:           body,
			})
			if err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}
			return fantasy.NewTextResponse(fmt.Sprintf("Message sent (id=%d)", result.MessageID)), nil
		},
	)
}

type sendTextInput struct {
	Text     string  `json:"text" jsonschema:"description=Message text in Markdown format"`
	Mentions []int32 `json:"mentions,omitempty" jsonschema:"description=User IDs to @mention in the message"`
}

// SendCard creates a tool that sends an Adaptive Card message.
func SendCard(c *client.Client) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"send_card",
		"Send a structured Adaptive Card to the current conversation. "+
			"Provide title, body text, and optional action buttons.",
		func(ctx context.Context, input sendCardInput, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
			convID := agentic.ConversationIDFromContext(ctx)
			if convID == 0 {
				return fantasy.NewTextErrorResponse("no conversation context"), nil
			}

			cardJSON := buildSimpleCardJSON(input)
			body := &sharedv1.MessageBody{
				Type: sharedv1.MessageType_MESSAGE_TYPE_CARD,
				Content: &sharedv1.MessageBody_Card{
					Card: &sharedv1.CardContent{
						CardJson:     cardJSON,
						FallbackText: input.Title,
					},
				},
			}

			ch := agentic.ChannelFromContext(ctx)
			result, err := ch.SendMessage(ctx, &agentic.SendMessageRequest{
				ConversationID: convID,
				Body:           body,
			})
			if err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}
			return fantasy.NewTextResponse(fmt.Sprintf("Card sent (id=%d)", result.MessageID)), nil
		},
	)
}

type sendCardInput struct {
	Title   string          `json:"title" jsonschema:"description=Card title text"`
	Body    string          `json:"body,omitempty" jsonschema:"description=Card body text"`
	Actions []cardActionDef `json:"actions,omitempty" jsonschema:"description=Action buttons"`
}

type cardActionDef struct {
	Label string `json:"label" jsonschema:"description=Button label"`
	Verb  string `json:"verb,omitempty" jsonschema:"description=Action verb for Action.Submit"`
	URL   string `json:"url,omitempty" jsonschema:"description=URL for Action.OpenUrl (mutually exclusive with verb)"`
}

// EditMessage creates a tool that edits a previously sent message.
func EditMessage(c *client.Client) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"edit_message",
		"Edit a previously sent message in the current conversation. "+
			"Only messages sent by this agent can be edited.",
		func(ctx context.Context, input editMessageInput, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
			convID := agentic.ConversationIDFromContext(ctx)
			if convID == 0 {
				return fantasy.NewTextErrorResponse("no conversation context"), nil
			}

			body := &sharedv1.MessageBody{
				Type: sharedv1.MessageType_MESSAGE_TYPE_MARKDOWN,
				Content: &sharedv1.MessageBody_Markdown{
					Markdown: &sharedv1.MarkdownContent{RawMarkdown: input.Text},
				},
			}

			ch := agentic.ChannelFromContext(ctx)
			if err := ch.EditMessage(ctx, convID, input.MessageID, body); err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}
			return fantasy.NewTextResponse("Message edited"), nil
		},
	)
}

type editMessageInput struct {
	MessageID int64  `json:"message_id" jsonschema:"description=ID of the message to edit"`
	Text      string `json:"text" jsonschema:"description=New message text in Markdown format"`
}

// ReplyMessage creates a tool that sends a reply to a specific message.
func ReplyMessage(c *client.Client) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"reply_message",
		"Send a reply to a specific message in the current conversation. "+
			"The reply will be visually linked to the original message.",
		func(ctx context.Context, input replyMessageInput, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
			convID := agentic.ConversationIDFromContext(ctx)
			if convID == 0 {
				return fantasy.NewTextErrorResponse("no conversation context"), nil
			}

			body := &sharedv1.MessageBody{
				Type: sharedv1.MessageType_MESSAGE_TYPE_MARKDOWN,
				Content: &sharedv1.MessageBody_Markdown{
					Markdown: &sharedv1.MarkdownContent{RawMarkdown: input.Text},
				},
			}

			ch := agentic.ChannelFromContext(ctx)
			result, err := ch.SendMessage(ctx, &agentic.SendMessageRequest{
				ConversationID:   convID,
				Body:             body,
				ReplyToMessageID: &input.ReplyToID,
			})
			if err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}
			return fantasy.NewTextResponse(fmt.Sprintf("Reply sent (id=%d)", result.MessageID)), nil
		},
	)
}

type replyMessageInput struct {
	ReplyToID int64  `json:"reply_to_id" jsonschema:"description=ID of the message to reply to"`
	Text      string `json:"text" jsonschema:"description=Reply text in Markdown format"`
}

// --- Helpers ---

func buildMentionEntities(_ string, userIDs []int32) []*sharedv1.MessageEntity {
	if len(userIDs) == 0 {
		return nil
	}
	var entities []*sharedv1.MessageEntity
	for _, uid := range userIDs {
		entities = append(entities, &sharedv1.MessageEntity{
			Type:   sharedv1.MessageEntityType_MESSAGE_ENTITY_TYPE_MENTION,
			Offset: 0,
			Length: 0,
			Data: &sharedv1.MessageEntity_Mention{
				Mention: &sharedv1.MentionEntity{UserId: uid},
			},
		})
	}
	return entities
}

func buildSimpleCardJSON(input sendCardInput) string {
	// Build a minimal Adaptive Card JSON.
	card := `{"type":"AdaptiveCard","version":"1.5","$schema":"https://adaptivecards.io/schemas/adaptive-card.json","body":[`

	bodyParts := ""
	if input.Title != "" {
		bodyParts += fmt.Sprintf(`{"type":"TextBlock","text":%q,"size":"Large","weight":"Bolder"}`, input.Title)
	}
	if input.Body != "" {
		if bodyParts != "" {
			bodyParts += ","
		}
		bodyParts += fmt.Sprintf(`{"type":"TextBlock","text":%q,"wrap":true}`, input.Body)
	}
	card += bodyParts + `]`

	if len(input.Actions) > 0 {
		card += `,"actions":[`
		for i, a := range input.Actions {
			if i > 0 {
				card += ","
			}
			if a.URL != "" {
				card += fmt.Sprintf(`{"type":"Action.OpenUrl","title":%q,"url":%q}`, a.Label, a.URL)
			} else {
				card += fmt.Sprintf(`{"type":"Action.Submit","title":%q,"data":{"verb":%q}}`, a.Label, a.Verb)
			}
		}
		card += `]`
	}

	card += `}`
	return card
}

// --- Query tools (Tier 2) ---

// QueryTools returns all Tier 2 tools.
func QueryTools(c *client.Client) []fantasy.AgentTool {
	return []fantasy.AgentTool{
		GetMessageHistory(c),
		GetMessage(c),
		GetConversation(c),
		SearchUsers(c),
	}
}

// GetMessageHistory creates a tool that retrieves message history.
func GetMessageHistory(c *client.Client) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"get_message_history",
		"Retrieve recent message history from the current conversation. "+
			"Returns up to `limit` messages. Use before_message_id for pagination.",
		func(ctx context.Context, input getHistoryInput, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
			convID := agentic.ConversationIDFromContext(ctx)
			if convID == 0 {
				return fantasy.NewTextErrorResponse("no conversation context"), nil
			}

			limit := input.Limit
			if limit <= 0 || limit > 50 {
				limit = 20
			}

			req := &apiv1.GetMessageHistoryRequest{
				ConversationId: convID,
				Limit:          limit,
			}
			if input.BeforeMessageID > 0 {
				req.BeforeMessageId = &input.BeforeMessageID
			}

			resp, err := c.Services().Messages.GetMessageHistory(ctx, connect.NewRequest(req))
			if err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}

			var result string
			for _, msg := range resp.Msg.GetMessages() {
				text := extractMessageText(msg)
				result += fmt.Sprintf("[%d] user=%d: %s\n", msg.GetMessageId(), msg.GetSenderId(), text)
			}
			if result == "" {
				result = "No messages found."
			}
			return fantasy.NewTextResponse(result), nil
		},
	)
}

type getHistoryInput struct {
	Limit           int32 `json:"limit,omitempty" jsonschema:"description=Max messages to return (default 20, max 50)"`
	BeforeMessageID int64 `json:"before_message_id,omitempty" jsonschema:"description=Return messages before this ID (for pagination)"`
}

// GetMessage creates a tool that retrieves a single message by ID.
func GetMessage(c *client.Client) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"get_message",
		"Retrieve a single message by ID from the current conversation.",
		func(ctx context.Context, input getMessageInput, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
			convID := agentic.ConversationIDFromContext(ctx)
			if convID == 0 {
				return fantasy.NewTextErrorResponse("no conversation context"), nil
			}

			resp, err := c.Services().Messages.GetMessage(ctx, connect.NewRequest(&apiv1.GetMessageRequest{
				ConversationId: convID,
				MessageId:      input.MessageID,
			}))
			if err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}

			msg := resp.Msg.GetMessage()
			text := extractMessageText(msg)
			result := fmt.Sprintf("message_id=%d sender=%d type=%s text=%s",
				msg.GetMessageId(), msg.GetSenderId(), msg.GetBody().GetType().String(), text)
			return fantasy.NewTextResponse(result), nil
		},
	)
}

type getMessageInput struct {
	MessageID int64 `json:"message_id" jsonschema:"description=ID of the message to retrieve"`
}

// GetConversation creates a tool that retrieves conversation info.
func GetConversation(c *client.Client) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"get_conversation",
		"Get information about the current conversation (type, members, last activity).",
		func(ctx context.Context, _ struct{}, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
			convID := agentic.ConversationIDFromContext(ctx)
			if convID == 0 {
				return fantasy.NewTextErrorResponse("no conversation context"), nil
			}

			resp, err := c.Services().Conversations.GetConversation(ctx, connect.NewRequest(&apiv1.GetConversationRequest{
				ConversationId: convID,
			}))
			if err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}

			conv := resp.Msg.GetConversation()
			result := fmt.Sprintf("conversation_id=%d type=%s muted=%v",
				conv.GetConversationId(), conv.GetType().String(), conv.GetIsMuted())
			return fantasy.NewTextResponse(result), nil
		},
	)
}

// SearchUsers creates a tool that searches for users by keyword.
func SearchUsers(c *client.Client) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"search_users",
		"Search for Nexus users by username or nickname. "+
			"Useful for finding user IDs to @mention.",
		func(ctx context.Context, input searchUsersInput, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
			resp, err := c.Services().Contacts.SearchUsers(ctx, connect.NewRequest(&apiv1.SearchUsersRequest{
				Query: input.Keyword,
			}))
			if err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}

			var result string
			for _, u := range resp.Msg.GetItems() {
				result += fmt.Sprintf("user_id=%d username=%s nickname=%s\n",
					u.GetUserId(), u.GetUsername(), u.GetNickname())
			}
			if result == "" {
				result = "No users found."
			}
			return fantasy.NewTextResponse(result), nil
		},
	)
}

type searchUsersInput struct {
	Keyword string `json:"keyword" jsonschema:"description=Search keyword (username or nickname)"`
}

// AllTools returns all built-in tools (Tier 1 + 2 + 3 + 4).
func AllTools(c *client.Client) []fantasy.AgentTool {
	var all []fantasy.AgentTool
	all = append(all, BasicTools(c)...)
	all = append(all, QueryTools(c)...)
	all = append(all, GroupTools(c)...)
	all = append(all, MediaTools(c)...)
	return all
}

// --- Tier 3: Group tools ---

// GroupTools returns all Tier 3 tools.
func GroupTools(c *client.Client) []fantasy.AgentTool {
	return []fantasy.AgentTool{
		GetGroupInfo(c),
		ListGroups(c),
		InviteMembers(c),
		RemoveMember(c),
	}
}

// GetGroupInfo creates a tool that retrieves group details and members.
func GetGroupInfo(c *client.Client) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"get_group_info",
		"Get group details including name, members, and roles. "+
			"Requires the group_id.",
		func(ctx context.Context, input getGroupInfoInput, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
			resp, err := c.Services().Groups.GetGroupInfo(ctx, connect.NewRequest(&apiv1.GetGroupInfoRequest{
				GroupId: input.GroupID,
			}))
			if err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}

			group := resp.Msg.GetGroup()
			result := fmt.Sprintf("group_id=%d name=%s member_count=%d\nMembers:\n",
				group.GetGroupId(), group.GetName(), len(resp.Msg.GetMembers()))

			userMap := make(map[int32]string)
			for _, u := range resp.Msg.GetUsers() {
				userMap[u.GetUserId()] = u.GetNickname()
			}
			for _, m := range resp.Msg.GetMembers() {
				nick := userMap[m.GetUserId()]
				result += fmt.Sprintf("  user_id=%d nickname=%s role=%s\n",
					m.GetUserId(), nick, m.GetRole().String())
			}
			return fantasy.NewTextResponse(result), nil
		},
	)
}

type getGroupInfoInput struct {
	GroupID int32 `json:"group_id" jsonschema:"description=Group ID to query"`
}

// ListGroups creates a tool that lists groups the agent is a member of.
func ListGroups(c *client.Client) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"list_groups",
		"List all groups this agent is a member of.",
		func(ctx context.Context, _ struct{}, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
			resp, err := c.Services().Groups.ListGroups(ctx, connect.NewRequest(&apiv1.ListGroupsRequest{}))
			if err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}

			var result string
			for _, g := range resp.Msg.GetGroups() {
				result += fmt.Sprintf("group_id=%d name=%s\n", g.GetGroupId(), g.GetName())
			}
			if result == "" {
				result = "Not a member of any groups."
			}
			return fantasy.NewTextResponse(result), nil
		},
	)
}

// InviteMembers creates a tool that invites users to a group.
func InviteMembers(c *client.Client) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"invite_members",
		"Invite users to a group. Only the group owner can invite.",
		func(ctx context.Context, input inviteMembersInput, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
			_, err := c.Services().Groups.InviteMembers(ctx, connect.NewRequest(&apiv1.InviteMembersRequest{
				GroupId:   input.GroupID,
				MemberIds: input.UserIDs,
			}))
			if err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}
			return fantasy.NewTextResponse(fmt.Sprintf("Invited %d members to group %d", len(input.UserIDs), input.GroupID)), nil
		},
	)
}

type inviteMembersInput struct {
	GroupID int32   `json:"group_id" jsonschema:"description=Group ID"`
	UserIDs []int32 `json:"user_ids" jsonschema:"description=User IDs to invite"`
}

// RemoveMember creates a tool that removes a member from a group.
func RemoveMember(c *client.Client) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"remove_member",
		"Remove a member from a group. Only the group owner can remove members.",
		func(ctx context.Context, input removeMemberInput, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
			_, err := c.Services().Groups.RemoveMember(ctx, connect.NewRequest(&apiv1.RemoveMemberRequest{
				GroupId:  input.GroupID,
				TargetId: input.UserID,
			}))
			if err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}
			return fantasy.NewTextResponse(fmt.Sprintf("Removed user %d from group %d", input.UserID, input.GroupID)), nil
		},
	)
}

type removeMemberInput struct {
	GroupID int32 `json:"group_id" jsonschema:"description=Group ID"`
	UserID  int32 `json:"user_id" jsonschema:"description=User ID to remove"`
}

// --- Tier 4: Media tools ---

// MediaTools returns all Tier 4 tools.
func MediaTools(c *client.Client) []fantasy.AgentTool {
	return []fantasy.AgentTool{
		SendImage(c),
		SendFile(c),
		GetDownloadURL(c),
	}
}

// SendImage creates a tool that sends an image message.
func SendImage(c *client.Client) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"send_image",
		"Send an image message to the current conversation. "+
			"Requires a file_id from a previously uploaded file.",
		func(ctx context.Context, input sendImageInput, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
			convID := agentic.ConversationIDFromContext(ctx)
			if convID == 0 {
				return fantasy.NewTextErrorResponse("no conversation context"), nil
			}

			body := &sharedv1.MessageBody{
				Type: sharedv1.MessageType_MESSAGE_TYPE_IMAGE,
				Content: &sharedv1.MessageBody_Image{
					Image: &sharedv1.ImageContent{
						FileId: input.FileID,
						Width:  input.Width,
						Height: input.Height,
					},
				},
			}

			ch := agentic.ChannelFromContext(ctx)
			result, err := ch.SendMessage(ctx, &agentic.SendMessageRequest{
				ConversationID: convID,
				Body:           body,
			})
			if err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}
			return fantasy.NewTextResponse(fmt.Sprintf("Image sent (id=%d)", result.MessageID)), nil
		},
	)
}

type sendImageInput struct {
	FileID string `json:"file_id" jsonschema:"description=File ID of the uploaded image"`
	Width  int32  `json:"width,omitempty" jsonschema:"description=Image width in pixels"`
	Height int32  `json:"height,omitempty" jsonschema:"description=Image height in pixels"`
}

// SendFile creates a tool that sends a file message.
func SendFile(c *client.Client) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"send_file",
		"Send a file message to the current conversation. "+
			"Requires a file_id from a previously uploaded file.",
		func(ctx context.Context, input sendFileInput, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
			convID := agentic.ConversationIDFromContext(ctx)
			if convID == 0 {
				return fantasy.NewTextErrorResponse("no conversation context"), nil
			}

			body := &sharedv1.MessageBody{
				Type: sharedv1.MessageType_MESSAGE_TYPE_FILE,
				Content: &sharedv1.MessageBody_File{
					File: &sharedv1.FileContent{
						FileId:   input.FileID,
						Filename: input.Filename,
						MimeType: input.MimeType,
					},
				},
			}

			ch := agentic.ChannelFromContext(ctx)
			result, err := ch.SendMessage(ctx, &agentic.SendMessageRequest{
				ConversationID: convID,
				Body:           body,
			})
			if err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}
			return fantasy.NewTextResponse(fmt.Sprintf("File sent (id=%d)", result.MessageID)), nil
		},
	)
}

type sendFileInput struct {
	FileID   string `json:"file_id" jsonschema:"description=File ID of the uploaded file"`
	Filename string `json:"filename" jsonschema:"description=Display filename"`
	MimeType string `json:"mime_type,omitempty" jsonschema:"description=MIME type of the file"`
}

// GetDownloadURL creates a tool that gets a download URL for a file.
func GetDownloadURL(c *client.Client) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"get_download_url",
		"Get a temporary download URL for a file by its file_id. "+
			"Use this to access files sent by users (images, documents, etc.).",
		func(ctx context.Context, input getDownloadURLInput, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
			resp, err := c.Services().Media.GetDownloadURL(ctx, connect.NewRequest(&apiv1.GetDownloadURLRequest{
				FileId: input.FileID,
			}))
			if err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}
			return fantasy.NewTextResponse(resp.Msg.GetUrl()), nil
		},
	)
}

type getDownloadURLInput struct {
	FileID string `json:"file_id" jsonschema:"description=File ID to get download URL for"`
}

// --- Internal helpers ---

func extractMessageText(msg *sharedv1.MessageEnvelope) string {
	if msg == nil || msg.GetBody() == nil {
		return ""
	}
	body := msg.GetBody()
	switch body.Type {
	case sharedv1.MessageType_MESSAGE_TYPE_TEXT:
		if t := body.GetText(); t != nil {
			return t.GetText()
		}
	case sharedv1.MessageType_MESSAGE_TYPE_MARKDOWN:
		if m := body.GetMarkdown(); m != nil {
			return m.GetRawMarkdown()
		}
	case sharedv1.MessageType_MESSAGE_TYPE_CARD:
		return "[card]"
	case sharedv1.MessageType_MESSAGE_TYPE_IMAGE:
		return "[image]"
	case sharedv1.MessageType_MESSAGE_TYPE_AUDIO:
		return "[audio]"
	case sharedv1.MessageType_MESSAGE_TYPE_VIDEO:
		return "[video]"
	case sharedv1.MessageType_MESSAGE_TYPE_FILE:
		return "[file]"
	}
	return ""
}
