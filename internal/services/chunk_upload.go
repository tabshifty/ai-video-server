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
	ID          string         `json:"id"`
	UserID      string         `json:"user_id"`
	Filename    string         `json:"filename"`
	FileSize    int64          `json:"file_size"`
	ChunkSize   int64          `json:"chunk_size"`
	TotalChunks int            `json:"total_chunks"`
	Hash        string         `json:"hash"`
	Type        string         `json:"type"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Tags        []string       `json:"tags"`
	ActorIDs    []string       `json:"actor_ids"`
	ActorNames  []string       `json:"actor_names"`
	Uploaded    map[int]bool   `json:"uploaded"`
	CreatedAt   time.Time      `json:"created_at"`
	LastUpdated time.Time      `json:"last_updated"`
	Extra       map[string]any `json:"extra,omitempty"`
}

func NewChunkUploadService(uploadTempDir string) *ChunkUploadService {
	return &ChunkUploadService{
		baseDir:      filepath.Join(uploadTempDir, "chunk-sessions"),
		assembledDir: filepath.Join(uploadTempDir, "assembled"),
	}
}

func (s *ChunkUploadService) Init(ctx context.Context, userID uuid.UUID, filename string, fileSize, chunkSize int64, totalChunks int, hash, typ, title, description string, tags []string, actorIDs, actorNames []string) (ChunkSession, error) {
	_ = ctx
	if chunkSize <= 0 || totalChunks <= 0 {
		return ChunkSession{}, fmt.Errorf("invalid chunk size or total chunks")
	}
	session := ChunkSession{
		ID:          uuid.NewString(),
		UserID:      userID.String(),
		Filename:    filename,
		FileSize:    fileSize,
		ChunkSize:   chunkSize,
		TotalChunks: totalChunks,
		Hash:        strings.TrimSpace(hash),
		Type:        strings.TrimSpace(typ),
		Title:       title,
		Description: description,
		Tags:        tags,
		ActorIDs:    actorIDs,
		ActorNames:  actorNames,
		Uploaded:    map[int]bool{},
		CreatedAt:   time.Now().UTC(),
		LastUpdated: time.Now().UTC(),
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
	if err := os.MkdirAll(filepath.Join(dir, "chunks"), 0o755); err != nil {
		return ChunkSession{}, fmt.Errorf("create chunk dir: %w", err)
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

func (s *ChunkUploadService) sessionDir(sessionID string) string {
	return filepath.Join(s.baseDir, sessionID)
}

func (s *ChunkUploadService) metaPath(sessionID string) string {
	return filepath.Join(s.sessionDir(sessionID), "meta.json")
}

func (s *ChunkUploadService) persist(session ChunkSession) error {
	dir := s.sessionDir(session.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create session dir: %w", err)
	}
	raw, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("encode session meta: %w", err)
	}
	if err := os.WriteFile(s.metaPath(session.ID), raw, 0o644); err != nil {
		return fmt.Errorf("write session meta: %w", err)
	}
	return nil
}
