package nxproto

import (
	sharedv1 "github.com/pinealctx/nexus-proto/gen/go/shared/v1"

	"github.com/pinealctx/nexus-x/nxutil"
)

// ConversationType returns the proto ConversationType enum for a conversation ID.
func ConversationType(conversationID int64) sharedv1.ConversationType {
	if nxutil.IsGroupConversation(conversationID) {
		return sharedv1.ConversationType_CONVERSATION_TYPE_GROUP
	}
	return sharedv1.ConversationType_CONVERSATION_TYPE_PRIVATE
}
