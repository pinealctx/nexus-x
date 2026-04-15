package nxutil_test

import (
	"testing"

	"github.com/pinealctx/nexus-x/nxutil"
)

func TestPrivateConversationID(t *testing.T) {
	id1 := nxutil.PrivateConversationID(10, 20)
	id2 := nxutil.PrivateConversationID(20, 10)
	if id1 != id2 {
		t.Errorf("PrivateConversationID(10,20) = %d, PrivateConversationID(20,10) = %d", id1, id2)
	}

	high := int32(id1 >> 32)
	low := int32(id1 & 0xFFFFFFFF)
	if high != 20 || low != 10 {
		t.Errorf("high=%d, low=%d, want 20, 10", high, low)
	}
}

func TestConversationType(t *testing.T) {
	privateID := nxutil.PrivateConversationID(10, 20)
	if !nxutil.IsPrivateConversation(privateID) {
		t.Error("expected PRIVATE")
	}
	if nxutil.IsGroupConversation(privateID) {
		t.Error("expected not GROUP")
	}

	groupID := int64(42)
	if !nxutil.IsGroupConversation(groupID) {
		t.Error("expected GROUP")
	}
	if nxutil.IsPrivateConversation(groupID) {
		t.Error("expected not PRIVATE")
	}
}

func TestConversationPeerID(t *testing.T) {
	convID := nxutil.PrivateConversationID(10, 20)
	if peer := nxutil.ConversationPeerID(convID, 10); peer != 20 {
		t.Errorf("peer for user 10 = %d, want 20", peer)
	}
	if peer := nxutil.ConversationPeerID(convID, 20); peer != 10 {
		t.Errorf("peer for user 20 = %d, want 10", peer)
	}
}

func TestConversationUserIDs(t *testing.T) {
	convID := nxutil.PrivateConversationID(5, 100)
	a, b := nxutil.ConversationUserIDs(convID)
	if a != 100 || b != 5 {
		t.Errorf("got (%d, %d), want (100, 5)", a, b)
	}

	a, b = nxutil.ConversationUserIDs(42)
	if a != 0 || b != 0 {
		t.Errorf("group got (%d, %d), want (0, 0)", a, b)
	}
}

func TestGroupIDFromConversation(t *testing.T) {
	if gid := nxutil.GroupIDFromConversation(42); gid != 42 {
		t.Errorf("got %d, want 42", gid)
	}
	privateID := nxutil.PrivateConversationID(10, 20)
	if gid := nxutil.GroupIDFromConversation(privateID); gid != 0 {
		t.Errorf("private got %d, want 0", gid)
	}
}

func TestWebhookSignVerify(t *testing.T) {
	secret := "test-secret"
	ts := nxutil.UnixNowString() // use current time to pass drift check
	body := []byte(`{"type":"message"}`)

	sig := nxutil.WebhookSign(secret, ts, body)
	if !nxutil.WebhookVerify(secret, sig, ts, body) {
		t.Error("valid signature rejected")
	}
	if nxutil.WebhookVerify(secret, "sha256=bad", ts, body) {
		t.Error("invalid signature accepted")
	}
	// Expired timestamp should fail.
	if nxutil.WebhookVerify(secret, sig, "1700000000", body) {
		t.Error("expired timestamp accepted")
	}
}

func TestInitDataSignVerify(t *testing.T) {
	secret := "agent-secret-key"

	params := make(map[string][]string)
	params["user"] = []string{`{"user_id":13}`}
	params["agent_user_id"] = []string{"42"}
	params["auth_date"] = []string{nxutil.UnixNowString()}
	params["platform"] = []string{"desktop"}

	hash := nxutil.InitDataSign(secret, params)
	params["hash"] = []string{hash}

	// Build query string using net/url.
	qs := ""
	for k, v := range params {
		if qs != "" {
			qs += "&"
		}
		qs += k + "=" + v[0]
	}

	result, err := nxutil.InitDataVerify(secret, qs)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if result.Get("user") != `{"user_id":13}` {
		t.Errorf("user = %q", result.Get("user"))
	}
}
