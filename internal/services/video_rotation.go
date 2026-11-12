package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sys/unix"
	"video-server/internal/models"
	"video-server/internal/repository"
	"video-server/pkg/ffmpeg"
)

// VideoRotationService performs durable, independently publishable rotations.
type VideoRotationService struct {
	repo *repository.VideoRepository
	root string
}

// NewVideoRotationService uses the same managed storage root as video playback.
func NewVideoRotationService(repo *repository.VideoRepository, root string) *VideoRotationService {
	return &VideoRotationService{repo: repo, root: root}
}

// RotationPlan binds confirmation to the current physical input version.
type RotationPlan struct {
	Version           string `json:"version"`
	Mode              string `json:"mode"`
	LosslessAvailable bool   `json:"lossless_available"`
	Reason            string `json:"reason"`
	TemporaryBytes    int64  `json:"temporary_bytes"`
}

// Preview probes without changing files or queuing work.
func (s *VideoRotationService) Preview(ctx context.Context, id uuid.UUID, mode string) (RotationPlan, error) {
	v, err := s.repo.GetVideoByID(ctx, id)
	if err != nil {
		return RotationPlan{}, err
	}
	p, info, err := s.probe(ctx, v)
	if err != nil {
		return RotationPlan{}, err
	}
	selected := "lossless"
	reason := "修改旋转标记，不重新编码；播放效果依赖播放器识别方向信息。"
	if !p.Lossless || mode == "reencode" {
		selected = "reencode"
		reason = "需要重新编码画面，处理较慢且可能损失画质。"
	}
	if mode != "" && mode != "auto" && mode != "lossless" && mode != "reencode" {
		return RotationPlan{}, fmt.Errorf("处理方式无效")
	}
	if mode == "lossless" && !p.Lossless {
		return RotationPlan{}, fmt.Errorf("该视频无法使用无损旋转，请选择重新编码并确认")
	}
	estimate := rotationSpaceEstimate(p, info.Size(), selected)
	return RotationPlan{Version: rotationVersion(v, info), Mode: selected, LosslessAvailable: p.Lossless, Reason: reason, TemporaryBytes: estimate}, nil
}

func (s *VideoRotationService) probe(ctx context.Context, v models.Video) (ffmpeg.RotationProbe, os.FileInfo, error) {
	if v.Status != "ready" {
		return ffmpeg.RotationProbe{}, nil, fmt.Errorf("仅已就绪视频可修正方向")
	}
	if SourcePlaybackPathFromMetadata(v.Metadata) != "" {
		return ffmpeg.RotationProbe{}, nil, fmt.Errorf("杜比视界源暂不支持方向修正")
	}
	if err := s.ownedPath(v.ID, v.TranscodedPath); err != nil {
		return ffmpeg.RotationProbe{}, nil, err
	}
	info, err := os.Stat(v.TranscodedPath)
	if err != nil || !info.Mode().IsRegular() || info.Size() <= 0 {
		return ffmpeg.RotationProbe{}, nil, fmt.Errorf("当前播放文件不可用")
	}
	compat, err := ffmpeg.ProbePlaybackCompatibility(ctx, v.TranscodedPath)
	if err != nil || !compat.VideoStreamFound {
		return ffmpeg.RotationProbe{}, nil, fmt.Errorf("无法确认视频兼容性，暂不能修正方向")
	}
	if compat.DolbyVision {
		return ffmpeg.RotationProbe{}, nil, fmt.Errorf("杜比视界源暂不支持方向修正")
	}
	if compat.HDR {
		return ffmpeg.RotationProbe{}, nil, fmt.Errorf("HDR 视频暂不支持方向修正，以避免画质损失")
	}
	p, err := ffmpeg.ProbeRotation(ctx, v.TranscodedPath)
	return p, info, err
}

func rotationVersion(v models.Video, info os.FileInfo) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%s\x00%d\x00%d\x00%s\x00%s", v.ID, v.TranscodedPath, info.Size(), info.ModTime().UnixNano(), v.OriginalPath, v.ThumbnailPath)))
	return hex.EncodeToString(h[:])
}
func rotationSpaceEstimate(p ffmpeg.RotationProbe, size int64, mode string) int64 {
	estimate := size
	if mode == "reencode" {
		bitrate := p.Bitrate
		if bitrate < 500000 {
			bitrate = 4000000
		}
		candidate := int64(p.Duration*float64(bitrate)/8*1.3) + size/4
		if candidate > estimate {
			estimate = candidate
		}
	}
	return estimate + estimate/5 + 64*1024*1024
}
func checkRotationSpace(dir string, required int64) error {
	var st unix.Statfs_t
	if err := unix.Statfs(dir, &st); err != nil {
		return fmt.Errorf("无法读取可用磁盘空间")
	}
	if uint64(st.Bavail)*uint64(st.Bsize) < uint64(required) {
		return fmt.Errorf("磁盘空间不足，至少需要约 %.1f MB 临时空间", float64(required)/(1024*1024))
	}
	return nil
}

