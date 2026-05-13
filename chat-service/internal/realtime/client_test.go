package realtime

import (
	"context"
	"testing"
)

type fakeTypingNotifier struct {
	userID         string
	conversationID string
	calls          int
}

func (f *fakeTypingNotifier) NotifyTyping(_ context.Context, userID, conversationID string) error {
	f.calls++
	f.userID = userID
	f.conversationID = conversationID
	return nil
}

func TestClientHandleMessageTyping(t *testing.T) {
	notifier := &fakeTypingNotifier{}
	client := &Client{
		UserID:         "user-a",
		TypingNotifier: notifier,
	}

	client.handleMessage([]byte(`{"type":"chat.typing","conversationId":"conv-1"}`))

	if notifier.calls != 1 {
		t.Fatalf("NotifyTyping calls = %d, want 1", notifier.calls)
	}
	if notifier.userID != "user-a" {
		t.Errorf("userID = %q, want user-a", notifier.userID)
	}
	if notifier.conversationID != "conv-1" {
		t.Errorf("conversationID = %q, want conv-1", notifier.conversationID)
	}
}

func TestClientHandleMessageIgnoresUnknownType(t *testing.T) {
	notifier := &fakeTypingNotifier{}
	client := &Client{
		UserID:         "user-a",
		TypingNotifier: notifier,
	}

	client.handleMessage([]byte(`{"type":"chat.message","conversationId":"conv-1"}`))

	if notifier.calls != 0 {
		t.Fatalf("NotifyTyping calls = %d, want 0", notifier.calls)
	}
}
