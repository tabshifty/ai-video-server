package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"video-server/internal/models"
	"video-server/internal/repository"
)

var (
	// ErrAuthorizationPhoneBindingRequired prevents QR login before a phone flow
	// establishes the account identity that later QR authorizations must match.
	ErrAuthorizationPhoneBindingRequired = errors.New("首次绑定 Telegram 账号后才能使用二维码授权")
)

// TelegramIdentity is the non-sensitive identity returned by Telegram after a
// successful authorization.
type TelegramIdentity struct {
	ID        int64
	Username  string
	FirstName string
	LastName  string
}

// AuthorizationSessionUpdate reports public authorization progress. It must
// never contain a phone code, password, raw QR token, or invitation token.
type AuthorizationSessionUpdate struct {
	Status         string
	QRImageDataURL string
	QRExpiresAt    time.Time
	Identity       *TelegramIdentity
	Err            error
}

// AuthorizationSession owns one temporary Telegram session. Promote makes the
// validated temporary session the persistent collector session; Discard removes
// it without touching the current persistent session.
type AuthorizationSession interface {
	Start(ctx context.Context, kind string, notify func(AuthorizationSessionUpdate)) error
	SubmitCode(ctx context.Context, code string) error
	SubmitPassword(ctx context.Context, password string) error
	Promote(ctx context.Context) error
	Discard() error
	Close()
}

// AuthorizationSessionFactory creates an isolated temporary session per flow.
type AuthorizationSessionFactory interface {
	NewAuthorizationSession(authorizationID uuid.UUID) (AuthorizationSession, error)
}

// AuthorizationMaintenance pauses the live ingestion runtime before its
// session can be replaced and restores it after a terminal outcome.
type AuthorizationMaintenance interface {
	Pause(ctx context.Context) error
	Resume()
}

// AuthorizationRepository contains only the durable, non-sensitive records
// managed by an authorization flow.
type AuthorizationRepository interface {
	GetTelegramAccountState(ctx context.Context) (models.TelegramAccountState, error)
	UpsertTelegramAccountState(ctx context.Context, account models.TelegramAccountState) error
	CreateTelegramAuthorization(ctx context.Context, item models.TelegramAuthorization) error
	UpdateTelegramAuthorization(ctx context.Context, id uuid.UUID, status, errorSummary string, telegramUserID *int64, finishedAt *time.Time) error
	CreateTelegramAuditLog(ctx context.Context, item models.TelegramAuditLog) error
}

// AuthorizationServiceConfig wires the process-local authorization owner.
type AuthorizationServiceConfig struct {
	Phone       string
	Repository  AuthorizationRepository
	Sessions    AuthorizationSessionFactory
	Maintenance AuthorizationMaintenance
	State       *AuthorizationStateMachine
	Clock       func() time.Time
}

// AuthorizationService coordinates the short-lived flow with durable account
// metadata. Secrets pass straight to AuthorizationSession and are not retained.
type AuthorizationService struct {
	phone       string
	repository  AuthorizationRepository
	sessions    AuthorizationSessionFactory
	maintenance AuthorizationMaintenance
	state       *AuthorizationStateMachine
	clock       func() time.Time

	mu     sync.Mutex
	active map[uuid.UUID]*activeAuthorization
}

type activeAuthorization struct {
	id                uuid.UUID
	ownerID           uuid.UUID
	kind              string
	previousAccount   models.TelegramAccountState
	accountWasMissing bool
	boundTelegramID   *int64
	session           AuthorizationSession
	paused            bool
	finishing         bool
	terminal          bool
}

// NewAuthorizationService constructs the process-local authorization
// coordinator. Invalid dependencies are surfaced by the individual operation
// so process construction can remain straightforward.
func NewAuthorizationService(config AuthorizationServiceConfig) *AuthorizationService {
	clock := config.Clock
	if clock == nil {
		clock = time.Now
	}
	state := config.State
	if state == nil {
		state = NewAuthorizationStateMachine(clock, 10*time.Minute)
	}
	return &AuthorizationService{
		phone:       strings.TrimSpace(config.Phone),
		repository:  config.Repository,
		sessions:    config.Sessions,
		maintenance: config.Maintenance,
		state:       state,
		clock:       clock,
		active:      make(map[uuid.UUID]*activeAuthorization),
	}
}

