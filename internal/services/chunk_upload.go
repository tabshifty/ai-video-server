package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ChunkUploadService struct {
	baseDir      string
	assembledDir string
}

type ChunkSession struct {
	ID                string         `json:"id"`
	UserID            string         `json:"user_id"`
	Filename          string         `json:"filename"`
	FileSize          int64          `json:"file_size"`
	ChunkSize         int64          `json:"chunk_size"`
	TotalChunks       int            `json:"total_chunks"`
	Hash              string         `json:"hash"`
	Type              string         `json:"type"`
	Title             string         `json:"title"`
	Description       string         `json:"description"`
	Tags              []string       `json:"tags"`
	ActorIDs          []string       `json:"actor_ids"`
	ActorNames        []string       `json:"actor_names"`
	ImageCollectionID string         `json:"image_collection_id"`
	CollectionIDs     []string       `json:"collection_ids"`
	SiteCategory      string         `json:"site_category"`
	Uploaded          map[int]bool   `json:"uploaded"`
	CreatedAt         time.Time      `json:"created_at"`
	LastUpdated       time.Time      `json:"last_updated"`
	Extra             map[string]any `json:"extra,omitempty"`
}

func NewChunkUploadService(uploadTempDir string) *ChunkUploadService {
	return &ChunkUploadService{
		baseDir:      filepath.Join(uploadTempDir, "chunk-sessions"),
		assembledDir: filepath.Join(uploadTempDir, "assembled"),
	}
}

func (s *ChunkUploadService) Init(ctx context.Context, userID uuid.UUID, filename string, fileSize, chunkSize int64, totalChunks int, hash, typ, title, description string, tags []string, actorIDs, actorNames []string, imageCollectionID, siteCategory string, collectionIDs []string) (ChunkSession, error) {
	_ = ctx
	if chunkSize <= 0 || totalChunks <= 0 {
		return ChunkSession{}, fmt.Errorf("invalid chunk size or total chunks")
	}
	session := ChunkSession{
		ID:                uuid.NewString(),
		UserID:            userID.String(),
		Filename:          filename,
		FileSize:          fileSize,
		ChunkSize:         chunkSize,
		TotalChunks:       totalChunks,
		Hash:              strings.TrimSpace(hash),
		Type:              strings.TrimSpace(typ),
		Title:             title,
		Description:       description,
		Tags:              tags,
		ActorIDs:          actorIDs,
		ActorNames:        actorNames,
		ImageCollectionID: strings.TrimSpace(imageCollectionID),
		CollectionIDs:     collectionIDs,
		SiteCategory:      strings.TrimSpace(siteCategory),
		Uploaded:          map[int]bool{},
		CreatedAt:         time.Now().UTC(),
		LastUpdated:       time.Now().UTC(),
	}
	if err := os.MkdirAll(filepath.Join(s.sessionDir(session.ID), "chunks"), 0o755); err != nil {
		return ChunkSession{}, fmt.Errorf("create session dir: %w", err)
	}
	if err := s.persist(session); err != nil {
		return ChunkSession{}, err
	}
	return session, nil
}

func (s *ChunkUploadService) SaveChunk(ctx context.Context, sessionID string, chunkIndex int, chunk io.Reader) (ChunkSession, error) {
	_ = ctx
	session, err := s.Load(sessionID)
	if err != nil {
		return ChunkSession{}, err
	}
	if chunkIndex < 0 || chunkIndex >= session.TotalChunks {
		return ChunkSession{}, fmt.Errorf("chunk index out of range")
	}
	dir := s.sessionDir(sessionID)
	// 目录仅由 Init 创建；取消中的分片请求不得重新创建会话。
	// 旧版 Init 未创建 chunks 子目录，仅允许在现存会话内补建这一层。
	if err := os.Mkdir(filepath.Join(dir, "chunks"), 0o755); err != nil && !os.IsExist(err) {
		return ChunkSession{}, fmt.Errorf("create chunk subdir: %w", err)
	}
	target := filepath.Join(dir, "chunks", strconv.Itoa(chunkIndex)+".part")
	f, err := os.Create(target)
	if err != nil {
		return ChunkSession{}, fmt.Errorf("create chunk file: %w", err)
	}
	if _, err := io.Copy(f, chunk); err != nil {
		f.Close()
		return ChunkSession{}, fmt.Errorf("write chunk: %w", err)
	}
	if err := f.Close(); err != nil {
		return ChunkSession{}, fmt.Errorf("close chunk file: %w", err)
	}
	session.Uploaded[chunkIndex] = true
	session.LastUpdated = time.Now().UTC()
	if err := s.persist(session); err != nil {
		return ChunkSession{}, err
	}
	return session, nil
}

