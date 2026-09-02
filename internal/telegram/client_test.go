package telegram

import (
	"context"
	"io"
	"testing"
	"time"
)

type fakeClient struct {
	historyChatID    int64
	historyOffsetID  int64
	historyLimit     int
	historyMessages  []Message
	refreshChatID    int64
	refreshMessageID int64
	refreshMessage   Message
}

var _ Client = (*fakeClient)(nil)

func (f *fakeClient) ResolveChat(context.Context, string) (Chat, error) {
	return Chat{}, nil
}

func (f *fakeClient) History(_ context.Context, chatID, offsetID int64, limit int, fn func(Message) error) error {
	f.historyChatID = chatID
	f.historyOffsetID = offsetID
	f.historyLimit = limit
	for _, message := range f.historyMessages {
		if err := fn(message); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeClient) Subscribe(context.Context, func(Message) error) error {
	return nil
}

func (f *fakeClient) RefreshMessage(_ context.Context, chatID, messageID int64) (Message, error) {
	f.refreshChatID = chatID
	f.refreshMessageID = messageID
	return f.refreshMessage, nil
}

func (f *fakeClient) Download(context.Context, Message, io.Writer) error {
	return nil
}

func TestClientHistoryForwardsOffsetAndLimit(t *testing.T) {
	t.Parallel()

	wantMessage := Message{ChatID: -100123, MessageID: 42, Filename: "clip.mp4"}
	fake := &fakeClient{historyMessages: []Message{wantMessage}}
	var got []Message
	if err := fake.History(context.Background(), -100123, 900, 50, func(message Message) error {
		got = append(got, message)
		return nil
	}); err != nil {
		t.Fatalf("History() error = %v", err)
	}
	if fake.historyChatID != -100123 || fake.historyOffsetID != 900 || fake.historyLimit != 50 {
		t.Fatalf("History arguments = chat=%d offset=%d limit=%d", fake.historyChatID, fake.historyOffsetID, fake.historyLimit)
	}
	if len(got) != 1 || got[0] != wantMessage {
		t.Fatalf("History callback messages = %+v, want %+v", got, []Message{wantMessage})
	}
}

func TestClientRefreshMessageForwardsMessageIdentity(t *testing.T) {
	t.Parallel()

	want := Message{ChatID: -100123, MessageID: 42, DocumentID: 777, SentAt: time.Now().UTC()}
	fake := &fakeClient{refreshMessage: want}
	got, err := fake.RefreshMessage(context.Background(), want.ChatID, want.MessageID)
	if err != nil {
		t.Fatalf("RefreshMessage() error = %v", err)
	}
	if fake.refreshChatID != want.ChatID || fake.refreshMessageID != want.MessageID {
		t.Fatalf("RefreshMessage arguments = chat=%d message=%d", fake.refreshChatID, fake.refreshMessageID)
	}
	if got != want {
		t.Fatalf("RefreshMessage() = %+v, want %+v", got, want)
	}
}
