package telegram

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	gotdtelegram "github.com/gotd/td/telegram"
	gotdauth "github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"go.uber.org/zap"
)

var _ Client = (*GotdClient)(nil)

const telegramChannelIDOffset int64 = 1_000_000_000_000

// GotdClientConfig contains the credentials and persistent session location
// needed to construct a gotd-backed client.
type GotdClientConfig struct {
	APIID       int
	APIHash     string
	Phone       string
	SessionPath string
	Logger      *zap.Logger
}

// LoginPrompter supplies interactive credentials when a session needs login.
type LoginPrompter interface {
	Code(ctx context.Context) (string, error)
	Password(ctx context.Context) (string, error)
}

// GotdClient adapts github.com/gotd/td to the ingestion Client boundary.
// Raw Telegram types are intentionally kept inside this package.
type GotdClient struct {
	raw        *gotdtelegram.Client
	api        *tg.Client
	phone      string
	resolving  sync.Mutex
	peersMu    sync.RWMutex
	peers      map[int64]tg.InputPeerClass
	chats      map[int64]Chat
	documents  map[documentKey]documentLocation
	subsMu     sync.RWMutex
	subs       map[uint64]*telegramSubscription
	nextSubID  uint64
	downloader *downloader.Downloader
}

type documentKey struct {
	chatID    int64
	messageID int64
}

type documentLocation struct {
	id            int64
	accessHash    int64
	fileReference []byte
}

type telegramSubscription struct {
	id    uint64
	fn    func(Message) error
	errCh chan error
}

type chatReferenceKind string

const (
	chatReferenceUsername chatReferenceKind = "username"
	chatReferenceInvite   chatReferenceKind = "invite"
	chatReferenceNumeric  chatReferenceKind = "numeric"
)

// NewGotdClient constructs a client with a file-backed session.
func NewGotdClient(config GotdClientConfig) (*GotdClient, error) {
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

	c := &GotdClient{
		phone:      strings.TrimSpace(config.Phone),
		peers:      make(map[int64]tg.InputPeerClass),
		chats:      make(map[int64]Chat),
		documents:  make(map[documentKey]documentLocation),
		subs:       make(map[uint64]*telegramSubscription),
		downloader: downloader.NewDownloader(),
	}
	c.raw = gotdtelegram.NewClient(config.APIID, strings.TrimSpace(config.APIHash), gotdtelegram.Options{
		Logger:         logger,
		SessionStorage: &gotdtelegram.FileSessionStorage{Path: sessionPath},
		UpdateHandler:  gotdtelegram.UpdateHandlerFunc(c.handleUpdates),
	})
	c.api = c.raw.API()
	return c, nil
}

// Run starts the MTProto session, authenticates if needed, and invokes fn
// while the connection is alive.
func (c *GotdClient) Run(ctx context.Context, prompter LoginPrompter, fn func(context.Context) error) error {
	if c == nil || c.raw == nil {
		return errors.New("Telegram client 未初始化")
	}
	if ctx == nil {
		return errors.New("Telegram client context 不能为空")
	}
	if fn == nil {
		return errors.New("Telegram client 回调不能为空")
	}

	authenticator := gotdAuthenticator{phone: c.phone, prompter: prompter}
	err := c.raw.Run(ctx, func(runCtx context.Context) error {
		flow := gotdauth.NewFlow(authenticator, gotdauth.SendCodeOptions{})
		if err := c.raw.Auth().IfNecessary(runCtx, flow); err != nil {
			return fmt.Errorf("Telegram 登录: %w", err)
		}
		if err := fn(runCtx); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return wrapTelegramError("运行 Telegram 客户端", err)
	}
	return nil
}

// ResolveChat resolves a username, Telegram link, invite link, or numeric
// chat identifier and caches the peer required by later operations.
func (c *GotdClient) ResolveChat(ctx context.Context, ref string) (Chat, error) {
	if err := requireTelegramContext(ctx, "解析 Telegram chat"); err != nil {
		return Chat{}, err
	}
	if err := c.ensureReady(); err != nil {
		return Chat{}, err
	}
	kind, value, err := parseChatReference(ref)
	if err != nil {
		return Chat{}, err
	}

	// A source can be resolved by multiple control tasks after a restart. The
	// lock prevents duplicate invite imports and duplicate peer refreshes.
	c.resolving.Lock()
	defer c.resolving.Unlock()

	switch kind {
	case chatReferenceUsername:
		return c.resolveUsername(ctx, value)
	case chatReferenceInvite:
		return c.resolveInvite(ctx, value)
	case chatReferenceNumeric:
		chatID, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return Chat{}, fmt.Errorf("解析 Telegram chat ID %q: %w", value, err)
		}
		return c.resolveNumeric(ctx, chatID)
	default:
		return Chat{}, fmt.Errorf("未知 Telegram chat 引用类型 %q", kind)
	}
}