// StartPhone starts the first binding or a phone-based reauthorization.
func (s *AuthorizationService) StartPhone(ctx context.Context, actorID string) (AuthorizationView, error) {
	return s.start(ctx, actorID, AuthorizationKindPhone)
}

// StartQR starts a QR-based reauthorization. QR login is deliberately blocked
// until a phone flow has established the expected Telegram user ID.
func (s *AuthorizationService) StartQR(ctx context.Context, actorID string) (AuthorizationView, error) {
	return s.start(ctx, actorID, AuthorizationKindQR)
}

func (s *AuthorizationService) start(ctx context.Context, actorValue, kind string) (AuthorizationView, error) {
	if s == nil || s.state == nil {
		return AuthorizationView{}, errors.New("Telegram 授权服务不可用")
	}
	if ctx == nil {
		return AuthorizationView{}, errors.New("Telegram 授权 context 不能为空")
	}
	if s.repository == nil {
		return AuthorizationView{}, errors.New("Telegram 授权仓储不可用")
	}
	if s.sessions == nil {
		return AuthorizationView{}, errors.New("Telegram 临时 session 工厂不可用")
	}
	actorID, err := parseAuthorizationActorID(actorValue)
	if err != nil {
		return AuthorizationView{}, err
	}

	previousAccount, accountWasMissing, err := s.accountState(ctx)
	if err != nil {
		return AuthorizationView{}, err
	}
	if kind == AuthorizationKindQR && previousAccount.TelegramUserID == nil {
		return AuthorizationView{}, ErrAuthorizationPhoneBindingRequired
	}

	view, err := s.state.Start(AuthorizationStart{Kind: kind, OwnerID: actorID})
	if err != nil {
		return AuthorizationView{}, err
	}

	active := &activeAuthorization{
		id:                view.ID,
		ownerID:           actorID,
		kind:              kind,
		previousAccount:   previousAccount,
		accountWasMissing: accountWasMissing,
		boundTelegramID:   cloneInt64(previousAccount.TelegramUserID),
	}
	if err := s.createAuthorizationRecord(ctx, view); err != nil {
		_ = s.state.Transition(view.ID, AuthorizationStatusFailed, "无法创建授权记录")
		return AuthorizationView{}, err
	}

	account := authorizationAccount(previousAccount, kind, s.phone, s.now())
	if err := s.repository.UpsertTelegramAccountState(ctx, account); err != nil {
		s.finishFailure(ctx, active, "无法更新 Telegram 账号状态", false)
		return AuthorizationView{}, fmt.Errorf("更新 Telegram 账号授权状态: %w", err)
	}
	if err := s.audit(ctx, actorID, "authorization.started", view.ID.String(), "succeeded", map[string]any{"kind": kind}); err != nil {
		s.finishFailure(ctx, active, "无法写入授权审计", false)
		return AuthorizationView{}, err
	}

	if s.maintenance != nil {
		if err := s.maintenance.Pause(ctx); err != nil {
			s.finishFailure(ctx, active, "无法暂停 Telegram 采集", false)
			return AuthorizationView{}, fmt.Errorf("暂停 Telegram 采集器: %w", err)
		}
		active.paused = true
	}

	session, err := s.sessions.NewAuthorizationSession(view.ID)
	if err != nil {
		s.finishFailure(ctx, active, "无法创建 Telegram 临时 session", true)
		return AuthorizationView{}, fmt.Errorf("创建 Telegram 临时 session: %w", err)
	}
	active.session = session
	s.mu.Lock()
	s.active[view.ID] = active
	s.mu.Unlock()

	if err := session.Start(ctx, kind, func(update AuthorizationSessionUpdate) {
		s.handleSessionUpdate(view.ID, update)
	}); err != nil {
		s.finishFailure(ctx, active, "Telegram 授权启动失败", true)
		return AuthorizationView{}, fmt.Errorf("启动 Telegram 授权: %w", err)
	}
	return s.view(view.ID, actorID)
}