// Submit requires the exact preview version and explicit re-encoding consent.
func (s *VideoRotationService) Submit(ctx context.Context, id, user, requestID uuid.UUID, angle int, mode, version string, confirmed bool) (models.VideoRotationJob, error) {
	if angle != 90 && angle != 180 && angle != 270 {
		return models.VideoRotationJob{}, fmt.Errorf("仅支持顺时针 90°、180°、逆时针 90°")
	}
	if mode != "lossless" && mode != "reencode" {
		return models.VideoRotationJob{}, fmt.Errorf("处理方式无效")
	}
	if mode == "reencode" && !confirmed {
		return models.VideoRotationJob{}, fmt.Errorf("请先确认重新编码的画质及耗时影响")
	}
	// Idempotent requests remain queryable after their old input has been removed.
	if old, err := s.repo.RotationByID(ctx, requestID); err == nil {
		if old.VideoID != id || old.RequestedBy != user || old.Angle != angle || old.Mode != mode {
			return old, fmt.Errorf("请求标识已被其它操作使用")
		}
		return old, nil
	} else if !repository.IsNotFound(err) {
		return models.VideoRotationJob{}, err
	}
	unlock, err := s.repo.LockVideoFiles(ctx, id)
	if err != nil {
		return models.VideoRotationJob{}, err
	}
	defer unlock()
	v, err := s.repo.GetVideoByID(ctx, id)
	if err != nil {
		return models.VideoRotationJob{}, err
	}
	p, info, err := s.probe(ctx, v)
	if err != nil {
		return models.VideoRotationJob{}, err
	}
	if version == "" || version != rotationVersion(v, info) {
		return models.VideoRotationJob{}, fmt.Errorf("视频文件已变化，请重新预检并确认")
	}
	if mode == "lossless" && !p.Lossless {
		return models.VideoRotationJob{}, fmt.Errorf("该视频需要重新编码，请重新预检并确认")
	}
	if err = checkRotationSpace(filepath.Dir(v.TranscodedPath), rotationSpaceEstimate(p, info.Size(), mode)); err != nil {
		return models.VideoRotationJob{}, err
	}
	dir := filepath.Join(s.root, "videos", id.String())
	dir, err = filepath.Abs(dir)
	if err != nil {
		return models.VideoRotationJob{}, err
	}
	j := models.VideoRotationJob{ID: requestID, VideoID: id, RequestedBy: user, Angle: angle, Mode: mode, InputPath: v.TranscodedPath, InputSize: info.Size(), InputMtime: info.ModTime().UnixNano(), OriginalPath: v.OriginalPath, OldThumbnail: v.ThumbnailPath, OutputPath: filepath.Join(dir, "rotation-"+requestID.String()+".mp4"), ThumbnailPath: filepath.Join(dir, "rotation-"+requestID.String()+".jpg")}
	return s.repo.CreateRotation(ctx, j)
}

