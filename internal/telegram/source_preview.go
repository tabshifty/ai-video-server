package telegram

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrSourcePreviewNotFound indicates an unknown, expired, or consumed source preview.
	ErrSourcePreviewNotFound = errors.New("Telegram 来源预览不存在或已过期")
	// ErrSourcePreviewNotOwner prevents another administrator from confirming a preview.
	ErrSourcePreviewNotOwner = errors.New("无权确认该 Telegram 来源预览")
	// ErrSourcePreviewReferenceMismatch prevents confirming a different chat reference.
	ErrSourcePreviewReferenceMismatch = errors.New("Telegram 来源引用与预览不一致")
	// ErrSourcePreviewUnresolved prevents a source from being created before it has a durable identity.
	ErrSourcePreviewUnresolved = errors.New("Telegram 来源尚未获得可保存的 chat ID")
)

// SourcePreviewResolver supplies the session-owning operations needed by the
// in-memory preview coordinator.
type SourcePreviewResolver interface {
	PreviewChat(ctx context.Context, ref string) (ChatPreviewResult, error)
	ConfirmChat(ctx context.Context, ref string) (ChatPreviewResult, error)
}

// SourcePreviewService binds one short-lived preview to its initiating
// administrator without retaining the original source reference or invite token.
type SourcePreviewService struct {
	resolver SourcePreviewResolver
	clock    func() time.Time
	ttl      time.Duration

	mu       sync.Mutex
	previews map[uuid.UUID]sourcePreviewRecord
}

type sourcePreviewRecord struct {
	ownerID        uuid.UUID
	refFingerprint [sha256.Size]byte
	result         ChatPreviewResult
	expiresAt      time.Time
}

// NewSourcePreviewService constructs an in-memory coordinator. Preview state
// intentionally disappears on collector restart.
func NewSourcePreviewService(resolver SourcePreviewResolver, clock func() time.Time, ttl time.Duration) *SourcePreviewService {
	if clock == nil {
		clock = time.Now
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &SourcePreviewService{
		resolver: resolver,
		clock:    clock,
		ttl:      ttl,
		previews: make(map[uuid.UUID]sourcePreviewRecord),
	}
}

// Preview obtains a sanitized source candidate and binds it to actorID for a
// short confirmation window. The source reference remains only on the stack.
func (s *SourcePreviewService) Preview(ctx context.Context, actorValue, chatRef string) (ChatPreview, error) {
	if s == nil || s.resolver == nil {
		return ChatPreview{}, errors.New("Telegram 来源预览服务不可用")
	}
	if ctx == nil {
		return ChatPreview{}, errors.New("Telegram 来源预览 context 不能为空")
	}
	actorID, err := parseSourcePreviewActorID(actorValue)
	if err != nil {
		return ChatPreview{}, err
	}
	fingerprint, err := sourcePreviewFingerprint(chatRef)
	if err != nil {
		return ChatPreview{}, err
	}
	result, err := s.resolver.PreviewChat(ctx, strings.TrimSpace(chatRef))
	if err != nil {
		return ChatPreview{}, err
	}
	if err := validateSourcePreviewResult(result, false); err != nil {
		return ChatPreview{}, err
	}

	now := s.now()
	expiresAt := now.Add(s.ttl)
	id := uuid.New()
	s.mu.Lock()
	s.expireLocked(now)
	s.previews[id] = sourcePreviewRecord{
		ownerID:        actorID,
		refFingerprint: fingerprint,
		result:         result,
		expiresAt:      expiresAt,
	}
	s.mu.Unlock()
	return chatPreviewFromResult(id, result, expiresAt), nil
}

// Confirm rechecks the exact reference through the session-owning resolver.
// It consumes the preview only after Telegram returns a durable source identity.
func (s *SourcePreviewService) Confirm(ctx context.Context, actorValue, previewValue, chatRef string) (ChatPreview, error) {
	if s == nil || s.resolver == nil {
		return ChatPreview{}, errors.New("Telegram 来源预览服务不可用")
	}
	if ctx == nil {
		return ChatPreview{}, errors.New("Telegram 来源确认 context 不能为空")
	}
	actorID, err := parseSourcePreviewActorID(actorValue)
	if err != nil {
		return ChatPreview{}, err
	}
	previewID, err := parseSourcePreviewID(previewValue)
	if err != nil {
		return ChatPreview{}, err
	}
	fingerprint, err := sourcePreviewFingerprint(chatRef)
	if err != nil {
		return ChatPreview{}, err
	}

	now := s.now()
	s.mu.Lock()
	s.expireLocked(now)
	record, ok := s.previews[previewID]
	s.mu.Unlock()
	if !ok {
		return ChatPreview{}, ErrSourcePreviewNotFound
	}
	if record.ownerID != actorID {
		return ChatPreview{}, ErrSourcePreviewNotOwner
	}
	if subtle.ConstantTimeCompare(record.refFingerprint[:], fingerprint[:]) != 1 {
		return ChatPreview{}, ErrSourcePreviewReferenceMismatch
	}

	confirmed, err := s.resolver.ConfirmChat(ctx, strings.TrimSpace(chatRef))
	if err != nil {
		return ChatPreview{}, err
	}
	if err := validateSourcePreviewResult(confirmed, true); err != nil {
		return ChatPreview{}, err
	}
	if record.result.ChatType != confirmed.ChatType {
		return ChatPreview{}, errors.New("Telegram 来源类型与预览不一致")
	}
	if record.result.ChatID != 0 && record.result.ChatID != confirmed.ChatID {
		return ChatPreview{}, errors.New("Telegram 来源 chat ID 与预览不一致")
	}

	s.mu.Lock()
	delete(s.previews, previewID)
	s.mu.Unlock()
	return chatPreviewFromResult(previewID, confirmed, record.expiresAt), nil
}

func parseSourcePreviewActorID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil || id == uuid.Nil {
		return uuid.Nil, errors.New("Telegram 来源预览操作者 ID 格式错误")
	}
	return id, nil
}

