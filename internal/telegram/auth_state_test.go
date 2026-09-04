package telegram

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAuthorizationStateMachineAllowsOnlyOneActiveOwner(t *testing.T) {
	t.Parallel()

	clock := func() time.Time { return time.Date(2026, 9, 3, 5, 0, 0, 0, time.UTC) }
	state := NewAuthorizationStateMachine(clock, 10*time.Minute)
	owner := uuid.New()
	first, err := state.Start(AuthorizationStart{Kind: AuthorizationKindPhone, OwnerID: owner})
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if first.Status != AuthorizationStatusAwaitingCode {
		t.Fatalf("first status = %q, want awaiting_code", first.Status)
	}
	if _, err := state.Start(AuthorizationStart{Kind: AuthorizationKindQR, OwnerID: uuid.New()}); !errors.Is(err, ErrAuthorizationInProgress) {
		t.Fatalf("second Start() error = %v, want ErrAuthorizationInProgress", err)
	}
	if state.CanSubmit(first.ID, owner) != true {
		t.Fatal("owner should be able to submit the active authorization")
	}
	if state.CanSubmit(first.ID, uuid.New()) {
		t.Fatal("another administrator must not be able to submit secrets")
	}
}

func TestAuthorizationStateMachineTransitionsAndExpiresWithoutSecrets(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 9, 3, 5, 0, 0, 0, time.UTC)
	current := now
	state := NewAuthorizationStateMachine(func() time.Time { return current }, 5*time.Minute)
	authorization, err := state.Start(AuthorizationStart{Kind: AuthorizationKindQR, OwnerID: uuid.New()})
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := state.Transition(authorization.ID, AuthorizationStatusScanning, ""); err != nil {
		t.Fatalf("Transition() error = %v", err)
	}
	view, ok := state.View(authorization.ID, uuid.New())
	if !ok || view.Status != AuthorizationStatusScanning {
		t.Fatalf("View() = %+v, ok=%v", view, ok)
	}
	current = current.Add(6 * time.Minute)
	view, ok = state.View(authorization.ID, uuid.New())
	if !ok || view.Status != AuthorizationStatusExpired {
		t.Fatalf("expired View() = %+v, ok=%v", view, ok)
	}
	if view.QRImageDataURL != "" || view.Error != "" {
		t.Fatalf("public view must not contain transient secrets: %+v", view)
	}
}

func TestAuthorizationStateMachineExpiresQRImageWithoutExpiringAuthorization(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 9, 4, 8, 0, 0, 0, time.UTC)
	current := now
	state := NewAuthorizationStateMachine(func() time.Time { return current }, 10*time.Minute)
	authorization, err := state.Start(AuthorizationStart{Kind: AuthorizationKindQR, OwnerID: uuid.New()})
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := state.SetQRImage(authorization.ID, "data:image/png;base64,temporary", now.Add(30*time.Second)); err != nil {
		t.Fatalf("SetQRImage() error = %v", err)
	}
	if authorization.ExpiresAt != now.Add(10*time.Minute) {
		t.Fatalf("authorization expiry changed after QR image update: %s", authorization.ExpiresAt)
	}

	current = current.Add(31 * time.Second)
	view, ok := state.View(authorization.ID, uuid.New())
	if !ok || view.Status != AuthorizationStatusScanning {
		t.Fatalf("View() = %+v, ok=%v; QR refresh must not expire authorization", view, ok)
	}
	if view.QRImageDataURL != "" || !view.QRImageExpiresAt.IsZero() {
		t.Fatalf("expired QR image leaked into view: %+v", view)
	}
}