// Process resumes processing or cleanup while holding the cross-process file lock.
func (s *VideoRotationService) Process(ctx context.Context, id uuid.UUID) error {
	j, err := s.repo.RotationByID(ctx, id)
	if err != nil {
		return err
	}
	unlock, err := s.repo.LockVideoFiles(ctx, j.VideoID)
	if err != nil {
		return err
	}
	defer unlock()
	j, err = s.repo.RotationByID(ctx, id)
	if err != nil {
		return err
	}
	if j.Status == "succeeded" || j.Status == "failed" {
		return nil
	}
	if j.Status == "cleanup" {
		return s.cleanup(ctx, j)
	}
	if err = s.repo.SetRotationState(ctx, id, "running", 0, ""); err != nil {
		return err
	}
	fail := func(cause error) error {
		stateCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		message := "方向修正失败，旧视频已保留；请重新预检后重试。"
		if ctx.Err() != nil {
			message = "处理被中断，旧视频已保留；请重新预检后重试。"
		}
		for _, path := range []string{j.OutputPath, j.ThumbnailPath} {
			if e := s.removeOwned(stateCtx, j.VideoID, path); e != nil {
				message += " 半成品清理失败，请检查存储权限。"
			}
		}
		if e := s.repo.SetRotationState(stateCtx, id, "failed", 0, message); e != nil {
			return fmt.Errorf("%w; 保存失败状态: %v", cause, e)
		}
		return cause
	}
	v, err := s.repo.GetVideoByID(ctx, j.VideoID)
	if err != nil {
		return fail(err)
	}
	p, info, err := s.probe(ctx, v)
	if err != nil {
		return fail(err)
	}
	if v.TranscodedPath != j.InputPath || info.Size() != j.InputSize || info.ModTime().UnixNano() != j.InputMtime {
		return fail(fmt.Errorf("输入文件版本已变化"))
	}
	if err = checkRotationSpace(filepath.Dir(j.OutputPath), rotationSpaceEstimate(p, info.Size(), j.Mode)); err != nil {
		return fail(err)
	}
	if err = s.ownedOutput(j); err != nil {
		return fail(err)
	}
	last := time.Time{}
	progress := func(p ffmpeg.TranscodeProgress) {
		if time.Since(last) < time.Second {
			return
		}
		last = time.Now()
		_ = s.repo.SetRotationState(ctx, id, "running", math.Min(p.ProgressPercent, 95), "")
	}
	if err = ffmpeg.RotateVideo(ctx, j.InputPath, j.OutputPath, j.Angle, j.Mode, p, progress); err != nil {
		return fail(err)
	}
	out, err := ffmpeg.ProbeRotation(ctx, j.OutputPath)
	if err != nil {
		return fail(err)
	}
	width, height := p.DisplaySize()
	if j.Angle%180 != 0 {
		width, height = height, width
	}
	ow, oh := out.DisplaySize()
	if ow != width || oh != height || math.Abs(out.Duration-p.Duration) > math.Max(1, p.Duration*.01) || out.AudioTracks != p.AudioTracks {
		return fail(fmt.Errorf("修正文件校验失败"))
	}
	if j.Mode == "lossless" && (out.Codec != p.Codec || out.Width != p.Width || out.Height != p.Height || out.Rotation != (p.Rotation-j.Angle+360)%360) {
		return fail(fmt.Errorf("无损旋转方向校验失败"))
	}
	if j.Mode == "reencode" && out.Rotation != 0 {
		return fail(fmt.Errorf("重编码输出仍包含方向标记"))
	}
	thumb := v.ThumbnailPath
	if automaticRotationThumbnail(v.ThumbnailPath) && !manualRotationThumbnail(v.Metadata) {
		if err = ffmpeg.Thumbnail(ctx, j.OutputPath, j.ThumbnailPath); err != nil {
			return fail(err)
		}
		thumb = j.ThumbnailPath
	}
	size, err := RequiredRegularFileSize(j.OutputPath)
	if err != nil {
		return fail(err)
	}
	for _, path := range []string{j.OutputPath, thumb} {
		if path == j.OutputPath || path == j.ThumbnailPath {
			if err = syncRotationFile(path); err != nil {
				return fail(err)
			}
		}
	}
	if err = syncRotationDirectory(filepath.Dir(j.OutputPath)); err != nil {
		return fail(err)
	}
	// Detect external replacement of the pinned input before publication.
	latest, err := os.Stat(j.InputPath)
	if err != nil || latest.Size() != j.InputSize || latest.ModTime().UnixNano() != j.InputMtime {
		return fail(fmt.Errorf("输入文件已被外部修改"))
	}
	sourceCompat, sourceErr := ffmpeg.ProbePlaybackCompatibility(ctx, j.InputPath)
	outputCompat, outputErr := ffmpeg.ProbePlaybackCompatibility(ctx, j.OutputPath)
	if sourceErr != nil || outputErr != nil {
		return fail(fmt.Errorf("兼容信息校验失败"))
	}
	patch := map[string]any{"playback_path": j.OutputPath, "codec": out.Codec, "playback_codec": out.Codec, PlaybackCompatibilityMetadataKey: BuildPlaybackCompatibilityMetadata(sourceCompat, sourceErr, outputCompat, outputErr), "direction_rotation": map[string]any{"job_id": j.ID, "mode": j.Mode, "display_rotation": out.Rotation}, "playback_file_revision": j.ID.String()}
	if err = s.repo.PublishRotation(ctx, j, size, width, height, thumb, patch); err != nil {
		// A commit response can be lost. Never remove output until the committed state is known.
		recoveryCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		saved, e := s.repo.RotationByID(recoveryCtx, id)
		if e != nil {
			return err
		}
		if saved.Status == "cleanup" {
			return s.cleanup(recoveryCtx, saved)
		}
		return fail(err)
	}
	j.Status = "cleanup"
	return s.cleanup(ctx, j)
}