// History reads one page of messages before offsetID and emits video
// documents through fn. A zero offset means the newest page.
func (c *GotdClient) History(ctx context.Context, chatID, offsetID int64, limit int, fn func(Message) error) error {
	if err := requireTelegramContext(ctx, "读取 Telegram 历史"); err != nil {
		return err
	}
	if err := c.ensureReady(); err != nil {
		return err
	}
	if chatID == 0 {
		return errors.New("Telegram chat ID 不能为空")
	}
	if offsetID < 0 {
		return errors.New("Telegram 历史 offset ID 不能为负数")
	}
	if limit <= 0 {
		return errors.New("Telegram 历史分页大小必须为正数")
	}
	if fn == nil {
		return errors.New("Telegram 历史回调不能为空")
	}

	peer, err := c.peerForChat(ctx, chatID)
	if err != nil {
		return err
	}
	result, err := c.api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
		Peer:     peer,
		OffsetID: int(offsetID),
		Limit:    limit,
	})
	if err != nil {
		return wrapTelegramError("读取 Telegram 历史", err)
	}
	modified, ok := result.AsModified()
	if !ok {
		return nil
	}
	for _, rawMessage := range modified.GetMessages() {
		message, ok, err := c.normalizeMessage(rawMessage, chatID)
		if err != nil {
			return fmt.Errorf("解析 Telegram 历史消息: %w", err)
		}
		if !ok {
			continue
		}
		if err := fn(message); err != nil {
			return fmt.Errorf("处理 Telegram 历史消息 %d: %w", message.MessageID, err)
		}
	}
	return nil
}

