package handlers

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"

	"video-server/internal/middleware"
	"video-server/internal/models"
	"video-server/internal/queue"
	"video-server/internal/repository"
	"video-server/internal/response"
	"video-server/internal/utils"
)

func (a *API) AdminStats(c *gin.Context) {
	shorts, movies, episodes, err := a.repo.CountVideosByType(c.Request.Context())
	if err != nil {
		response.Error(c, 1001, err.Error())
		return
	}
	totalUsers, err := a.repo.CountUsers(c.Request.Context())
	if err != nil {
		response.Error(c, 1002, err.Error())
		return
	}
	todayUploads, err := a.repo.CountTodayUploads(c.Request.Context())
	if err != nil {
		response.Error(c, 1003, err.Error())
		return
	}
	trend, err := a.repo.WeeklyUploadTrend(c.Request.Context())
	if err != nil {
		response.Error(c, 1004, err.Error())
		return
	}

	var queueLength int64
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: a.redisAddr, Password: a.redisPassword})
	if qi, err := inspector.GetQueueInfo(a.asynqQueue); err == nil {
		queueLength = int64(qi.Pending + qi.Active + qi.Retry + qi.Scheduled)
	}
	_ = inspector.Close()

	var diskTotal uint64
	var diskFree uint64
	var fs syscall.Statfs_t
	if err := syscall.Statfs(a.storageRoot, &fs); err == nil {
		diskTotal = fs.Blocks * uint64(fs.Bsize)
		diskFree = fs.Bavail * uint64(fs.Bsize)
	}

	ok(c, models.AdminStats{
		TotalVideos:       shorts + movies + episodes,
		ShortVideos:       shorts,
		MovieVideos:       movies,
		EpisodeVideos:     episodes,
		TotalUsers:        totalUsers,
		TodayUploads:      todayUploads,
		QueueLength:       queueLength,
		DiskTotalBytes:    diskTotal,
		DiskFreeBytes:     diskFree,
		WeeklyUploadTrend: trend,
	})
}

func (a *API) AdminVideos(c *gin.Context) {
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 20)

	var startTime *time.Time
	var endTime *time.Time
	if v := strings.TrimSpace(c.Query("start_time")); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			startTime = &t
		}
	}
	if v := strings.TrimSpace(c.Query("end_time")); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			endTime = &t
		}
	}

	items, total, err := a.repo.AdminListVideos(c.Request.Context(), models.AdminVideoFilter{
		Page:      page,
		PageSize:  pageSize,
		Keyword:   c.Query("q"),
		Type:      c.Query("type"),
		Status:    c.Query("status"),
		User:      c.Query("user"),
		Tag:       c.Query("tag"),
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		response.Error(c, 1005, err.Error())
		return
	}
	ok(c, gin.H{
		"items":       items,
		"total_count": total,
		"page":        page,
		"page_size":   pageSize,
	})
}

func (a *API) AdminVideoDetail(c *gin.Context) {
	videoID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "invalid video id")
		return
	}
	detail, err := a.repo.AdminVideoDetail(c.Request.Context(), videoID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.Error(c, 404, "video not found")
			return
		}
		response.Error(c, 1006, err.Error())
		return
	}
	ok(c, detail)
}

func (a *API) AdminVideoPlayURL(c *gin.Context) {
	videoID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "invalid video id")
		return
	}
	video, err := a.repo.GetVideoByID(c.Request.Context(), videoID)
	if err != nil {
		if repository.IsNotFound(err) {
			response.Error(c, 404, "video not found")
			return
		}
		response.Error(c, 1022, err.Error())
		return
	}
	if video.Status != "ready" {
		response.Error(c, 1023, "video not ready")
		return
	}

	exp := time.Now().Add(a.playSignTTL).Unix()
	sig := utils.SignVideoSource(a.playSignSecret, videoID, exp)
	ok(c, gin.H{
		"signed_url": fmt.Sprintf("/api/v1/videos/%s/source/signed?exp=%d&sig=%s", videoID.String(), exp, sig),
		"expires_at": exp,
	})
}