// SubmitCode forwards a one-time phone code to the in-memory temporary
// session. The code is intentionally never placed in a struct or audit record.
func (s *AuthorizationService) SubmitCode(ctx context.Context, authorizationID, actorID, code string) (AuthorizationView, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return AuthorizationView{}, errors.New("Telegram 验证码不能为空")
	}
	active, owner, err := s.submittableAuthorization(authorizationID, actorID)
	if err != nil {
		return AuthorizationView{}, err
	}
	if err := active.session.SubmitCode(ctx, code); err != nil {
		s.finishFailure(ctx, active, "Telegram 验证码验证失败", true)
		return AuthorizationView{}, errors.New("Telegram 验证码验证失败")
	}
	return s.view(active.id, owner)
}

// SubmitPassword forwards a Telegram 2FA password to the in-memory temporary
// session. Password whitespace is meaningful and is preserved.
func (s *AuthorizationService) SubmitPassword(ctx context.Context, authorizationID, actorID, password string) (AuthorizationView, error) {
	if strings.TrimSpace(password) == "" {
		return AuthorizationView{}, errors.New("Telegram 二次验证密码不能为空")
	}
	active, owner, err := s.submittableAuthorization(authorizationID, actorID)
	if err != nil {
		return AuthorizationView{}, err
	}
	if err := active.session.SubmitPassword(ctx, password); err != nil {
		s.finishFailure(ctx, active, "Telegram 二次验证失败", true)
		return AuthorizationView{}, errors.New("Telegram 二次验证失败")
	}
	return s.view(active.id, owner)
}

// GetAuthorization returns the sanitized progress view for any administrator.
// Only the initiating administrator may submit a code or password.
func (s *AuthorizationService) GetAuthorization(ctx context.Context, authorizationID, actorID string) (AuthorizationView, error) {
	if ctx == nil {
		return AuthorizationView{}, errors.New("Telegram 授权 context 不能为空")
	}
	id, err := ParseControlAuthorizationID(authorizationID)
	if err != nil {
		return AuthorizationView{}, err
	}
	owner, err := parseAuthorizationActorID(actorID)
	if err != nil {
		return AuthorizationView{}, err
	}
	view, err := s.view(id, owner)
	if err != nil {
		return AuthorizationView{}, err
	}
	if view.Status == AuthorizationStatusExpired {
		s.mu.Lock()
		active := s.active[id]
		s.mu.Unlock()
		if active != nil {
			s.finishTerminal(ctx, active, AuthorizationStatusExpired, "", nil, true)
			view, _ = s.view(id, owner)
		}
	}
	return view, nil
}

// ActiveAuthorization returns the single in-progress authorization for the
// control status projection. Any authenticated administrator may view its
// sanitized state, but only its owner can submit secrets through the existing
// ownership checks.
func (s *AuthorizationService) ActiveAuthorization(ctx context.Context, actorID string) (*AuthorizationView, error) {
	if s == nil || s.state == nil {
		return nil, errors.New("Telegram 授权服务不可用")
	}
	if ctx == nil {
		return nil, errors.New("Telegram 授权 context 不能为空")
	}
	observerID, err := parseAuthorizationActorID(actorID)
	if err != nil {
		return nil, err
	}
	view, ok := s.state.Active()
	if !ok {
		return nil, nil
	}
	if view.Status == AuthorizationStatusExpired {
		s.mu.Lock()
		active := s.active[view.ID]
		s.mu.Unlock()
		if active != nil {
			if err := s.finishTerminal(ctx, active, AuthorizationStatusExpired, "", nil, true); err != nil {
				return nil, err
			}
			view, err = s.view(view.ID, observerID)
			if err != nil {
				return nil, err
			}
		}
	}
	return &view, nil
}

