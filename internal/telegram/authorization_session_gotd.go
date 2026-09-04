package telegram

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	gotdtelegram "github.com/gotd/td/telegram"
	gotdauth "github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/auth/qrlogin"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
	"rsc.io/qr"
)

// GotdAuthorizationSessionFactoryConfig configures temporary gotd sessions.
// SessionPath is the long-lived collector session that a validated temporary
// session atomically replaces.
type GotdAuthorizationSessionFactoryConfig struct {
	APIID       int
	APIHash     string
	Phone       string
	SessionPath string
	Logger      *zap.Logger
}

// GotdAuthorizationSessionFactory creates one isolated gotd client per
// authorization attempt. Temporary session files live beside the persistent
// session so promotion can use an atomic rename.
type GotdAuthorizationSessionFactory struct {
	apiID       int
	apiHash     string
	phone       string
	sessionPath string
	logger      *zap.Logger
}

// NewGotdAuthorizationSessionFactory validates the fixed Telegram credentials
// and prepares the directory used by both persistent and temporary sessions.
func NewGotdAuthorizationSessionFactory(config GotdAuthorizationSessionFactoryConfig) (*GotdAuthorizationSessionFactory, error) {
	if config.APIID <= 0 {
		return nil, errors.New("Telegram API ID 必须为正数")
	}
	if strings.TrimSpace(config.APIHash) == "" {
		return nil, errors.New("Telegram API hash 不能为空")
	}
	if strings.TrimSpace(config.Phone) == "" {
		return nil, errors.New("Telegram 电话号码不能为空")
	}
	sessionPath := strings.TrimSpace(config.SessionPath)
	if sessionPath == "" {
		return nil, errors.New("Telegram session 路径不能为空")
	}
	sessionPath = filepath.Clean(sessionPath)
	if err := os.MkdirAll(filepath.Dir(sessionPath), 0o700); err != nil {
		return nil, fmt.Errorf("创建 Telegram session 目录: %w", err)
	}
	logger := config.Logger
	if logger == nil {
		logger = zap.NewNop()
	}
	return &GotdAuthorizationSessionFactory{
		apiID:       config.APIID,
		apiHash:     strings.TrimSpace(config.APIHash),
		phone:       strings.TrimSpace(config.Phone),
		sessionPath: sessionPath,
		logger:      logger,
	}, nil
}

// NewAuthorizationSession creates a session file name that cannot collide with
// a different authorization flow. The service state machine still limits the
// process to one active flow.
func (f *GotdAuthorizationSessionFactory) NewAuthorizationSession(authorizationID uuid.UUID) (AuthorizationSession, error) {
	if f == nil {
		return nil, errors.New("Telegram 临时 session 工厂不可用")
	}
	if authorizationID == uuid.Nil {
		return nil, errors.New("Telegram 授权 ID 不能为空")
	}
	temporaryPath := temporaryAuthorizationSessionPath(f.sessionPath, authorizationID)
	if err := os.Remove(temporaryPath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("清理遗留 Telegram 临时 session: %w", err)
	}
	return &gotdAuthorizationSession{
		apiID:                 f.apiID,
		apiHash:               f.apiHash,
		phone:                 f.phone,
		persistentSessionPath: f.sessionPath,
		temporarySessionPath:  temporaryPath,
		logger:                f.logger,
		codeInput:             make(chan string, 1),
		passwordInput:         make(chan string, 1),
		startResult:           make(chan error, 1),
	}, nil
}

type authorizationSessionStage string

const (
	authorizationSessionStarting          authorizationSessionStage = "starting"
	authorizationSessionAwaitingCode      authorizationSessionStage = "awaiting_code"
	authorizationSessionVerifyingCode     authorizationSessionStage = "verifying_code"
	authorizationSessionAwaitingPassword  authorizationSessionStage = "awaiting_password"
	authorizationSessionVerifyingPassword authorizationSessionStage = "verifying_password"
	authorizationSessionScanning          authorizationSessionStage = "scanning"
	authorizationSessionAuthorized        authorizationSessionStage = "authorized"
)

// gotdAuthorizationSession owns no durable account metadata. Its only durable
// side effect is writing a temporary gotd session file that Promote can rename.
type gotdAuthorizationSession struct {
	apiID                 int
	apiHash               string
	phone                 string
	persistentSessionPath string
	temporarySessionPath  string
	logger                *zap.Logger

	mu            sync.Mutex
	startOnce     sync.Once
	started       bool
	closed        bool
	authorized    bool
	stage         authorizationSessionStage
	notify        func(AuthorizationSessionUpdate)
	cancel        context.CancelFunc
	runDone       chan struct{}
	startResult   chan error
	codeInput     chan string
	passwordInput chan string
}

