package telegram

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	// AuthorizationKindPhone starts an authorization with the configured phone number.
	AuthorizationKindPhone = "phone"
	// AuthorizationKindQR starts an authorization with a Telegram QR code.
	AuthorizationKindQR = "qr"
)

const (
	// AuthorizationStatusPending is the short internal startup state.
	AuthorizationStatusPending = "pending"
	// AuthorizationStatusAwaitingCode waits for the phone login code.
	AuthorizationStatusAwaitingCode = "awaiting_code"
	// AuthorizationStatusAwaitingPassword waits for the Telegram 2FA password.
	AuthorizationStatusAwaitingPassword = "awaiting_password"
	// AuthorizationStatusScanning waits for a QR scan and confirmation.
	AuthorizationStatusScanning = "scanning"
	// AuthorizationStatusSucceeded means the temporary authorization was accepted.
	AuthorizationStatusSucceeded = "succeeded"
	// AuthorizationStatusFailed means Telegram rejected or could not complete the flow.
	AuthorizationStatusFailed = "failed"
	// AuthorizationStatusCancelled means an administrator cancelled the flow.
	AuthorizationStatusCancelled = "cancelled"
	// AuthorizationStatusExpired means the short-lived flow passed its deadline.
	AuthorizationStatusExpired = "expired"
)

var (
	// ErrAuthorizationInProgress indicates that another authorization owns the singleton slot.
	ErrAuthorizationInProgress = errors.New("Telegram 授权正在进行中")
	// ErrAuthorizationNotFound indicates that an authorization ID is unknown or expired from memory.
	ErrAuthorizationNotFound = errors.New("Telegram 授权不存在")
	// ErrAuthorizationNotOwner prevents another administrator from submitting secrets.
	ErrAuthorizationNotOwner = errors.New("无权提交该 Telegram 授权")
	// ErrAuthorizationInvalidTransition indicates an illegal lifecycle transition.
	ErrAuthorizationInvalidTransition = errors.New("Telegram 授权状态转换无效")
)

// AuthorizationStart describes a new short-lived authorization request.
type AuthorizationStart struct {
	Kind    string
	OwnerID uuid.UUID
}

// AuthorizationView is the sanitized representation exposed to administrators.
// It intentionally has no code, password, login token, or QR token field.
type AuthorizationView struct {
	ID               uuid.UUID `json:"id"`
	Kind             string    `json:"kind"`
	Status           string    `json:"status"`
	OwnerID          uuid.UUID `json:"owner_id,omitempty"`
	ExpiresAt        time.Time `json:"expires_at"`
	QRImageDataURL   string    `json:"qr_image_data_url,omitempty"`
	QRImageExpiresAt time.Time `json:"qr_image_expires_at,omitempty"`
	Error            string    `json:"error,omitempty"`
	TelegramUserID   *int64    `json:"telegram_user_id,omitempty"`
	Username         string    `json:"username,omitempty"`
	FirstName        string    `json:"first_name,omitempty"`
	LastName         string    `json:"last_name,omitempty"`
}

type authorizationRecord struct {
	AuthorizationView
	startedAt time.Time
}

// AuthorizationStateMachine serializes the one active authorization flow.
// Transient secrets belong to the control service, not this state machine.
type AuthorizationStateMachine struct {
	mu      sync.Mutex
	clock   func() time.Time
	ttl     time.Duration
	active  *authorizationRecord
	records map[uuid.UUID]*authorizationRecord
}

// NewAuthorizationStateMachine constructs a thread-safe singleton state machine.
func NewAuthorizationStateMachine(clock func() time.Time, ttl time.Duration) *AuthorizationStateMachine {
	if clock == nil {
		clock = time.Now
	}
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &AuthorizationStateMachine{
		clock:   clock,
		ttl:     ttl,
		records: make(map[uuid.UUID]*authorizationRecord),
	}
}