func manualRotationThumbnail(raw []byte) bool {
	var metadata map[string]any
	if json.Unmarshal(raw, &metadata) != nil {
		return true
	}
	return metadata["thumbnail_source"] == "manual"
}

func automaticRotationThumbnail(path string) bool {
	if strings.Contains(path, "://") {
		return false
	}
	base := filepath.Base(path)
	if path == "" || base == "thumb.jpg" {
		return true
	}
	if !strings.HasPrefix(base, "rotation-") || !strings.HasSuffix(base, ".jpg") {
		return false
	}
	_, err := uuid.Parse(strings.TrimSuffix(strings.TrimPrefix(base, "rotation-"), ".jpg"))
	return err == nil
}
func syncRotationDirectory(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}
func syncRotationFile(path string) error {
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}
func (s *VideoRotationService) ownedOutput(j models.VideoRotationJob) error {
	for _, p := range []string{j.OutputPath, j.ThumbnailPath} {
		if err := s.ownedPath(j.VideoID, p); err != nil {
			return err
		}
	}
	return nil
}
func (s *VideoRotationService) ownedPath(id uuid.UUID, path string) error {
	base, err := filepath.Abs(filepath.Join(s.root, "videos", id.String()))
	if err != nil {
		return err
	}
	actual, err := filepath.Abs(path)
	if err != nil || path == "" {
		return fmt.Errorf("视频文件不在受管目录中")
	}
	rel, err := filepath.Rel(base, actual)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("视频文件不在该视频的独占存储目录中")
	}
	realBase, err := filepath.EvalSymlinks(base)
	if err != nil {
		return err
	}
	trustedRoot, err := filepath.Abs(filepath.Join(s.root, "videos"))
	if err != nil {
		return err
	}
	trustedRoot, err = filepath.EvalSymlinks(trustedRoot)
	if err != nil {
		return err
	}
	if realBase != filepath.Join(trustedRoot, id.String()) {
		return fmt.Errorf("视频独占目录不能是软链接")
	}
	realParent, err := filepath.EvalSymlinks(filepath.Dir(actual))
	if err != nil {
		return err
	}
	rel, err = filepath.Rel(realBase, realParent)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("视频目录包含越界软链接")
	}
	if st, err := os.Lstat(actual); err == nil && st.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("暂不支持软链接视频文件")
	}
	return nil
}
func (s *VideoRotationService) removeOwned(ctx context.Context, id uuid.UUID, path string) error {
	if path == "" {
		return nil
	}
	if err := s.ownedPath(id, path); err != nil {
		return err
	}
	if info, err := os.Lstat(path); err == nil && !info.Mode().IsRegular() {
		return fmt.Errorf("拒绝清理非普通文件")
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	used, err := s.repo.RotationFileReferenced(ctx, path)
	if err != nil {
		return err
	}
	if used {
		return fmt.Errorf("旧文件仍被资源引用")
	}
	if err = os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
func (s *VideoRotationService) cleanup(ctx context.Context, j models.VideoRotationJob) error {
	current, err := s.repo.GetVideoByID(ctx, j.VideoID)
	if err != nil {
		return err
	}
	paths := []string{j.InputPath}
	// Shared/external import originals and custom posters are outside rotation ownership.
	if j.OriginalPath != "" && s.ownedPath(j.VideoID, j.OriginalPath) == nil {
		paths = append(paths, j.OriginalPath)
	}
	if automaticRotationThumbnail(j.OldThumbnail) && j.OldThumbnail != "" {
		paths = append(paths, j.OldThumbnail)
	}
	for _, path := range paths {
		if path == j.OutputPath || path == j.ThumbnailPath || path == current.ThumbnailPath {
			continue
		}
		if err := s.removeOwned(ctx, j.VideoID, path); err != nil {
			stateCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			defer cancel()
			if e := s.repo.SetRotationState(stateCtx, j.ID, "cleanup", 100, "画面已修正，旧文件清理待重试；请检查存储权限及共享引用。"); e != nil {
				return e
			}
			return err
		}
	}
	return s.repo.SetRotationState(ctx, j.ID, "succeeded", 100, "")
}
