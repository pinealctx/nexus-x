package nxutil

// PrivateConversationID computes the deterministic PRIVATE conversation
// ID from two user IDs (D04 §11.1):
//
//	int64(max(a,b)) << 32 | int64(min(a,b))
func PrivateConversationID(a, b int32) int64 {
	hi, lo := a, b
	if lo > hi {
		hi, lo = lo, hi
	}
	return int64(hi)<<32 | int64(lo)
}

// IsPrivateConversation returns true if the conversation ID encodes a
// PRIVATE conversation (high 32 bits > 0).
func IsPrivateConversation(conversationID int64) bool {
	return conversationID>>32 != 0
}

// IsGroupConversation returns true if the conversation ID encodes a
// GROUP conversation (high 32 bits == 0).
func IsGroupConversation(conversationID int64) bool {
	return conversationID>>32 == 0
}

// ConversationPeerID extracts the peer user ID from a PRIVATE conversation ID
// given the current user's ID. For GROUP conversations it returns the group ID.
func ConversationPeerID(conversationID int64, currentUserID int32) int32 {
	if IsGroupConversation(conversationID) {
		return int32(conversationID) //nolint:gosec
	}
	high := int32(conversationID >> 32)       //nolint:gosec
	low := int32(conversationID & 0xFFFFFFFF) //nolint:gosec
	if high == currentUserID {
		return low
	}
	return high
}

// ConversationUserIDs extracts both user IDs from a PRIVATE conversation ID.
// Returns (0, 0) for GROUP conversations.
func ConversationUserIDs(conversationID int64) (int32, int32) {
	high := int32(conversationID >> 32) //nolint:gosec
	if high == 0 {
		return 0, 0
	}
	low := int32(conversationID & 0xFFFFFFFF) //nolint:gosec
	return high, low
}

// GroupIDFromConversation extracts the group ID from a GROUP conversation ID.
// Returns 0 for PRIVATE conversations.
func GroupIDFromConversation(conversationID int64) int32 {
	if IsPrivateConversation(conversationID) {
		return 0
	}
	return int32(conversationID) //nolint:gosec
}