var _ AuthorizationSession = (*gotdAuthorizationSession)(nil)
var _ AuthorizationSessionFactory = (*GotdAuthorizationSessionFactory)(nil)

// Start connects the temporary client and announces the first public step of
// the selected flow. Its background context intentionally outlives the HTTP
// request that initiated the authorization.
func (s *gotdAuthorizationSession) Start(ctx context.Context, kind string, notify func(AuthorizationSessionUpdate)) error {
	if s == nil {
		return errors.New("Telegram 临时 session 不可用")
	}
	if ctx == nil {
		return errors.New("Telegram 授权 context 不能为空")
	}
	if notify == nil {
		return errors.New("Telegram 授权状态回调不能为空")
	}
	kind = strings.ToLower(strings.TrimSpace(kind))
	if kind != AuthorizationKindPhone && kind != AuthorizationKindQR {
		return fmt.Errorf("不支持的 Telegram 授权方式 %q", kind)
	}

	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return errors.New("Telegram 临时 session 已启动")
	}
	runCtx, cancel := context.WithCancel(context.Background())
	s.started = true
	s.stage = authorizationSessionStarting
	s.notify = notify
	s.cancel = cancel
	s.runDone = make(chan struct{})
	s.mu.Unlock()

	dispatcher := tg.NewUpdateDispatcher()
	options := gotdtelegram.Options{
		Logger:         s.logger,
		SessionStorage: &gotdtelegram.FileSessionStorage{Path: s.temporarySessionPath},
		NoUpdates:      kind != AuthorizationKindQR,
	}
	var loggedIn qrlogin.LoggedIn
	if kind == AuthorizationKindQR {
		loggedIn = qrlogin.OnLoginToken(dispatcher)
		options.UpdateHandler = dispatcher
	}
	raw := gotdtelegram.NewClient(s.apiID, s.apiHash, options)

	go s.run(runCtx, raw, kind, loggedIn)
	select {
	case err := <-s.startResult:
		return err
	case <-ctx.Done():
		s.Close()
		return ctx.Err()
	}
}

// SubmitCode transfers one phone code to the background flow without retaining
// it in durable state or logs.
func (s *gotdAuthorizationSession) SubmitCode(ctx context.Context, code string) error {
	if ctx == nil {
		return errors.New("Telegram 授权 context 不能为空")
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return errors.New("Telegram 验证码不能为空")
	}
	if !s.beginSecretSubmission(authorizationSessionAwaitingCode, authorizationSessionVerifyingCode) {
		return errors.New("Telegram 授权当前不接受验证码")
	}
	select {
	case s.codeInput <- code:
		return nil
	case <-ctx.Done():
		s.setStage(authorizationSessionAwaitingCode)
		return ctx.Err()
	}
}

// SubmitPassword transfers one 2FA password to the background flow. Leading
// and trailing spaces are preserved because they may be part of the password.
func (s *gotdAuthorizationSession) SubmitPassword(ctx context.Context, password string) error {
	if ctx == nil {
		return errors.New("Telegram 授权 context 不能为空")
	}
	if strings.TrimSpace(password) == "" {
		return errors.New("Telegram 二次验证密码不能为空")
	}
	if !s.beginSecretSubmission(authorizationSessionAwaitingPassword, authorizationSessionVerifyingPassword) {
		return errors.New("Telegram 授权当前不接受二次验证密码")
	}
	select {
	case s.passwordInput <- password:
		return nil
	case <-ctx.Done():
		s.setStage(authorizationSessionAwaitingPassword)
		return ctx.Err()
	}
}

// Promote stops the temporary client before atomically replacing the persistent
// session file. The runtime maintenance gate has already stopped the live
// collector client before this method is called.
func (s *gotdAuthorizationSession) Promote(ctx context.Context) error {
	if ctx == nil {
		return errors.New("Telegram 授权 context 不能为空")
	}
	s.mu.Lock()
	authorized := s.authorized
	s.mu.Unlock()
	if !authorized {
		return errors.New("Telegram 临时 session 尚未完成授权")
	}
	if err := s.stopAndWait(ctx); err != nil {
		return err
	}
	return promoteAuthorizationSession(s.temporarySessionPath, s.persistentSessionPath)
}