func parseSourcePreviewID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil || id == uuid.Nil {
		return uuid.Nil, errors.New("Telegram 来源预览 ID 格式错误")
	}
	return id, nil
}

func sourcePreviewFingerprint(chatRef string) ([sha256.Size]byte, error) {
	chatRef = strings.TrimSpace(chatRef)
	if chatRef == "" {
		return [sha256.Size]byte{}, errors.New("Telegram 来源引用不能为空")
	}
	return sha256.Sum256([]byte(chatRef)), nil
}

func validateSourcePreviewResult(result ChatPreviewResult, requireResolved bool) error {
	if strings.TrimSpace(result.Title) == "" {
		return errors.New("Telegram 来源预览缺少标题")
	}
	switch result.ChatType {
	case TelegramChatTypeGroup, TelegramChatTypeSupergroup, TelegramChatTypeChannel:
	default:
		return fmt.Errorf("Telegram 来源类型 %q 不受支持", result.ChatType)
	}
	if result.RequiresApproval && !result.RequiresJoin {
		return errors.New("Telegram 来源审批状态无效")
	}
	if result.ChatID == 0 && !result.RequiresJoin {
		return ErrSourcePreviewUnresolved
	}
	if requireResolved && (result.ChatID == 0 || result.RequiresJoin || result.RequiresApproval) {
		return ErrSourcePreviewUnresolved
	}
	return nil
}

func chatPreviewFromResult(id uuid.UUID, result ChatPreviewResult, expiresAt time.Time) ChatPreview {
	return ChatPreview{
		PreviewID:        id.String(),
		ChatID:           result.ChatID,
		Title:            strings.TrimSpace(result.Title),
		Username:         strings.TrimSpace(result.Username),
		ChatType:         result.ChatType,
		RequiresJoin:     result.RequiresJoin,
		RequiresApproval: result.RequiresApproval,
		ExpiresAt:        expiresAt.UTC().Format(time.RFC3339),
	}
}

func (s *SourcePreviewService) now() time.Time {
	if s == nil || s.clock == nil {
		return time.Now().UTC()
	}
	return s.clock().UTC()
}

func (s *SourcePreviewService) expireLocked(now time.Time) {
	for id, record := range s.previews {
		if !now.Before(record.expiresAt) {
			delete(s.previews, id)
		}
	}
}