// Subscribe registers an update callback and blocks until the context ends
// or the callback returns an error.
func (c *GotdClient) Subscribe(ctx context.Context, fn func(Message) error) error {
	if err := c.ensureReady(); err != nil {
		return err
	}
	if ctx == nil {
		return errors.New("Telegram 订阅 context 不能为空")
	}
	if fn == nil {
		return errors.New("Telegram 订阅回调不能为空")
	}

	c.subsMu.Lock()
	c.nextSubID++
	sub := &telegramSubscription{id: c.nextSubID, fn: fn, errCh: make(chan error, 1)}
	c.subs[sub.id] = sub
	c.subsMu.Unlock()
	defer c.removeSubscription(sub.id)

	select {
	case err := <-sub.errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// RefreshMessage fetches a message again so its document file reference is
// current before a download.
func (c *GotdClient) RefreshMessage(ctx context.Context, chatID, messageID int64) (Message, error) {
	if err := requireTelegramContext(ctx, "刷新 Telegram 消息"); err != nil {
		return Message{}, err
	}
	if err := c.ensureReady(); err != nil {
		return Message{}, err
	}
	if chatID == 0 || messageID <= 0 {
		return Message{}, errors.New("Telegram 消息标识必须为正数且 chat ID 不能为空")
	}
	peer, err := c.peerForChat(ctx, chatID)
	if err != nil {
		return Message{}, err
	}
	inputMessage := []tg.InputMessageClass{&tg.InputMessageID{ID: int(messageID)}}
	var result tg.MessagesMessagesClass
	if channel, ok := inputChannelFromPeer(peer); ok {
		result, err = c.api.ChannelsGetMessages(ctx, &tg.ChannelsGetMessagesRequest{
			Channel: channel,
			ID:      inputMessage,
		})
	} else {
		result, err = c.api.MessagesGetMessages(ctx, inputMessage)
	}
	if err != nil {
		return Message{}, wrapTelegramError("刷新 Telegram 消息", err)
	}
	modified, ok := result.AsModified()
	if !ok {
		return Message{}, fmt.Errorf("Telegram 消息 %d 不存在", messageID)
	}
	for _, rawMessage := range modified.GetMessages() {
		message, ok, err := c.normalizeMessage(rawMessage, chatID)
		if err != nil {
			return Message{}, fmt.Errorf("解析 Telegram 刷新消息: %w", err)
		}
		if ok && message.MessageID == messageID {
			return message, nil
		}
	}
	return Message{}, fmt.Errorf("Telegram 消息 %d 不是可导入视频", messageID)
}

// Download streams the latest document reference into dst.
func (c *GotdClient) Download(ctx context.Context, message Message, dst io.Writer) error {
	if err := requireTelegramContext(ctx, "下载 Telegram 视频"); err != nil {
		return err
	}
	if err := c.ensureReady(); err != nil {
		return err
	}
	if dst == nil {
		return errors.New("Telegram 下载目标不能为空")
	}
	key := documentKey{chatID: message.ChatID, messageID: message.MessageID}
	c.peersMu.RLock()
	document, ok := c.documents[key]
	c.peersMu.RUnlock()
	if !ok {
		return fmt.Errorf("Telegram 消息 %d 缺少最新 document file reference", message.MessageID)
	}

	_, err := c.downloader.Download(c.api, &tg.InputDocumentFileLocation{
		ID:            document.id,
		AccessHash:    document.accessHash,
		FileReference: append([]byte(nil), document.fileReference...),
	}).Stream(ctx, dst)
	if err != nil {
		return wrapTelegramError("下载 Telegram 视频", err)
	}
	return nil
}

func (c *GotdClient) ensureReady() error {
	if c == nil || c.raw == nil || c.api == nil {
		return errors.New("Telegram client 未初始化")
	}
	return nil
}

func requireTelegramContext(ctx context.Context, operation string) error {
	if ctx == nil {
		return fmt.Errorf("%s context 不能为空", operation)
	}
	return nil
}

func (c *GotdClient) peerForChat(ctx context.Context, chatID int64) (tg.InputPeerClass, error) {
	c.peersMu.RLock()
	peer, ok := c.peers[chatID]
	c.peersMu.RUnlock()
	if ok {
		return peer, nil
	}
	if _, err := c.ResolveChat(ctx, strconv.FormatInt(chatID, 10)); err != nil {
		return nil, fmt.Errorf("获取 Telegram chat %d peer: %w", chatID, err)
	}
	c.peersMu.RLock()
	peer, ok = c.peers[chatID]
	c.peersMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("Telegram chat %d peer 未缓存", chatID)
	}
	return peer, nil
}

func (c *GotdClient) resolveUsername(ctx context.Context, username string) (Chat, error) {
	result, err := c.api.ContactsResolveUsername(ctx, strings.TrimPrefix(username, "@"))
	if err != nil {
		return Chat{}, wrapTelegramError("解析 Telegram 用户名", err)
	}
	chat, peer, err := c.chatFromResolved(result)
	if err != nil {
		return Chat{}, err
	}
	c.rememberChat(chat, peer)
	return chat, nil
}

func (c *GotdClient) resolveInvite(ctx context.Context, hash string) (Chat, error) {
	invite, err := c.api.MessagesCheckChatInvite(ctx, hash)
	if err != nil {
		return Chat{}, wrapTelegramError("检查 Telegram 邀请链接", err)
	}
	if already, ok := invite.(*tg.ChatInviteAlready); ok {
		chat, peer, err := c.chatFromClass(already.GetChat())
		if err != nil {
			return Chat{}, err
		}
		c.rememberChat(chat, peer)
		return chat, nil
	}

	updates, err := c.api.MessagesImportChatInvite(ctx, hash)
	if err != nil {
		return Chat{}, wrapTelegramError("加入 Telegram 邀请群组", err)
	}
	withChats, ok := updates.(interface{ GetChats() []tg.ChatClass })
	if !ok {
		return Chat{}, fmt.Errorf("Telegram 邀请响应类型 %T 不包含群组", updates)
	}
	for _, rawChat := range withChats.GetChats() {
		chat, peer, err := c.chatFromClass(rawChat)
		if err == nil {
			c.rememberChat(chat, peer)
			return chat, nil
		}
	}
	return Chat{}, errors.New("Telegram 邀请响应未返回可访问群组")
}

func (c *GotdClient) resolveNumeric(ctx context.Context, chatID int64) (Chat, error) {
	if chatID <= -telegramChannelIDOffset {
		result, err := c.api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{Limit: 1000})
		if err != nil {
			return Chat{}, wrapTelegramError("按 ID 查找 Telegram 频道", err)
		}
		modified, ok := result.AsModified()
		if !ok {
			return Chat{}, fmt.Errorf("Telegram chat %d 不在当前账号对话列表", chatID)
		}
		for _, rawChat := range modified.GetChats() {
			chat, peer, err := c.chatFromClass(rawChat)
			if err == nil && chat.ID == chatID {
				c.rememberChat(chat, peer)
				return chat, nil
			}
		}
		return Chat{}, fmt.Errorf("Telegram 频道 %d 不在当前账号对话列表", chatID)
	}
	if chatID <= 0 {
		return Chat{}, errors.New("Telegram 基础群组 ID 必须为正数；频道请使用 -100 开头的 ID")
	}
	result, err := c.api.MessagesGetChats(ctx, []int64{chatID})
	if err != nil {
		return Chat{}, wrapTelegramError("按 ID 查找 Telegram 群组", err)
	}
	for _, rawChat := range result.GetChats() {
		chat, peer, err := c.chatFromClass(rawChat)
		if err == nil && chat.ID == chatID {
			c.rememberChat(chat, peer)
			return chat, nil
		}
	}
	return Chat{}, fmt.Errorf("Telegram 群组 %d 不存在或当前账号无权访问", chatID)
}

