package telegram

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
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

func TestChatPreviewFromClassClassifiesSupportedSourceTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		chat tg.ChatClass
		want ChatPreviewResult
	}{
		{
			name: "basic group",
			chat: &tg.Chat{ID: 42, Title: "基础群"},
			want: ChatPreviewResult{ChatID: 42, Title: "基础群", ChatType: TelegramChatTypeGroup},
		},
		{
			name: "supergroup",
			chat: &tg.Channel{ID: 43, Title: "超级群", Megagroup: true, Username: "supergroup"},
			want: ChatPreviewResult{ChatID: -1000000000043, Title: "超级群", Username: "supergroup", ChatType: TelegramChatTypeSupergroup},
		},
		{
			name: "channel",
			chat: &tg.Channel{ID: 44, Title: "频道", Broadcast: true},
			want: ChatPreviewResult{ChatID: -1000000000044, Title: "频道", ChatType: TelegramChatTypeChannel},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := chatPreviewFromClass(tt.chat)
			if err != nil {
				t.Fatalf("chatPreviewFromClass() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("chatPreviewFromClass() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestChatPreviewFromInviteDoesNotRequireJoiningToInspect(t *testing.T) {
	t.Parallel()

	preview, _, err := chatPreviewFromInvite(&tg.ChatInvite{
		Title:         "私密超级群",
		Channel:       true,
		Megagroup:     true,
		RequestNeeded: true,
	})
	if err != nil {
		t.Fatalf("chatPreviewFromInvite() error = %v", err)
	}
	want := ChatPreviewResult{
		Title:            "私密超级群",
		ChatType:         TelegramChatTypeSupergroup,
		RequiresJoin:     true,
		RequiresApproval: true,
	}
	if preview != want {
		t.Fatalf("chatPreviewFromInvite() = %+v, want %+v", preview, want)
	}
}

func TestChatPreviewFromInviteDistinguishesJoinedAndPeekedChats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		invite tg.ChatInviteClass
		want   ChatPreviewResult
	}{
		{
			name: "already joined",
			invite: &tg.ChatInviteAlready{Chat: &tg.Channel{
				ID:        75,
				Title:     "已加入频道",
				Broadcast: true,
			}},
			want: ChatPreviewResult{
				ChatID:   -1000000000075,
				Title:    "已加入频道",
				ChatType: TelegramChatTypeChannel,
			},
		},
		{
			name: "peek requires confirmation",
			invite: &tg.ChatInvitePeek{Chat: &tg.Channel{
				ID:        76,
				Title:     "可窥视超级群",
				Megagroup: true,
			}},
			want: ChatPreviewResult{
				ChatID:       -1000000000076,
				Title:        "可窥视超级群",
				ChatType:     TelegramChatTypeSupergroup,
				RequiresJoin: true,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := chatPreviewFromInvite(tt.invite)
			if err != nil {
				t.Fatalf("chatPreviewFromInvite() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("chatPreviewFromInvite() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestRequireTelegramAuthorizedRejectsMissingSessionAuthorization(t *testing.T) {
	t.Parallel()

	if err := requireTelegramAuthorized(nil); !errors.Is(err, ErrTelegramSessionUnauthorized) {
		t.Fatalf("requireTelegramAuthorized(nil) error = %v, want ErrTelegramSessionUnauthorized", err)
	}
	if err := requireTelegramAuthorized(&auth.Status{}); !errors.Is(err, ErrTelegramSessionUnauthorized) {
		t.Fatalf("requireTelegramAuthorized() error = %v, want ErrTelegramSessionUnauthorized", err)
	}
	if err := requireTelegramAuthorized(&auth.Status{Authorized: true}); err != nil {
		t.Fatalf("requireTelegramAuthorized() error = %v", err)
	}
}