// Cancel stops an active authorization. It does not modify an existing bound
// account or its persistent session.
func (s *AuthorizationService) Cancel(ctx context.Context, authorizationID, actorID string) error {
	if ctx == nil {
		return errors.New("Telegram 授权 context 不能为空")
	}
	active, _, err := s.submittableAuthorization(authorizationID, actorID)
	if err != nil {
		return err
	}
	return s.finishTerminal(ctx, active, AuthorizationStatusCancelled, "", nil, true)
}

func (s *AuthorizationService) submittableAuthorization(authorizationValue, actorValue string) (*activeAuthorization, uuid.UUID, error) {
	if s == nil || s.state == nil {
		return nil, uuid.Nil, errors.New("Telegram 授权服务不可用")
	}
	id, err := ParseControlAuthorizationID(authorizationValue)
	if err != nil {
		return nil, uuid.Nil, err
	}
	owner, err := parseAuthorizationActorID(actorValue)
	if err != nil {
		return nil, uuid.Nil, err
	}
	view, exists := s.state.View(id, owner)
	if !exists {
		return nil, uuid.Nil, ErrAuthorizationNotFound
	}
	if view.OwnerID != owner {
		return nil, uuid.Nil, ErrAuthorizationNotOwner
	}
	if !s.state.CanSubmit(id, owner) {
		return nil, uuid.Nil, ErrAuthorizationNotFound
	}
	s.mu.Lock()
	active := s.active[id]
	s.mu.Unlock()
	if active == nil || active.session == nil {
		return nil, uuid.Nil, ErrAuthorizationNotFound
	}
	return active, owner, nil
}

func (s *AuthorizationService) handleSessionUpdate(authorizationID uuid.UUID, update AuthorizationSessionUpdate) {
	s.mu.Lock()
	active := s.active[authorizationID]
	if active == nil || active.terminal || active.finishing {
		s.mu.Unlock()
		return
	}
	if update.Identity != nil || update.Err != nil {
		active.finishing = true
	}
	s.mu.Unlock()
	if active == nil {
		return
	}

	ctx := context.Background()
	if update.Err != nil {
		s.finishFailure(ctx, active, "Telegram 授权失败", true)
		return
	}
	if strings.TrimSpace(update.QRImageDataURL) != "" {
		if err := s.state.SetQRImage(authorizationID, update.QRImageDataURL, update.QRExpiresAt); err != nil {
			s.finishFailure(ctx, active, "无法更新 Telegram 二维码", true)
		}
		return
	}
	if update.Identity != nil {
		s.finishSuccess(ctx, active, *update.Identity)
		return
	}
	if status := strings.TrimSpace(update.Status); status != "" {
		if err := s.state.Transition(authorizationID, status, ""); err != nil {
			s.finishFailure(ctx, active, "Telegram 授权状态异常", true)
			return
		}
		if err := s.updateAuthorization(ctx, authorizationID, status, "", nil, nil); err != nil {
			s.finishFailure(ctx, active, "无法更新 Telegram 授权状态", true)
		}
	}
}

func (s *AuthorizationService) finishSuccess(ctx context.Context, active *activeAuthorization, identity TelegramIdentity) {
	if identity.ID <= 0 {
		s.finishFailure(ctx, active, "Telegram 账号标识无效", true)
		return
	}
	if active.boundTelegramID != nil && *active.boundTelegramID != identity.ID {
		s.finishFailure(ctx, active, "二维码授权账号与已绑定 Telegram 账号不一致", true)
		return
	}
	if err := s.state.SetTelegramIdentity(active.id, &identity.ID, identity.Username, identity.FirstName, identity.LastName); err != nil {
		s.finishFailure(ctx, active, "无法记录 Telegram 账号身份", true)
		return
	}
	if active.session == nil {
		s.finishFailure(ctx, active, "Telegram 临时 session 不可用", true)
		return
	}
	if err := active.session.Promote(ctx); err != nil {
		s.finishFailure(ctx, active, "无法替换 Telegram session", true)
		return
	}

	now := s.now()
	account := active.previousAccount
	account.ID = 1
	account.TelegramUserID = cloneInt64(&identity.ID)
	account.Username = strings.TrimSpace(identity.Username)
	account.FirstName = strings.TrimSpace(identity.FirstName)
	account.LastName = strings.TrimSpace(identity.LastName)
	if account.PhoneMasked == "" {
		account.PhoneMasked = maskTelegramPhone(s.phone)
	}
	account.Status = models.TelegramAccountStatusAuthorized
	account.LastError = ""
	account.AuthorizedAt = &now
	account.LastSeenAt = &now
	account.UpdatedAt = now
	if account.CreatedAt.IsZero() {
		account.CreatedAt = now
	}
	if err := s.repository.UpsertTelegramAccountState(ctx, account); err != nil {
		s.finishFailure(ctx, active, "无法保存 Telegram 授权账号", false)
		return
	}
	if err := s.finishTerminal(ctx, active, AuthorizationStatusSucceeded, "", &identity.ID, false); err != nil {
		return
	}
	_ = s.audit(ctx, active.ownerID, "authorization.succeeded", active.id.String(), "succeeded", map[string]any{
		"kind":             active.kind,
		"telegram_user_id": identity.ID,
	})
}