// Start claims the singleton authorization slot for one administrator.
func (s *AuthorizationStateMachine) Start(input AuthorizationStart) (AuthorizationView, error) {
	if s == nil {
		return AuthorizationView{}, errors.New("Telegram 授权状态机不可用")
	}
	if input.OwnerID == uuid.Nil {
		return AuthorizationView{}, errors.New("Telegram 授权缺少操作者")
	}
	kind := strings.ToLower(strings.TrimSpace(input.Kind))
	if kind != AuthorizationKindPhone && kind != AuthorizationKindQR {
		return AuthorizationView{}, fmt.Errorf("不支持的 Telegram 授权方式 %q", input.Kind)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	s.expireActiveLocked(now)
	if s.active != nil {
		return AuthorizationView{}, ErrAuthorizationInProgress
	}
	status := AuthorizationStatusAwaitingCode
	if kind == AuthorizationKindQR {
		status = AuthorizationStatusScanning
	}
	record := &authorizationRecord{
		AuthorizationView: AuthorizationView{
			ID:        uuid.New(),
			Kind:      kind,
			Status:    status,
			OwnerID:   input.OwnerID,
			ExpiresAt: now.Add(s.ttl),
		},
		startedAt: now,
	}
	s.active = record
	s.records[record.ID] = record
	return record.view(), nil
}

// CanSubmit reports whether ownerID may submit a secret for authorizationID.
func (s *AuthorizationStateMachine) CanSubmit(authorizationID, ownerID uuid.UUID) bool {
	if s == nil || authorizationID == uuid.Nil || ownerID == uuid.Nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireActiveLocked(s.now())
	record := s.records[authorizationID]
	return record != nil && s.active == record && record.OwnerID == ownerID && isAuthorizationOpen(record.Status)
}

// Transition moves an authorization to a validated lifecycle state.
func (s *AuthorizationStateMachine) Transition(authorizationID uuid.UUID, status, errorMessage string) error {
	if s == nil {
		return errors.New("Telegram 授权状态机不可用")
	}
	status = strings.ToLower(strings.TrimSpace(status))
	if !isAuthorizationStatus(status) {
		return fmt.Errorf("%w: %s", ErrAuthorizationInvalidTransition, status)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if authorizationID == uuid.Nil {
		return ErrAuthorizationNotFound
	}
	record := s.records[authorizationID]
	if record == nil {
		return ErrAuthorizationNotFound
	}
	s.expireActiveLocked(s.now())
	if s.active != record && isAuthorizationOpen(record.Status) {
		return ErrAuthorizationNotFound
	}
	if !validAuthorizationTransition(record.Status, status) {
		return fmt.Errorf("%w: %s -> %s", ErrAuthorizationInvalidTransition, record.Status, status)
	}
	record.Status = status
	record.Error = sanitizeAuthorizationError(errorMessage)
	if isAuthorizationTerminal(status) && s.active == record {
		s.active = nil
	}
	return nil
}

// SetQRImage stores a short-lived rendered image without storing the raw QR token.
func (s *AuthorizationStateMachine) SetQRImage(authorizationID uuid.UUID, dataURL string, expiresAt time.Time) error {
	if s == nil {
		return errors.New("Telegram 授权状态机不可用")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	record := s.records[authorizationID]
	if record == nil {
		return ErrAuthorizationNotFound
	}
	s.expireActiveLocked(s.now())
	if s.active != record || !isAuthorizationOpen(record.Status) {
		return ErrAuthorizationNotFound
	}
	if strings.TrimSpace(dataURL) == "" {
		return errors.New("Telegram 二维码图像不能为空")
	}
	record.QRImageDataURL = strings.TrimSpace(dataURL)
	if !expiresAt.IsZero() {
		record.QRImageExpiresAt = expiresAt.UTC()
	}
	return nil
}

// SetTelegramIdentity attaches non-sensitive Telegram user metadata to a flow.
func (s *AuthorizationStateMachine) SetTelegramIdentity(authorizationID uuid.UUID, userID *int64, username, firstName, lastName string) error {
	if s == nil {
		return errors.New("Telegram 授权状态机不可用")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	record := s.records[authorizationID]
	if record == nil {
		return ErrAuthorizationNotFound
	}
	record.TelegramUserID = userID
	record.Username = strings.TrimSpace(username)
	record.FirstName = strings.TrimSpace(firstName)
	record.LastName = strings.TrimSpace(lastName)
	return nil
}

// View returns a sanitized view. The viewer argument is intentionally not used
// to hide progress: non-owners may observe status but cannot submit secrets.
func (s *AuthorizationStateMachine) View(authorizationID, _ uuid.UUID) (AuthorizationView, bool) {
	if s == nil || authorizationID == uuid.Nil {
		return AuthorizationView{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	record := s.records[authorizationID]
	if record == nil {
		return AuthorizationView{}, false
	}
	now := s.now()
	s.expireActiveLocked(now)
	s.expireQRImageLocked(record, now)
	return record.view(), true
}

// Cancel terminates an active authorization owned by ownerID.
func (s *AuthorizationStateMachine) Cancel(authorizationID, ownerID uuid.UUID) error {
	if !s.CanSubmit(authorizationID, ownerID) {
		return ErrAuthorizationNotOwner
	}
	return s.Transition(authorizationID, AuthorizationStatusCancelled, "")
}

func (s *AuthorizationStateMachine) now() time.Time {
	if s.clock == nil {
		return time.Now().UTC()
	}
	return s.clock().UTC()
}

func (s *AuthorizationStateMachine) expireActiveLocked(now time.Time) {
	if s.active == nil || !isAuthorizationOpen(s.active.Status) || now.Before(s.active.ExpiresAt) {
		return
	}
	s.active.Status = AuthorizationStatusExpired
	s.active.Error = ""
	s.active.QRImageDataURL = ""
	s.active.QRImageExpiresAt = time.Time{}
	s.active = nil
}

func (s *AuthorizationStateMachine) expireQRImageLocked(record *authorizationRecord, now time.Time) {
	if record == nil || record.QRImageExpiresAt.IsZero() || now.Before(record.QRImageExpiresAt) {
		return
	}
	record.QRImageDataURL = ""
	record.QRImageExpiresAt = time.Time{}
}

func (r *authorizationRecord) view() AuthorizationView {
	view := r.AuthorizationView
	if view.Status == AuthorizationStatusExpired {
		view.QRImageDataURL = ""
		view.QRImageExpiresAt = time.Time{}
	}
	return view
}

func isAuthorizationStatus(status string) bool {
	switch status {
	case AuthorizationStatusPending, AuthorizationStatusAwaitingCode, AuthorizationStatusAwaitingPassword,
		AuthorizationStatusScanning, AuthorizationStatusSucceeded, AuthorizationStatusFailed,
		AuthorizationStatusCancelled, AuthorizationStatusExpired:
		return true
	default:
		return false
	}
}

func isAuthorizationOpen(status string) bool {
	switch status {
	case AuthorizationStatusPending, AuthorizationStatusAwaitingCode, AuthorizationStatusAwaitingPassword, AuthorizationStatusScanning:
		return true
	default:
		return false
	}
}

func isAuthorizationTerminal(status string) bool {
	switch status {
	case AuthorizationStatusSucceeded, AuthorizationStatusFailed, AuthorizationStatusCancelled, AuthorizationStatusExpired:
		return true
	default:
		return false
	}
}

func validAuthorizationTransition(from, to string) bool {
	if from == to {
		return true
	}
	switch from {
	case AuthorizationStatusPending:
		return to == AuthorizationStatusAwaitingCode || to == AuthorizationStatusScanning || isAuthorizationTerminal(to)
	case AuthorizationStatusAwaitingCode:
		return to == AuthorizationStatusAwaitingPassword || to == AuthorizationStatusSucceeded || isAuthorizationTerminal(to)
	case AuthorizationStatusAwaitingPassword:
		return to == AuthorizationStatusSucceeded || isAuthorizationTerminal(to)
	case AuthorizationStatusScanning:
		return to == AuthorizationStatusSucceeded || isAuthorizationTerminal(to)
	default:
		return false
	}
}

func sanitizeAuthorizationError(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len([]rune(value)) > 500 {
		return string([]rune(value)[:500])
	}
	return value
}
