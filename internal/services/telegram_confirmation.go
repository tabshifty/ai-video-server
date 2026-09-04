package services

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"video-server/internal/models"
)

const (
	// TelegramConfirmationActionPhoneAuthorization authorizes starting a phone
	// binding or reauthorization flow.
	TelegramConfirmationActionPhoneAuthorization = "authorization_phone"
	// TelegramConfirmationActionQRAuthorization authorizes starting a QR
	// reauthorization flow for an already bound account.
	TelegramConfirmationActionQRAuthorization = "authorization_qr"
	// TelegramConfirmationActionPrivateSourcePreview authorizes resolving a
	// private invitation before its final source confirmation.
	TelegramConfirmationActionPrivateSourcePreview = "source_private_preview"
	// TelegramConfirmationActionPrivateSourceConfirm authorizes the explicit
	// join-capable confirmation of a private invitation after preview.
	TelegramConfirmationActionPrivateSourceConfirm = "source_private_confirm"
)

var (
	// ErrTelegramConfirmationRequired indicates that the UI's explicit second
	// confirmation was absent before a high-risk operation.
	ErrTelegramConfirmationRequired = errors.New("请确认继续高风险 Telegram 操作")
	// ErrTelegramConfirmationInvalidPassword indicates that the current admin
	// password could not be verified.
	ErrTelegramConfirmationInvalidPassword = errors.New("当前管理员密码不正确")
	// ErrTelegramConfirmationNotFound covers unknown, expired, or consumed
	// one-time confirmations without revealing their lifecycle.
	ErrTelegramConfirmationNotFound = errors.New("Telegram 确认票据不存在或已过期")
	// ErrTelegramConfirmationNotOwner prevents ticket replay by another admin.
	ErrTelegramConfirmationNotOwner = errors.New("无权使用该 Telegram 确认票据")
	// ErrTelegramConfirmationActionMismatch prevents one operation's ticket
	// from authorizing another operation.
	ErrTelegramConfirmationActionMismatch = errors.New("Telegram 确认票据不适用于当前操作")
)

// TelegramConfirmationUserRepository provides the current administrator's
// password hash. The service never persists or returns a raw password.
type TelegramConfirmationUserRepository interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (models.User, error)
}

// TelegramConfirmationServiceConfig configures short-lived in-memory
// confirmations for high-risk Telegram management operations.
type TelegramConfirmationServiceConfig struct {
	Repository TelegramConfirmationUserRepository
	Clock      func() time.Time
	TTL        time.Duration
}

// TelegramConfirmationView is the browser-facing, one-time ticket response.
// The ticket must remain in transient page state and never be persisted.
type TelegramConfirmationView struct {
	Ticket    string    `json:"ticket"`
	Action    string    `json:"action"`
	ExpiresAt time.Time `json:"expires_at"`
}

// TelegramConfirmationService verifies an administrator password once and
// creates a short-lived, action-bound ticket for the immediately following
// control operation.
type TelegramConfirmationService struct {
	repository TelegramConfirmationUserRepository
	clock      func() time.Time
	ttl        time.Duration

	mu      sync.Mutex
	tickets map[uuid.UUID]telegramConfirmationRecord
}

type telegramConfirmationRecord struct {
	ownerID   uuid.UUID
	action    string
	expiresAt time.Time
}

// NewTelegramConfirmationService constructs an in-memory one-time ticket
// manager. Tickets intentionally disappear when the API process restarts.
func NewTelegramConfirmationService(config TelegramConfirmationServiceConfig) *TelegramConfirmationService {
	clock := config.Clock
	if clock == nil {
		clock = time.Now
	}
	ttl := config.TTL
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &TelegramConfirmationService{
		repository: config.Repository,
		clock:      clock,
		ttl:        ttl,
		tickets:    make(map[uuid.UUID]telegramConfirmationRecord),
	}
}

// Issue verifies the current administrator password and creates a one-time
// ticket only after the page has explicitly confirmed the requested action.
func (s *TelegramConfirmationService) Issue(ctx context.Context, actorID uuid.UUID, action, password string, confirmed bool) (TelegramConfirmationView, error) {
	if s == nil || s.repository == nil {
		return TelegramConfirmationView{}, errors.New("Telegram 确认服务不可用")
	}
	if ctx == nil {
		return TelegramConfirmationView{}, errors.New("Telegram 确认 context 不能为空")
	}
	if actorID == uuid.Nil {
		return TelegramConfirmationView{}, errors.New("Telegram 确认缺少管理员身份")
	}
	action, err := normalizeTelegramConfirmationAction(action)
	if err != nil {
		return TelegramConfirmationView{}, err
	}
	if !confirmed {
		return TelegramConfirmationView{}, ErrTelegramConfirmationRequired
	}
	if strings.TrimSpace(password) == "" {
		return TelegramConfirmationView{}, ErrTelegramConfirmationInvalidPassword
	}
	user, err := s.repository.GetUserByID(ctx, actorID)
	if err != nil || user.ID != actorID || user.Role != "admin" {
		return TelegramConfirmationView{}, ErrTelegramConfirmationInvalidPassword
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return TelegramConfirmationView{}, ErrTelegramConfirmationInvalidPassword
	}

	now := s.now()
	id := uuid.New()
	expiresAt := now.Add(s.ttl)
	s.mu.Lock()
	s.expireLocked(now)
	s.tickets[id] = telegramConfirmationRecord{ownerID: actorID, action: action, expiresAt: expiresAt}
	s.mu.Unlock()
	return TelegramConfirmationView{Ticket: id.String(), Action: action, ExpiresAt: expiresAt}, nil
}

// Consume atomically validates and removes one ticket for the requested
// administrator and action. A ticket can never be reused after success.
func (s *TelegramConfirmationService) Consume(actorID uuid.UUID, action, ticket string) error {
	if s == nil {
		return ErrTelegramConfirmationNotFound
	}
	if actorID == uuid.Nil {
		return ErrTelegramConfirmationNotOwner
	}
	action, err := normalizeTelegramConfirmationAction(action)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(strings.TrimSpace(ticket))
	if err != nil || id == uuid.Nil {
		return ErrTelegramConfirmationNotFound
	}
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireLocked(now)
	record, ok := s.tickets[id]
	if !ok {
		return ErrTelegramConfirmationNotFound
	}
	if record.ownerID != actorID {
		return ErrTelegramConfirmationNotOwner
	}
	if record.action != action {
		return ErrTelegramConfirmationActionMismatch
	}
	delete(s.tickets, id)
	return nil
}

func normalizeTelegramConfirmationAction(value string) (string, error) {
	action := strings.TrimSpace(value)
	switch action {
	case TelegramConfirmationActionPhoneAuthorization,
		TelegramConfirmationActionQRAuthorization,
		TelegramConfirmationActionPrivateSourcePreview,
		TelegramConfirmationActionPrivateSourceConfirm:
		return action, nil
	default:
		return "", errors.New("Telegram 确认操作无效")
	}
}

func (s *TelegramConfirmationService) now() time.Time {
	if s == nil || s.clock == nil {
		return time.Now().UTC()
	}
	return s.clock().UTC()
}

func (s *TelegramConfirmationService) expireLocked(now time.Time) {
	for id, record := range s.tickets {
		if !now.Before(record.expiresAt) {
			delete(s.tickets, id)
		}
	}
}