func (s *AuthorizationService) finishFailure(ctx context.Context, active *activeAuthorization, summary string, discard bool) {
	if active == nil {
		return
	}
	_ = s.finishTerminal(ctx, active, AuthorizationStatusFailed, summary, nil, discard)
}

func (s *AuthorizationService) finishTerminal(ctx context.Context, active *activeAuthorization, status, summary string, telegramUserID *int64, discard bool) error {
	if active == nil {
		return ErrAuthorizationNotFound
	}
	s.mu.Lock()
	alreadyTerminal := active.terminal
	active.terminal = true
	delete(s.active, active.id)
	s.mu.Unlock()
	if alreadyTerminal {
		return nil
	}
	if active.paused && s.maintenance != nil {
		defer s.maintenance.Resume()
	}

	if status != AuthorizationStatusSucceeded {
		if active.session != nil {
			active.session.Close()
			if discard {
				_ = active.session.Discard()
			}
		}
		if err := s.restoreAccount(ctx, active); err != nil {
			summary = chooseAuthorizationSummary(summary, "无法恢复 Telegram 账号状态")
		}
	}

	if err := s.state.Transition(active.id, status, summary); err != nil && !errors.Is(err, ErrAuthorizationNotFound) {
		return err
	}
	now := s.now()
	if err := s.updateAuthorization(ctx, active.id, status, summary, telegramUserID, &now); err != nil {
		return err
	}
	result := "failed"
	if status == AuthorizationStatusCancelled {
		result = "cancelled"
	} else if status == AuthorizationStatusSucceeded {
		result = "succeeded"
	}
	action := "authorization.failed"
	if status == AuthorizationStatusCancelled {
		action = "authorization.cancelled"
	} else if status == AuthorizationStatusExpired {
		action = "authorization.expired"
	}
	if status != AuthorizationStatusSucceeded {
		_ = s.audit(ctx, active.ownerID, action, active.id.String(), result, map[string]any{"kind": active.kind})
	}
	return nil
}

func (s *AuthorizationService) restoreAccount(ctx context.Context, active *activeAuthorization) error {
	if s.repository == nil {
		return errors.New("Telegram 授权仓储不可用")
	}
	account := active.previousAccount
	now := s.now()
	if active.accountWasMissing {
		account = models.TelegramAccountState{
			ID:           1,
			Status:       models.TelegramAccountStatusUnconfigured,
			PhoneMasked:  maskTelegramPhone(s.phone),
			CreatedAt:    now,
			UpdatedAt:    now,
			LastSeenAt:   nil,
			AuthorizedAt: nil,
		}
	} else {
		account.ID = 1
		account.UpdatedAt = now
		if account.CreatedAt.IsZero() {
			account.CreatedAt = now
		}
	}
	return s.repository.UpsertTelegramAccountState(ctx, account)
}

func (s *AuthorizationService) accountState(ctx context.Context) (models.TelegramAccountState, bool, error) {
	account, err := s.repository.GetTelegramAccountState(ctx)
	if err == nil {
		return account, false, nil
	}
	if errors.Is(err, repository.ErrTelegramAccountStateNotFound) {
		return models.TelegramAccountState{}, true, nil
	}
	return models.TelegramAccountState{}, false, fmt.Errorf("读取 Telegram 账号状态: %w", err)
}