func (c *GotdClient) chatFromResolved(result *tg.ContactsResolvedPeer) (Chat, tg.InputPeerClass, error) {
	if result == nil || result.Peer == nil {
		return Chat{}, nil, errors.New("Telegram 用户名解析未返回 peer")
	}
	targetID, ok := chatIDFromPeer(result.Peer)
	if !ok {
		return Chat{}, nil, errors.New("Telegram 引用不是群组或频道")
	}
	for _, rawChat := range result.Chats {
		chat, peer, err := c.chatFromClass(rawChat)
		if err == nil && chat.ID == targetID {
			return chat, peer, nil
		}
	}
	return Chat{}, nil, fmt.Errorf("Telegram peer %d 缺少群组详情", targetID)
}

func (c *GotdClient) chatFromClass(rawChat tg.ChatClass) (Chat, tg.InputPeerClass, error) {
	switch chat := rawChat.(type) {
	case *tg.Chat:
		return Chat{ID: chat.ID, Title: chat.Title}, chat.AsInputPeer(), nil
	case *tg.Channel:
		return Chat{
			ID:       channelChatID(chat.ID),
			Title:    chat.Title,
			Username: strings.TrimSpace(chat.Username),
		}, chat.AsInputPeer(), nil
	default:
		return Chat{}, nil, fmt.Errorf("Telegram chat 类型 %T 不可访问", rawChat)
	}
}

func (c *GotdClient) rememberChat(chat Chat, peer tg.InputPeerClass) {
	c.peersMu.Lock()
	c.peers[chat.ID] = peer
	c.chats[chat.ID] = chat
	c.peersMu.Unlock()
}

func (c *GotdClient) removeSubscription(id uint64) {
	c.subsMu.Lock()
	delete(c.subs, id)
	c.subsMu.Unlock()
}

func (c *GotdClient) handleUpdates(ctx context.Context, updates tg.UpdatesClass) error {
	for _, update := range flattenUpdates(updates) {
		var rawMessage tg.MessageClass
		switch value := update.(type) {
		case *tg.UpdateNewMessage:
			rawMessage = value.GetMessage()
		case *tg.UpdateNewChannelMessage:
			rawMessage = value.GetMessage()
		default:
			continue
		}
		message, ok, err := c.normalizeMessage(rawMessage, 0)
		if err != nil {
			return fmt.Errorf("解析 Telegram 实时消息: %w", err)
		}
		if !ok {
			continue
		}
		if err := c.notifySubscribers(ctx, message); err != nil {
			return err
		}
	}
	return nil
}

func flattenUpdates(updates tg.UpdatesClass) []tg.UpdateClass {
	switch value := updates.(type) {
	case *tg.UpdateShort:
		return []tg.UpdateClass{value.GetUpdate()}
	case *tg.Updates:
		return value.GetUpdates()
	case *tg.UpdatesCombined:
		return value.GetUpdates()
	default:
		return nil
	}
}

func (c *GotdClient) notifySubscribers(_ context.Context, message Message) error {
	c.subsMu.RLock()
	subs := make([]*telegramSubscription, 0, len(c.subs))
	for _, sub := range c.subs {
		subs = append(subs, sub)
	}
	c.subsMu.RUnlock()

	for _, sub := range subs {
		if err := sub.fn(message); err != nil {
			wrapped := fmt.Errorf("Telegram 订阅回调: %w", err)
			select {
			case sub.errCh <- wrapped:
			default:
			}
			return wrapped
		}
	}
	return nil
}

