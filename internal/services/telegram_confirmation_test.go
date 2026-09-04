package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"video-server/internal/models"
)

func TestTelegramConfirmationIssuesAndConsumesOneTimeTicket(t *testing.T) {
	t.Parallel()

	actorID := uuid.New()
	now := time.Date(2026, time.September, 4, 8, 40, 0, 0, time.UTC)
	service := newTelegramConfirmationTestService(t, actorID, now)

	view, err := service.Issue(context.Background(), actorID, TelegramConfirmationActionPhoneAuthorization, "admin-password", true)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if view.Ticket == "" || view.Action != TelegramConfirmationActionPhoneAuthorization || !view.ExpiresAt.Equal(now.Add(5*time.Minute)) {
		t.Fatalf("confirmation view = %+v", view)
	}
	if err := service.Consume(actorID, TelegramConfirmationActionPhoneAuthorization, view.Ticket); err != nil {
		t.Fatalf("Consume() error = %v", err)
	}
	if err := service.Consume(actorID, TelegramConfirmationActionPhoneAuthorization, view.Ticket); !errors.Is(err, ErrTelegramConfirmationNotFound) {
		t.Fatalf("second Consume() error = %v, want ErrTelegramConfirmationNotFound", err)
	}
}

func TestTelegramConfirmationBindsTicketToOwnerAndAction(t *testing.T) {
	t.Parallel()

	actorID := uuid.New()
	service := newTelegramConfirmationTestService(t, actorID, time.Date(2026, time.September, 4, 8, 41, 0, 0, time.UTC))
	view, err := service.Issue(context.Background(), actorID, TelegramConfirmationActionPrivateSourcePreview, "admin-password", true)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if err := service.Consume(uuid.New(), TelegramConfirmationActionPrivateSourcePreview, view.Ticket); !errors.Is(err, ErrTelegramConfirmationNotOwner) {
		t.Fatalf("Consume() other owner error = %v", err)
	}
	if err := service.Consume(actorID, TelegramConfirmationActionQRAuthorization, view.Ticket); !errors.Is(err, ErrTelegramConfirmationActionMismatch) {
		t.Fatalf("Consume() other action error = %v", err)
	}
	if err := service.Consume(actorID, TelegramConfirmationActionPrivateSourcePreview, view.Ticket); err != nil {
		t.Fatalf("Consume() rightful owner error = %v", err)
	}
}

func TestTelegramConfirmationRejectsMissingConfirmationInvalidPasswordAndExpiredTicket(t *testing.T) {
	t.Parallel()

	actorID := uuid.New()
	now := time.Date(2026, time.September, 4, 8, 42, 0, 0, time.UTC)
	clock := now
	service := newTelegramConfirmationTestService(t, actorID, now)
	service.clock = func() time.Time { return clock }

	if _, err := service.Issue(context.Background(), actorID, TelegramConfirmationActionPhoneAuthorization, "admin-password", false); !errors.Is(err, ErrTelegramConfirmationRequired) {
		t.Fatalf("Issue() without confirmation error = %v", err)
	}
	if _, err := service.Issue(context.Background(), actorID, TelegramConfirmationActionPhoneAuthorization, "wrong-password", true); !errors.Is(err, ErrTelegramConfirmationInvalidPassword) {
		t.Fatalf("Issue() invalid password error = %v", err)
	}
	view, err := service.Issue(context.Background(), actorID, TelegramConfirmationActionPhoneAuthorization, "admin-password", true)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	clock = now.Add(5 * time.Minute)
	if err := service.Consume(actorID, TelegramConfirmationActionPhoneAuthorization, view.Ticket); !errors.Is(err, ErrTelegramConfirmationNotFound) {
		t.Fatalf("Consume() expired ticket error = %v", err)
	}
}

func newTelegramConfirmationTestService(t *testing.T, actorID uuid.UUID, now time.Time) *TelegramConfirmationService {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("admin-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword() error = %v", err)
	}
	return NewTelegramConfirmationService(TelegramConfirmationServiceConfig{
		Repository: telegramConfirmationTestRepository{user: models.User{
			ID:           actorID,
			Role:         "admin",
			PasswordHash: string(hash),
		}},
		Clock: func() time.Time { return now },
		TTL:   5 * time.Minute,
	})
}

type telegramConfirmationTestRepository struct {
	user models.User
	err  error
}

func (r telegramConfirmationTestRepository) GetUserByID(context.Context, uuid.UUID) (models.User, error) {
	if r.err != nil {
		return models.User{}, r.err
	}
	return r.user, nil
}