func (s *AuthorizationService) createAuthorizationRecord(ctx context.Context, view AuthorizationView) error {
	now := s.now()
	return s.repository.CreateTelegramAuthorization(ctx, models.TelegramAuthorization{
		ID:          view.ID,
		Kind:        view.Kind,
		Status:      view.Status,
		ActorUserID: view.OwnerID,
		StartedAt:   now,
		ExpiresAt:   view.ExpiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
}

func (s *AuthorizationService) updateAuthorization(ctx context.Context, id uuid.UUID, status, summary string, telegramUserID *int64, finishedAt *time.Time) error {
	if s.repository == nil {
		return errors.New("Telegram 授权仓储不可用")
	}
	if err := s.repository.UpdateTelegramAuthorization(ctx, id, status, sanitizeAuthorizationError(summary), cloneInt64(telegramUserID), finishedAt); err != nil {
		return fmt.Errorf("更新 Telegram 授权记录: %w", err)
	}
	return nil
}

func (s *AuthorizationService) audit(ctx context.Context, actorID uuid.UUID, action, targetID, result string, summary map[string]any) error {
	if s.repository == nil {
		return errors.New("Telegram 授权仓储不可用")
	}
	payload, err := json.Marshal(summary)
	if err != nil {
		return fmt.Errorf("编码 Telegram 授权审计: %w", err)
	}
	if err := s.repository.CreateTelegramAuditLog(ctx, models.TelegramAuditLog{
		ActorUserID: actorID,
		Action:      action,
		TargetType:  "authorization",
		TargetID:    targetID,
		Result:      result,
		Summary:     payload,
		CreatedAt:   s.now(),
	}); err != nil {
		return fmt.Errorf("写入 Telegram 授权审计: %w", err)
	}
	return nil
}

func (s *AuthorizationService) view(id, ownerID uuid.UUID) (AuthorizationView, error) {
	if s == nil || s.state == nil {
		return AuthorizationView{}, errors.New("Telegram 授权服务不可用")
	}
	view, ok := s.state.View(id, ownerID)
	if !ok {
		return AuthorizationView{}, ErrAuthorizationNotFound
	}
	return view, nil
}

func (s *AuthorizationService) now() time.Time {
	if s == nil || s.clock == nil {
		return time.Now().UTC()
	}
	return s.clock().UTC()
}

func authorizationAccount(previous models.TelegramAccountState, kind, phone string, now time.Time) models.TelegramAccountState {
	account := previous
	account.ID = 1
	if account.TelegramUserID == nil {
		account.Status = models.TelegramAccountStatusAuthorizing
		account.PhoneMasked = maskTelegramPhone(phone)
	} else {
		account.Status = models.TelegramAccountStatusReauthorizing
	}
	if kind == AuthorizationKindPhone && account.PhoneMasked == "" {
		account.PhoneMasked = maskTelegramPhone(phone)
	}
	account.LastError = ""
	account.UpdatedAt = now
	if account.CreatedAt.IsZero() {
		account.CreatedAt = now
	}
	return account
}

func parseAuthorizationActorID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil || id == uuid.Nil {
		return uuid.Nil, errors.New("Telegram 授权操作者 ID 格式错误")
	}
	return id, nil
}

func maskTelegramPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return ""
	}
	runes := []rune(phone)
	if len(runes) <= 4 {
		return strings.Repeat("*", len(runes))
	}
	visiblePrefix := 4
	if visiblePrefix >= len(runes) {
		visiblePrefix = 1
	}
	visibleSuffix := 4
	if visiblePrefix+visibleSuffix >= len(runes) {
		visibleSuffix = 1
	}
	return string(runes[:visiblePrefix]) + "****" + string(runes[len(runes)-visibleSuffix:])
}

func cloneInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func chooseAuthorizationSummary(current, fallback string) string {
	if strings.TrimSpace(current) != "" {
		return current
	}
	return fallback
}