func (c *GotdClient) normalizeMessage(raw tg.MessageClass, fallbackChatID int64) (Message, bool, error) {
	message, ok := raw.(*tg.Message)
	if !ok || message == nil {
		return Message{}, false, nil
	}
	media, ok := message.GetMedia()
	if !ok {
		return Message{}, false, nil
	}
	documentMedia, ok := media.(*tg.MessageMediaDocument)
	if !ok {
		return Message{}, false, nil
	}
	documentClass, ok := documentMedia.GetDocument()
	if !ok {
		return Message{}, false, nil
	}
	document, ok := documentClass.AsNotEmpty()
	if !ok || document == nil {
		return Message{}, false, nil
	}
	if !isVideoDocument(documentMedia, document) {
		return Message{}, false, nil
	}

	chatID := fallbackChatID
	if chatID == 0 {
		chatID, ok = chatIDFromPeer(message.GetPeerID())
		if !ok {
			return Message{}, false, nil
		}
	}
	messageID := int64(message.GetID())
	if messageID <= 0 {
		return Message{}, false, nil
	}
	filename := documentFilename(document)
	if filename == "" {
		filename = fmt.Sprintf("telegram-%d", document.ID)
	}
	sentAt := time.Unix(int64(message.GetDate()), 0).UTC()
	normalized := Message{
		ChatID:       chatID,
		MessageID:    messageID,
		DocumentID:   document.ID,
		DocumentDCID: document.DCID,
		Filename:     filename,
		MIMEType:     document.MimeType,
		Size:         document.Size,
		Caption:      message.GetMessage(),
		MessageURL:   c.messageURL(chatID, messageID),
		SentAt:       sentAt,
	}
	c.peersMu.Lock()
	c.documents[documentKey{chatID: chatID, messageID: messageID}] = documentLocation{
		id:            document.ID,
		accessHash:    document.AccessHash,
		fileReference: append([]byte(nil), document.FileReference...),
	}
	c.peersMu.Unlock()
	return normalized, true, nil
}

func isVideoDocument(media *tg.MessageMediaDocument, document *tg.Document) bool {
	if media.Video || strings.HasPrefix(strings.ToLower(strings.TrimSpace(document.MimeType)), "video/") {
		return true
	}
	for _, attribute := range document.Attributes {
		if _, ok := attribute.(*tg.DocumentAttributeVideo); ok {
			return true
		}
	}
	return false
}

func documentFilename(document *tg.Document) string {
	for _, attribute := range document.Attributes {
		if filename, ok := attribute.(*tg.DocumentAttributeFilename); ok && strings.TrimSpace(filename.FileName) != "" {
			return filename.FileName
		}
	}
	return ""
}

func (c *GotdClient) messageURL(chatID, messageID int64) string {
	c.peersMu.RLock()
	chat := c.chats[chatID]
	c.peersMu.RUnlock()
	if chat.Username != "" {
		return fmt.Sprintf("https://t.me/%s/%d", strings.TrimPrefix(chat.Username, "@"), messageID)
	}
	if chatID <= -telegramChannelIDOffset {
		channelID := channelIDFromChatID(chatID)
		if channelID > 0 {
			return fmt.Sprintf("https://t.me/c/%d/%d", channelID, messageID)
		}
	}
	return ""
}

func inputChannelFromPeer(peer tg.InputPeerClass) (tg.InputChannelClass, bool) {
	switch value := peer.(type) {
	case *tg.InputPeerChannel:
		return &tg.InputChannel{ChannelID: value.ChannelID, AccessHash: value.AccessHash}, true
	case *tg.InputPeerChannelFromMessage:
		return &tg.InputChannelFromMessage{
			Peer:      value.Peer,
			MsgID:     value.MsgID,
			ChannelID: value.ChannelID,
		}, true
	default:
		return nil, false
	}
}