func (a *API) AdminUpdateVideo(c *gin.Context) {
	videoID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "invalid video id")
		return
	}
	var req struct {
		Title       string         `json:"title"`
		Description string         `json:"description"`
		Thumbnail   string         `json:"thumbnail"`
		Tags        []string       `json:"tags"`
		ActorIDs    *[]string      `json:"actor_ids"`
		ActorNames  *[]string      `json:"actor_names"`
		Status      string         `json:"status"`
		Metadata    map[string]any `json:"metadata"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	if req.Metadata == nil {
		req.Metadata = map[string]any{}
	}
	updateActors := req.ActorIDs != nil || req.ActorNames != nil
	var actorIDs []uuid.UUID
	var actorNames []string
	if updateActors {
		var err error
		if req.ActorIDs != nil {
			actorIDs, err = parseUUIDStrings(*req.ActorIDs)
			if err != nil {
				bad(c, "演员ID格式错误")
				return
			}
		}
		if req.ActorNames != nil {
			actorNames = normalizeActorNames(*req.ActorNames)
		}
	}
	if err := a.repo.AdminUpdateVideo(c.Request.Context(), videoID, req.Title, req.Description, req.Thumbnail, req.Tags, req.Metadata, actorIDs, actorNames, updateActors); err != nil {
		response.Error(c, 1007, err.Error())
		return
	}
	if req.Status != "" {
		if err := a.repo.AdminUpdateVideoStatus(c.Request.Context(), videoID, req.Status); err != nil {
			response.Error(c, 1008, err.Error())
			return
		}
	}
	ok(c, gin.H{"updated": true})
}

func (a *API) AdminDeleteVideo(c *gin.Context) {
	videoID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "invalid video id")
		return
	}
	video, err := a.repo.GetVideoByID(c.Request.Context(), videoID)
	if err != nil {
		response.Error(c, 1009, err.Error())
		return
	}
	if err := a.repo.DeleteVideoByID(c.Request.Context(), videoID); err != nil {
		response.Error(c, 1010, err.Error())
		return
	}
	_ = os.Remove(video.OriginalPath)
	_ = os.Remove(video.TranscodedPath)
	_ = os.Remove(video.ThumbnailPath)
	ok(c, gin.H{"deleted": true, "video_id": videoID})
}

func (a *API) AdminRetranscodeVideo(c *gin.Context) {
	videoID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "invalid video id")
		return
	}
	video, err := a.repo.GetVideoByID(c.Request.Context(), videoID)
	if err != nil {
		response.Error(c, 1011, err.Error())
		return
	}
	if strings.TrimSpace(video.OriginalPath) == "" {
		response.Error(c, 1012, "video has no original path")
		return
	}
	if err := a.repo.UpdateVideoStatus(c.Request.Context(), videoID, "uploaded"); err != nil {
		response.Error(c, 1013, err.Error())
		return
	}

	payload := queue.TranscodePayload{
		VideoID:      videoID.String(),
		InputPath:    video.OriginalPath,
		OutputDir:    filepath.Join(a.storageRoot, "videos"),
		TargetFormat: "mp4",
	}
	if err := a.enqueuer.EnqueueTranscode(payload); err != nil {
		response.Error(c, 1014, err.Error())
		return
	}
	ok(c, gin.H{"enqueued": true, "video_id": videoID})
}

func (a *API) AdminUsers(c *gin.Context) {
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 20)
	items, total, err := a.repo.AdminListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, 1015, err.Error())
		return
	}
	ok(c, gin.H{
		"items":       items,
		"total_count": total,
		"page":        page,
		"page_size":   pageSize,
	})
}

func (a *API) AdminActors(c *gin.Context) {
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 20)
	keyword := c.Query("q")
	var active *bool
	if raw := strings.TrimSpace(c.Query("active")); raw != "" {
		switch strings.ToLower(raw) {
		case "1", "true":
			v := true
			active = &v
		case "0", "false":
			v := false
			active = &v
		default:
			bad(c, "active 参数只能是 1 或 0")
			return
		}
	}

	items, total, err := a.repo.ListActors(c.Request.Context(), keyword, active, page, pageSize)
	if err != nil {
		response.Error(c, 1024, err.Error())
		return
	}
	ok(c, gin.H{
		"items":       items,
		"total_count": total,
		"page":        page,
		"page_size":   pageSize,
	})
}

func (a *API) AdminCreateActor(c *gin.Context) {
	var req models.AdminActorInput
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	actor, err := a.repo.CreateActor(c.Request.Context(), req)
	if err != nil {
		if repository.IsUniqueViolation(err) {
			payload := gin.H{
				"existing_actor_name": strings.TrimSpace(req.Name),
				"reason":              "duplicate_name",
			}
			if existing, lookupErr := a.repo.GetActorByName(c.Request.Context(), req.Name); lookupErr == nil {
				payload["existing_actor_id"] = existing.ID
				payload["existing_actor"] = existing
				payload["existing_actor_name"] = existing.Name
			}
			response.JSON(c, 1025, "演员名称已存在", payload)
			return
		}
		response.Error(c, 1028, err.Error())
		return
	}
	ok(c, actor)
}

func (a *API) AdminUpdateActor(c *gin.Context) {
	actorID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "演员ID格式错误")
		return
	}

	var req models.AdminActorInput
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}

	actor, err := a.repo.UpdateActor(c.Request.Context(), actorID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.Error(c, 404, "演员不存在")
			return
		}
		if repository.IsUniqueViolation(err) {
			response.Error(c, 1026, "演员名称已存在")
			return
		}
		response.Error(c, 1026, err.Error())
		return
	}
	ok(c, actor)
}

func (a *API) AdminUpdateUserRole(c *gin.Context) {
	userID, okID := parseUUID(c.Param("id"))
	if !okID {
		bad(c, "invalid user id")
		return
	}
	var req struct {
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, "invalid payload")
		return
	}
	req.Role = strings.ToLower(strings.TrimSpace(req.Role))
	if req.Role != "user" && req.Role != "admin" {
		bad(c, "role must be user or admin")
		return
	}
	callerID, found := middleware.UserIDFromContext(c)
	if found && callerID == userID && req.Role != "admin" {
		response.Error(c, 1016, "cannot downgrade current admin account")
		return
	}
	if err := a.repo.AdminUpdateUserRole(c.Request.Context(), userID, req.Role); err != nil {
		response.Error(c, 1017, err.Error())
		return
	}
	ok(c, gin.H{"updated": true})
}

func (a *API) AdminTasks(c *gin.Context) {
	page := parsePage(c.Query("page"), 1)
	pageSize := parsePageSize(c.Query("page_size"), 20)
	items, total, err := a.repo.AdminListTranscodingTasks(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, 1018, err.Error())
		return
	}
	ok(c, gin.H{
		"items":       items,
		"total_count": total,
		"page":        page,
		"page_size":   pageSize,
	})
}

func (a *API) AdminSystemCleanup(c *gin.Context) {
	var req struct {
		OlderThanHours int `json:"older_than_hours"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.OlderThanHours <= 0 {
		req.OlderThanHours = 24
	}
	deadline := time.Now().Add(-time.Duration(req.OlderThanHours) * time.Hour)
	deleted := 0
	err := filepath.WalkDir(a.uploadTempDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, statErr := d.Info()
		if statErr != nil {
			return nil
		}
		if info.ModTime().Before(deadline) {
			if rmErr := os.Remove(path); rmErr == nil {
				deleted++
			}
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		response.Error(c, 1019, err.Error())
		return
	}
	ok(c, gin.H{"deleted": deleted, "older_than_hours": req.OlderThanHours})
}

func (a *API) AdminSystemLogs(c *gin.Context) {
	file, err := os.Open(a.serverLogPath)
	if err != nil {
		response.Error(c, 1020, err.Error())
		return
	}
	defer file.Close()

	if c.Query("download") == "1" {
		c.Header("Content-Disposition", "attachment; filename=server.log")
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.Status(http.StatusOK)
		_, _ = io.Copy(c.Writer, file)
		return
	}

	maxLines := 1000
	if raw := c.Query("lines"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 && n <= 5000 {
			maxLines = n
		}
	}

	lines, err := tailLines(file, maxLines)
	if err != nil {
		response.Error(c, 1021, err.Error())
		return
	}
	ok(c, gin.H{"lines": lines, "line_count": len(lines)})
}

func tailLines(f *os.File, n int) ([]string, error) {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(f)
	lines := make([]string, 0, n)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > n {
			lines = lines[1:]
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func parseUUIDStrings(raw []string) ([]uuid.UUID, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	out := make([]uuid.UUID, 0, len(raw))
	seen := map[uuid.UUID]struct{}{}
	for _, item := range raw {
		v := strings.TrimSpace(item)
		if v == "" {
			continue
		}
		id, err := uuid.Parse(v)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out, nil
}