// Discard stops the temporary client and removes only its temporary file.
func (s *gotdAuthorizationSession) Discard() error {
	if s == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := s.stopAndWait(ctx); err != nil {
		return err
	}
	if err := os.Remove(s.temporarySessionPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除 Telegram 临时 session: %w", err)
	}
	return nil
}

// Close requests shutdown without blocking the caller. Discard and Promote wait
// for shutdown before touching the temporary session file.
func (s *gotdAuthorizationSession) Close() {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.closed = true
	cancel := s.cancel
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (s *gotdAuthorizationSession) run(ctx context.Context, raw *gotdtelegram.Client, kind string, loggedIn qrlogin.LoggedIn) {
	err := raw.Run(ctx, func(runCtx context.Context) error {
		switch kind {
		case AuthorizationKindPhone:
			return s.runPhoneAuthorization(runCtx, raw)
		case AuthorizationKindQR:
			return s.runQRAuthorization(runCtx, raw, loggedIn)
		default:
			return errors.New("不支持的 Telegram 授权方式")
		}
	})
	s.signalStart(err)

	s.mu.Lock()
	closed := s.closed
	authorized := s.authorized
	notify := s.notify
	done := s.runDone
	s.mu.Unlock()
	if err != nil && !closed && !authorized && notify != nil {
		go notify(AuthorizationSessionUpdate{Err: errors.New("Telegram 授权会话已结束")})
	}
	if done != nil {
		close(done)
	}
}

func (s *gotdAuthorizationSession) runPhoneAuthorization(ctx context.Context, raw *gotdtelegram.Client) error {
	sentCode, err := raw.Auth().SendCode(ctx, s.phone, gotdauth.SendCodeOptions{})
	if err != nil {
		return fmt.Errorf("发送 Telegram 验证码: %w", err)
	}
	switch result := sentCode.(type) {
	case *tg.AuthSentCode:
		if strings.TrimSpace(result.PhoneCodeHash) == "" {
			return errors.New("Telegram 验证码响应缺少校验标识")
		}
		s.setStage(authorizationSessionAwaitingCode)
		s.notifyAsync(AuthorizationSessionUpdate{Status: AuthorizationStatusAwaitingCode})
		s.signalStart(nil)

		code, err := s.waitForSecret(ctx, s.codeInput)
		if err != nil {
			return err
		}
		authorization, err := raw.Auth().SignIn(ctx, s.phone, code, result.PhoneCodeHash)
		if errors.Is(err, gotdauth.ErrPasswordAuthNeeded) {
			s.setStage(authorizationSessionAwaitingPassword)
			s.notifyAsync(AuthorizationSessionUpdate{Status: AuthorizationStatusAwaitingPassword})
			password, waitErr := s.waitForSecret(ctx, s.passwordInput)
			if waitErr != nil {
				return waitErr
			}
			authorization, err = raw.Auth().Password(ctx, password)
		}
		if err != nil {
			return fmt.Errorf("验证 Telegram 登录信息: %w", err)
		}
		return s.completeAuthorization(ctx, authorization)
	case *tg.AuthSentCodeSuccess:
		authorization, ok := result.Authorization.(*tg.AuthAuthorization)
		if !ok {
			return fmt.Errorf("Telegram 验证码响应类型 %T 不支持", result.Authorization)
		}
		s.signalStart(nil)
		return s.completeAuthorization(ctx, authorization)
	default:
		return fmt.Errorf("Telegram 验证码响应类型 %T 不支持", sentCode)
	}
}

func (s *gotdAuthorizationSession) runQRAuthorization(ctx context.Context, raw *gotdtelegram.Client, loggedIn qrlogin.LoggedIn) error {
	s.setStage(authorizationSessionScanning)
	authorization, err := raw.QR().Auth(ctx, loggedIn, func(_ context.Context, token qrlogin.Token) error {
		dataURL, imageErr := renderTelegramQRDataURL(token)
		if imageErr != nil {
			return imageErr
		}
		s.notifyAsync(AuthorizationSessionUpdate{
			QRImageDataURL: dataURL,
			QRExpiresAt:    token.Expires(),
		})
		s.signalStart(nil)
		return nil
	})
	if err != nil {
		return fmt.Errorf("等待 Telegram 二维码授权: %w", err)
	}
	return s.completeAuthorization(ctx, authorization)
}

func (s *gotdAuthorizationSession) completeAuthorization(ctx context.Context, authorization *tg.AuthAuthorization) error {
	identity, err := telegramIdentityFromAuthorization(authorization)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.authorized = true
	s.stage = authorizationSessionAuthorized
	s.mu.Unlock()
	// This callback can promote and wait for this client to stop, so it must not
	// run on gotd's Run callback goroutine.
	s.notifyAsync(AuthorizationSessionUpdate{Identity: &identity})
	<-ctx.Done()
	return ctx.Err()
}

func (s *gotdAuthorizationSession) waitForSecret(ctx context.Context, input <-chan string) (string, error) {
	select {
	case value := <-input:
		return value, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func (s *gotdAuthorizationSession) beginSecretSubmission(expected, next authorizationSessionStage) bool {
	if s == nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.stage != expected {
		return false
	}
	s.stage = next
	return true
}

func (s *gotdAuthorizationSession) setStage(stage authorizationSessionStage) {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.stage = stage
	s.mu.Unlock()
}

func (s *gotdAuthorizationSession) signalStart(err error) {
	if s == nil {
		return
	}
	s.startOnce.Do(func() {
		s.startResult <- err
	})
}

func (s *gotdAuthorizationSession) notifyAsync(update AuthorizationSessionUpdate) {
	if s == nil {
		return
	}
	s.mu.Lock()
	notify := s.notify
	s.mu.Unlock()
	if notify != nil {
		go notify(update)
	}
}

func (s *gotdAuthorizationSession) stopAndWait(ctx context.Context) error {
	if s == nil {
		return nil
	}
	if ctx == nil {
		return errors.New("Telegram 授权 context 不能为空")
	}
	s.Close()
	s.mu.Lock()
	started := s.started
	done := s.runDone
	s.mu.Unlock()
	if !started || done == nil {
		return nil
	}
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("等待 Telegram 临时 session 停止: %w", ctx.Err())
	}
}

func temporaryAuthorizationSessionPath(sessionPath string, authorizationID uuid.UUID) string {
	return filepath.Join(filepath.Dir(sessionPath), "."+filepath.Base(sessionPath)+"."+authorizationID.String()+".pending")
}

func promoteAuthorizationSession(temporaryPath, persistentPath string) error {
	temporaryPath = filepath.Clean(strings.TrimSpace(temporaryPath))
	persistentPath = filepath.Clean(strings.TrimSpace(persistentPath))
	if temporaryPath == "." || persistentPath == "." || temporaryPath == persistentPath {
		return errors.New("Telegram 临时 session 路径无效")
	}
	if filepath.Dir(temporaryPath) != filepath.Dir(persistentPath) {
		return errors.New("Telegram 临时 session 必须与正式 session 位于同一目录")
	}
	info, err := os.Stat(temporaryPath)
	if err != nil {
		return fmt.Errorf("检查 Telegram 临时 session: %w", err)
	}
	if !info.Mode().IsRegular() {
		return errors.New("Telegram 临时 session 不是常规文件")
	}
	if err := os.Chmod(temporaryPath, 0o600); err != nil {
		return fmt.Errorf("设置 Telegram 临时 session 权限: %w", err)
	}
	if err := os.Rename(temporaryPath, persistentPath); err != nil {
		return fmt.Errorf("原子替换 Telegram session: %w", err)
	}
	return nil
}

func renderTelegramQRDataURL(token qrlogin.Token) (string, error) {
	image, err := token.Image(qr.M)
	if err != nil {
		return "", fmt.Errorf("生成 Telegram 二维码: %w", err)
	}
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, image); err != nil {
		return "", fmt.Errorf("编码 Telegram 二维码: %w", err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buffer.Bytes()), nil
}

func telegramIdentityFromAuthorization(authorization *tg.AuthAuthorization) (TelegramIdentity, error) {
	if authorization == nil || authorization.User == nil {
		return TelegramIdentity{}, errors.New("Telegram 授权响应缺少用户身份")
	}
	user, ok := authorization.User.AsNotEmpty()
	if !ok || user == nil || user.ID <= 0 {
		return TelegramIdentity{}, errors.New("Telegram 授权响应用户身份无效")
	}
	return TelegramIdentity{
		ID:        user.ID,
		Username:  strings.TrimSpace(user.Username),
		FirstName: strings.TrimSpace(user.FirstName),
		LastName:  strings.TrimSpace(user.LastName),
	}, nil
}