func parseChatReference(ref string) (chatReferenceKind, string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", "", errors.New("Telegram chat 引用不能为空")
	}
	if _, err := strconv.ParseInt(ref, 10, 64); err == nil {
		return chatReferenceNumeric, ref, nil
	}

	if strings.HasPrefix(ref, "tg:") {
		ref = "tg://" + strings.TrimPrefix(ref, "tg://")
	}
	if strings.HasPrefix(ref, "t.me/") || strings.HasPrefix(ref, "telegram.me/") || strings.HasPrefix(ref, "telegram.dog/") {
		ref = "https://" + ref
	}
	if strings.HasPrefix(ref, "https://") || strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "tg://") {
		parsed, err := url.Parse(ref)
		if err != nil {
			return "", "", fmt.Errorf("解析 Telegram chat 链接: %w", err)
		}
		if parsed.Scheme == "tg" {
			switch parsed.Hostname() {
			case "resolve":
				return parseUsernameValue(parsed.Query().Get("domain"))
			case "join":
				return parseInviteValue(parsed.Query().Get("invite"))
			default:
				return "", "", fmt.Errorf("不支持 Telegram 链接类型 %q", parsed.Hostname())
			}
		}
		if parsed.Hostname() != "t.me" && parsed.Hostname() != "telegram.me" && parsed.Hostname() != "telegram.dog" {
			return "", "", fmt.Errorf("不支持 Telegram 链接域名 %q", parsed.Hostname())
		}
		parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
		if len(parts) == 0 || parts[0] == "" {
			return "", "", errors.New("Telegram chat 链接缺少目标")
		}
		switch parts[0] {
		case "joinchat":
			if len(parts) < 2 {
				return "", "", errors.New("Telegram 邀请链接缺少 hash")
			}
			return parseInviteValue(parts[1])
		case "c":
			if len(parts) < 2 {
				return "", "", errors.New("Telegram 内部频道链接缺少 ID")
			}
			channelID, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil || channelID <= 0 {
				return "", "", errors.New("Telegram 内部频道 ID 无效")
			}
			return chatReferenceNumeric, strconv.FormatInt(-telegramChannelIDOffset-channelID, 10), nil
		default:
			if strings.HasPrefix(parts[0], "+") {
				return parseInviteValue(parts[0])
			}
			return parseUsernameValue(parts[0])
		}
	}
	return parseUsernameValue(ref)
}

func parseUsernameValue(value string) (chatReferenceKind, string, error) {
	value = strings.TrimSpace(strings.TrimPrefix(value, "@"))
	if value == "" {
		return "", "", errors.New("Telegram 用户名不能为空")
	}
	return chatReferenceUsername, value, nil
}

func parseInviteValue(value string) (chatReferenceKind, string, error) {
	value = strings.TrimSpace(strings.TrimPrefix(value, "+"))
	if value == "" {
		return "", "", errors.New("Telegram 邀请 hash 不能为空")
	}
	return chatReferenceInvite, value, nil
}

func chatIDFromPeer(peer tg.PeerClass) (int64, bool) {
	switch value := peer.(type) {
	case *tg.PeerChat:
		return value.ChatID, true
	case *tg.PeerChannel:
		return channelChatID(value.ChannelID), true
	default:
		return 0, false
	}
}

func channelChatID(channelID int64) int64 {
	if channelID <= 0 {
		return 0
	}
	return -telegramChannelIDOffset - channelID
}

func channelIDFromChatID(chatID int64) int64 {
	if chatID > -telegramChannelIDOffset {
		return 0
	}
	return -chatID - telegramChannelIDOffset
}

func wrapTelegramError(operation string, err error) error {
	if err == nil {
		return nil
	}
	if wait, ok := tgerr.AsFloodWait(err); ok {
		seconds := int(wait / time.Second)
		if wait%time.Second != 0 {
			seconds++
		}
		if seconds < 1 {
			seconds = 1
		}
		err = &FloodWaitError{Seconds: seconds, Err: err}
	}
	return fmt.Errorf("%s: %w", operation, err)
}

type gotdAuthenticator struct {
	phone    string
	prompter LoginPrompter
}

func (a gotdAuthenticator) Phone(context.Context) (string, error) {
	return a.phone, nil
}

func (a gotdAuthenticator) Code(ctx context.Context, _ *tg.AuthSentCode) (string, error) {
	if a.prompter == nil {
		return "", errors.New("Telegram 未配置验证码输入器")
	}
	return a.prompter.Code(ctx)
}

func (a gotdAuthenticator) Password(ctx context.Context) (string, error) {
	if a.prompter == nil {
		return "", errors.New("Telegram 未配置二次验证密码输入器")
	}
	return a.prompter.Password(ctx)
}

func (gotdAuthenticator) AcceptTermsOfService(_ context.Context, tos tg.HelpTermsOfService) error {
	return &gotdauth.SignUpRequired{TermsOfService: tos}
}

func (gotdAuthenticator) SignUp(context.Context) (gotdauth.UserInfo, error) {
	return gotdauth.UserInfo{}, errors.New("Telegram 个人账号采集器不支持注册新账号")
}