func (s *ChunkUploadService) Complete(ctx context.Context, sessionID string) (ChunkSession, string, error) {
	_ = ctx
	session, err := s.Load(sessionID)
	if err != nil {
		return ChunkSession{}, "", err
	}
	if len(session.Uploaded) != session.TotalChunks {
		return ChunkSession{}, "", fmt.Errorf("not all chunks uploaded")
	}
	dir := s.sessionDir(sessionID)
	if err := os.MkdirAll(s.assembledDir, 0o755); err != nil {
		return ChunkSession{}, "", fmt.Errorf("create assembled dir: %w", err)
	}
	finalPath := filepath.Join(s.assembledDir, session.ID+"-"+session.Filename)
	dst, err := os.Create(finalPath)
	if err != nil {
		return ChunkSession{}, "", fmt.Errorf("create assembled file: %w", err)
	}
	keys := make([]int, 0, len(session.Uploaded))
	for idx := range session.Uploaded {
		keys = append(keys, idx)
	}
	sort.Ints(keys)
	for _, idx := range keys {
		part := filepath.Join(dir, "chunks", strconv.Itoa(idx)+".part")
		src, err := os.Open(part)
		if err != nil {
			dst.Close()
			return ChunkSession{}, "", fmt.Errorf("open chunk: %w", err)
		}
		if _, err := io.Copy(dst, src); err != nil {
			src.Close()
			dst.Close()
			return ChunkSession{}, "", fmt.Errorf("merge chunk: %w", err)
		}
		_ = src.Close()
	}
	if err := dst.Close(); err != nil {
		return ChunkSession{}, "", fmt.Errorf("close assembled file: %w", err)
	}
	return session, finalPath, nil
}

func (s *ChunkUploadService) Abort(sessionID string) error {
	return os.RemoveAll(s.sessionDir(sessionID))
}

func (s *ChunkUploadService) Load(sessionID string) (ChunkSession, error) {
	raw, err := os.ReadFile(s.metaPath(sessionID))
	if err != nil {
		return ChunkSession{}, fmt.Errorf("read session meta: %w", err)
	}
	var session ChunkSession
	if err := json.Unmarshal(raw, &session); err != nil {
		return ChunkSession{}, fmt.Errorf("decode session meta: %w", err)
	}
	if session.Uploaded == nil {
		session.Uploaded = map[int]bool{}
	}
	return session, nil
}

// CollectionReferences includes every retained upload session, including old ones.
// Callers serialize new session creation with cleanup before reading this snapshot.
func (s *ChunkUploadService) CollectionReferences(ctx context.Context, kind string) (map[uuid.UUID]bool, error) {
	refs := map[uuid.UUID]bool{}
	entries, err := os.ReadDir(s.baseDir)
	if os.IsNotExist(err) {
		return refs, nil
	}
	if err != nil {
		return nil, fmt.Errorf("读取上传会话目录: %w", err)
	}
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if !entry.IsDir() {
			continue
		}
		if _, err := uuid.Parse(entry.Name()); err != nil {
			continue
		}
		raw, err := os.ReadFile(s.metaPath(entry.Name()))
		if os.IsNotExist(err) {
			continue
		} // 允许同时取消会话。
		if err != nil {
			return nil, fmt.Errorf("读取上传会话引用: %w", err)
		}
		var session ChunkSession
		if err := json.Unmarshal(raw, &session); err != nil {
			return nil, fmt.Errorf("上传会话元数据损坏，无法安全清理合集: %w", err)
		}
		ids := session.CollectionIDs
		if kind == "image" {
			ids = []string{session.ImageCollectionID}
		}
		for _, rawID := range ids {
			if strings.TrimSpace(rawID) == "" {
				continue
			}
			id, err := uuid.Parse(strings.TrimSpace(rawID))
			if err != nil {
				return nil, fmt.Errorf("上传会话合集引用无效: %w", err)
			}
			refs[id] = true
		}
	}
	return refs, nil
}

func (s *ChunkUploadService) sessionDir(sessionID string) string {
	return filepath.Join(s.baseDir, sessionID)
}

func (s *ChunkUploadService) metaPath(sessionID string) string {
	return filepath.Join(s.sessionDir(sessionID), "meta.json")
}

func (s *ChunkUploadService) persist(session ChunkSession) error {
	dir := s.sessionDir(session.ID)
	raw, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("encode session meta: %w", err)
	}
	// 原子替换，避免引用扫描撞上 SaveChunk 截断写入中的元数据。
	f, err := os.CreateTemp(dir, ".meta-*")
	if err != nil {
		return fmt.Errorf("create session meta: %w", err)
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(raw); err != nil {
		_ = f.Close()
		return fmt.Errorf("write session meta: %w", err)
	}
	if err = f.Close(); err != nil {
		return fmt.Errorf("close session meta: %w", err)
	}
	if err = os.Rename(f.Name(), s.metaPath(session.ID)); err != nil {
		return fmt.Errorf("replace session meta: %w", err)
	}
	return nil
}
