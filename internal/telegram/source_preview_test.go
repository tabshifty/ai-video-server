package telegram

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSourcePreviewServiceRequiresSameOwnerAndReferenceForConfirmation(t *testing.T) {
	t.Parallel()

	actorID := uuid.New()
	resolver := &sourcePreviewFakeResolver{
		preview: ChatPreviewResult{
			Title:        "私密超级群",
			ChatType:     TelegramChatTypeSupergroup,
			RequiresJoin: true,
		},
		confirmed: ChatPreviewResult{
			ChatID:   -1000000000077,
			Title:    "私密超级群",
			ChatType: TelegramChatTypeSupergroup,
		},
	}
	service := NewSourcePreviewService(resolver, nil, time.Minute)
	const inviteRef = "https://t.me/+private-invite-token"

	preview, err := service.Preview(context.Background(), actorID.String(), inviteRef)
	if err != nil {
		t.Fatalf("Preview() error = %v", err)
	}
	if preview.PreviewID == "" || !preview.RequiresJoin || preview.RequiresApproval {
		t.Fatalf("Preview() = %+v", preview)
	}
	if resolver.previewCalls != 1 || resolver.confirmCalls != 0 {
		t.Fatalf("resolver calls after preview = preview:%d confirm:%d", resolver.previewCalls, resolver.confirmCalls)
	}
	if strings.Contains(fmt.Sprintf("%+v", service.previews), "private-invite-token") {
		t.Fatalf("source preview state retained private invite token: %+v", service.previews)
	}

	if _, err := service.Confirm(context.Background(), uuid.New().String(), preview.PreviewID, inviteRef); !errors.Is(err, ErrSourcePreviewNotOwner) {
		t.Fatalf("Confirm() by another actor error = %v, want ErrSourcePreviewNotOwner", err)
	}
	if _, err := service.Confirm(context.Background(), actorID.String(), preview.PreviewID, "https://t.me/+different-token"); !errors.Is(err, ErrSourcePreviewReferenceMismatch) {
		t.Fatalf("Confirm() with another reference error = %v, want ErrSourcePreviewReferenceMismatch", err)
	}
	if resolver.confirmCalls != 0 {
		t.Fatalf("resolver confirm calls before matching confirmation = %d, want 0", resolver.confirmCalls)
	}

	confirmed, err := service.Confirm(context.Background(), actorID.String(), preview.PreviewID, inviteRef)
	if err != nil {
		t.Fatalf("Confirm() error = %v", err)
	}
	if confirmed.ChatID != -1000000000077 || confirmed.RequiresJoin || confirmed.RequiresApproval {
		t.Fatalf("Confirm() = %+v", confirmed)
	}
	if resolver.confirmCalls != 1 {
		t.Fatalf("resolver confirm calls = %d, want 1", resolver.confirmCalls)
	}
}

func TestSourcePreviewServiceExpiresAndRejectsUnresolvedConfirmation(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.September, 4, 8, 0, 0, 0, time.UTC)
	resolver := &sourcePreviewFakeResolver{
		preview: ChatPreviewResult{ChatID: -1000000000088, Title: "频道", ChatType: TelegramChatTypeChannel},
		confirmed: ChatPreviewResult{
			Title:            "仍需审核",
			ChatType:         TelegramChatTypeSupergroup,
			RequiresJoin:     true,
			RequiresApproval: true,
		},
	}
	service := NewSourcePreviewService(resolver, func() time.Time { return now }, time.Minute)
	actorID := uuid.New()
	preview, err := service.Preview(context.Background(), actorID.String(), "@channel")
	if err != nil {
		t.Fatalf("Preview() error = %v", err)
	}
	if _, err := service.Confirm(context.Background(), actorID.String(), preview.PreviewID, "@channel"); !errors.Is(err, ErrSourcePreviewUnresolved) {
		t.Fatalf("Confirm() unresolved result error = %v, want ErrSourcePreviewUnresolved", err)
	}

	now = now.Add(2 * time.Minute)
	if _, err := service.Confirm(context.Background(), actorID.String(), preview.PreviewID, "@channel"); !errors.Is(err, ErrSourcePreviewNotFound) {
		t.Fatalf("Confirm() after expiry error = %v, want ErrSourcePreviewNotFound", err)
	}
}

type sourcePreviewFakeResolver struct {
	preview      ChatPreviewResult
	confirmed    ChatPreviewResult
	previewCalls int
	confirmCalls int
}

func (f *sourcePreviewFakeResolver) PreviewChat(context.Context, string) (ChatPreviewResult, error) {
	f.previewCalls++
	return f.preview, nil
}

func (f *sourcePreviewFakeResolver) ConfirmChat(context.Context, string) (ChatPreviewResult, error) {
	f.confirmCalls++
	return f.confirmed, nil
}
