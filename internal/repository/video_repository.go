package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"video-server/internal/models"
)

// VideoRepository handles video persistence and feed queries.
type VideoRepository struct {
	pool *pgxpool.Pool
}

func NewVideoRepository(pool *pgxpool.Pool) *VideoRepository {
	return &VideoRepository{pool: pool}
}

func (r *VideoRepository) CreateVideo(ctx context.Context, v models.Video) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO videos (id, user_id, tmdb_id, title, description, type, status, original_path, metadata)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
`, v.ID, v.UserID, v.TMDBID, v.Title, v.Description, v.Type, v.Status, v.OriginalPath, v.Metadata)
	if err != nil {
		return fmt.Errorf("insert video: %w", err)
	}
	return nil
}

func (r *VideoRepository) AddTags(ctx context.Context, videoID uuid.UUID, tags []string) error {
	if len(tags) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for _, tag := range tags {
		t := strings.TrimSpace(strings.ToLower(tag))
		if t == "" {
			continue
		}
		batch.Queue("INSERT INTO video_tags(video_id, tag, weight) VALUES ($1,$2,1.0) ON CONFLICT(video_id, tag) DO NOTHING", videoID, t)
	}
	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()
	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("insert tag: %w", err)
		}
	}
	return nil
}

func (r *VideoRepository) FindVideoByHash(ctx context.Context, hash string, fileSize int64) (uuid.UUID, bool, error) {
	var videoID uuid.UUID
	err := r.pool.QueryRow(ctx, `
SELECT video_id
FROM file_hashes
WHERE hash = $1 AND file_size = $2
`, hash, fileSize).Scan(&videoID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, false, nil
		}
		return uuid.Nil, false, fmt.Errorf("find video by hash: %w", err)
	}
	return videoID, true, nil
}

func (r *VideoRepository) InsertFileHash(ctx context.Context, hash string, videoID uuid.UUID, fileSize int64) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO file_hashes(hash, video_id, file_size)
VALUES ($1,$2,$3)
`, hash, videoID, fileSize)
	if err != nil {
		return fmt.Errorf("insert file hash: %w", err)
	}
	return nil
}

func (r *VideoRepository) GetVideoIDByHash(ctx context.Context, hash string) (uuid.UUID, error) {
	var videoID uuid.UUID
	err := r.pool.QueryRow(ctx, `
SELECT video_id
FROM file_hashes
WHERE hash = $1
`, hash).Scan(&videoID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("get video id by hash: %w", err)
	}
	return videoID, nil
}

func (r *VideoRepository) DeleteVideoByID(ctx context.Context, videoID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx for delete video: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := deleteVideoDependencies(ctx, tx, videoID); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit delete video tx: %w", err)
	}
	return nil
}

type execer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func deleteVideoDependencies(ctx context.Context, db execer, videoID uuid.UUID) error {
	if _, err := db.Exec(ctx, "DELETE FROM transcoding_jobs WHERE video_id = $1", videoID); err != nil {
		return fmt.Errorf("delete transcoding jobs by video id: %w", err)
	}
	if _, err := db.Exec(ctx, "DELETE FROM user_video_actions WHERE video_id = $1", videoID); err != nil {
		return fmt.Errorf("delete user actions by video id: %w", err)
	}
	if _, err := db.Exec(ctx, "DELETE FROM videos WHERE id = $1", videoID); err != nil {
		return fmt.Errorf("delete video by id: %w", err)
	}
	return nil
}

func (r *VideoRepository) RandomShorts(ctx context.Context, limit int) ([]models.RecommendedVideo, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id, title, type, transcoded_path, duration_seconds
FROM videos
WHERE type = 'short' AND status = 'ready'
ORDER BY random()
LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("query random shorts: %w", err)
	}
	defer rows.Close()

	out := make([]models.RecommendedVideo, 0, limit)
	for rows.Next() {
		var v models.RecommendedVideo
		if err := rows.Scan(&v.ID, &v.Title, &v.Type, &v.TranscodedPath, &v.Duration); err != nil {
			return nil, fmt.Errorf("scan random short: %w", err)
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *VideoRepository) UpsertAction(ctx context.Context, userID, videoID uuid.UUID, action string, watchSeconds int) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO user_video_actions(user_id, video_id, action_type, watch_seconds)
VALUES ($1,$2,$3,$4)
ON CONFLICT(user_id, video_id, action_type)
DO UPDATE SET watch_seconds = EXCLUDED.watch_seconds, updated_at = NOW()
`, userID, videoID, action, watchSeconds)
	if err != nil {
		return fmt.Errorf("upsert user action: %w", err)
	}
	return nil
}

func (r *VideoRepository) UpdateTranscodeResult(ctx context.Context, videoID uuid.UUID, transcodedPath, thumbPath string, duration, width, height int, metadata map[string]any) error {
	metaRaw, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}
	_, err = r.pool.Exec(ctx, `
UPDATE videos
SET transcoded_path=$2, thumbnail_path=$3, duration_seconds=$4, width=$5, height=$6, metadata=$7, status='ready', updated_at=NOW()
WHERE id=$1`, videoID, transcodedPath, thumbPath, duration, width, height, metaRaw)
	if err != nil {
		return fmt.Errorf("update transcode result: %w", err)
	}
	return nil
}

func (r *VideoRepository) MarkVideoFailed(ctx context.Context, videoID uuid.UUID, reason string) error {
	_, err := r.pool.Exec(ctx, `
UPDATE videos
SET status='failed', metadata = coalesce(metadata, '{}'::jsonb) || jsonb_build_object('error', $2), updated_at=NOW()
WHERE id=$1`, videoID, reason)
	if err != nil {
		return fmt.Errorf("mark video failed: %w", err)
	}
	return nil
}

func (r *VideoRepository) UpdateVideoStatus(ctx context.Context, videoID uuid.UUID, status string) error {
	_, err := r.pool.Exec(ctx, "UPDATE videos SET status=$2, updated_at=NOW() WHERE id=$1", videoID, status)
	if err != nil {
		return fmt.Errorf("update video status: %w", err)
	}
	return nil
}

func (r *VideoRepository) GetVideoByID(ctx context.Context, videoID uuid.UUID) (models.Video, error) {
	var v models.Video
	var durationSeconds sql.NullInt32
	var width sql.NullInt32
	var height sql.NullInt32
	err := r.pool.QueryRow(ctx, `
SELECT
  id,
  user_id,
  tmdb_id,
  COALESCE(title, ''),
  COALESCE(description, ''),
  type,
  status,
  COALESCE(duration_seconds, 0),
  COALESCE(width, 0),
  COALESCE(height, 0),
  COALESCE(original_path, ''),
  COALESCE(transcoded_path, ''),
  COALESCE(thumbnail_path, ''),
  COALESCE(metadata, '{}'::jsonb),
  created_at,
  updated_at
FROM videos WHERE id=$1`, videoID).Scan(
		&v.ID, &v.UserID, &v.TMDBID, &v.Title, &v.Description, &v.Type, &v.Status, &durationSeconds, &width, &height,
		&v.OriginalPath, &v.TranscodedPath, &v.ThumbnailPath, &v.Metadata, &v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		return models.Video{}, fmt.Errorf("get video by id: %w", err)
	}
	v.DurationSeconds = nullInt32ToInt(durationSeconds)
	v.Width = nullInt32ToInt(width)
	v.Height = nullInt32ToInt(height)
	return v, nil
}

func (r *VideoRepository) ListActiveOriginalPaths(ctx context.Context) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
SELECT original_path
FROM videos
WHERE status IN ('uploaded','scraping','processing','failed') AND original_path IS NOT NULL AND original_path <> ''
`)
	if err != nil {
		return nil, fmt.Errorf("list active original paths: %w", err)
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, fmt.Errorf("scan active original path: %w", err)
		}
		paths = append(paths, path)
	}
	return paths, rows.Err()
}

func (r *VideoRepository) GetVideoByOriginalPath(ctx context.Context, originalPath string) (models.Video, error) {
	var v models.Video
	var durationSeconds sql.NullInt32
	var width sql.NullInt32
	var height sql.NullInt32
	err := r.pool.QueryRow(ctx, `
SELECT
  id,
  user_id,
  tmdb_id,
  COALESCE(title, ''),
  COALESCE(description, ''),
  type,
  status,
  COALESCE(duration_seconds, 0),
  COALESCE(width, 0),
  COALESCE(height, 0),
  COALESCE(original_path, ''),
  COALESCE(transcoded_path, ''),
  COALESCE(thumbnail_path, ''),
  COALESCE(metadata, '{}'::jsonb),
  created_at,
  updated_at
FROM videos WHERE original_path=$1`, originalPath).Scan(
		&v.ID, &v.UserID, &v.TMDBID, &v.Title, &v.Description, &v.Type, &v.Status, &durationSeconds, &width, &height,
		&v.OriginalPath, &v.TranscodedPath, &v.ThumbnailPath, &v.Metadata, &v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		return models.Video{}, fmt.Errorf("get video by original path: %w", err)
	}
	v.DurationSeconds = nullInt32ToInt(durationSeconds)
	v.Width = nullInt32ToInt(width)
	v.Height = nullInt32ToInt(height)
	return v, nil
}

func nullInt32ToInt(v sql.NullInt32) int {
	if !v.Valid {
		return 0
	}
	return int(v.Int32)
}

func (r *VideoRepository) UpdateVideoMetadata(ctx context.Context, videoID uuid.UUID, title, description string, metadata map[string]any) error {
	metaRaw, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}
	_, err = r.pool.Exec(ctx, `
UPDATE videos
SET title=$2, description=$3, metadata=$4, updated_at=NOW()
WHERE id=$1`, videoID, title, description, metaRaw)
	if err != nil {
		return fmt.Errorf("update video metadata: %w", err)
	}
	return nil
}

func (r *VideoRepository) UpdateVideoScrapeResult(ctx context.Context, videoID uuid.UUID, tmdbID *int, title, description, thumbnailPath string, metadata map[string]any, status string) error {
	metaRaw, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal scrape metadata: %w", err)
	}
	_, err = r.pool.Exec(ctx, `
UPDATE videos
SET tmdb_id=$2, title=$3, description=$4, thumbnail_path=$5, metadata=$6, status=$7, updated_at=NOW()
WHERE id=$1
`, videoID, tmdbID, title, description, thumbnailPath, metaRaw, status)
	if err != nil {
		return fmt.Errorf("update scrape result: %w", err)
	}
	return nil
}

func (r *VideoRepository) AppendVideoMetadata(ctx context.Context, videoID uuid.UUID, key string, value any) error {
	_, err := r.pool.Exec(ctx, `
UPDATE videos
SET metadata = coalesce(metadata, '{}'::jsonb) || jsonb_build_object($2, to_jsonb($3)),
    updated_at = NOW()
WHERE id=$1
`, videoID, key, value)
	if err != nil {
		return fmt.Errorf("append video metadata: %w", err)
	}
	return nil
}

func (r *VideoRepository) FindVideoByTypeTMDB(ctx context.Context, typ string, tmdbID int, excludeVideoID uuid.UUID) (uuid.UUID, bool, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
SELECT id
FROM videos
WHERE type=$1 AND tmdb_id=$2 AND id <> $3
LIMIT 1
`, typ, tmdbID, excludeVideoID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, false, nil
		}
		return uuid.Nil, false, fmt.Errorf("find video by tmdb: %w", err)
	}
	return id, true, nil
}

func (r *VideoRepository) FetchUserTagAffinity(ctx context.Context, userID uuid.UUID, since time.Time, limit int) (map[string]float64, error) {
	rows, err := r.pool.Query(ctx, `
SELECT vt.tag,
       SUM(CASE
            WHEN uva.action_type IN ('like','favorite','view') THEN vt.weight * GREATEST(uva.watch_seconds, 1)
            WHEN uva.action_type IN ('dislike','skip') THEN -vt.weight * GREATEST(uva.watch_seconds, 1)
            ELSE 0 END) AS score
FROM user_video_actions uva
JOIN video_tags vt ON vt.video_id = uva.video_id
WHERE uva.user_id = $1 AND uva.created_at >= $2
GROUP BY vt.tag
ORDER BY score DESC
LIMIT $3`, userID, since, limit)
	if err != nil {
		return nil, fmt.Errorf("fetch user tag affinity: %w", err)
	}
	defer rows.Close()

	result := map[string]float64{}
	for rows.Next() {
		var tag string
		var score float64
		if err := rows.Scan(&tag, &score); err != nil {
			return nil, fmt.Errorf("scan affinity: %w", err)
		}
		result[tag] = score
	}
	return result, rows.Err()
}

func (r *VideoRepository) FetchCandidateVideos(ctx context.Context, userID uuid.UUID, tags []string, limit int) ([]models.RecommendedVideo, error) {
	if len(tags) == 0 {
		return r.FetchHotVideos(ctx, limit)
	}
	rows, err := r.pool.Query(ctx, `
WITH excluded AS (
	SELECT DISTINCT video_id FROM user_video_actions WHERE user_id = $1 AND action_type IN ('view','dislike','skip')
)
SELECT DISTINCT v.id, v.title, v.type, v.transcoded_path, v.duration_seconds
FROM videos v
JOIN video_tags vt ON vt.video_id = v.id
WHERE v.status = 'ready'
  AND vt.tag = ANY($2)
  AND NOT EXISTS (SELECT 1 FROM excluded e WHERE e.video_id = v.id)
ORDER BY v.created_at DESC
LIMIT $3`, userID, tags, limit)
	if err != nil {
		return nil, fmt.Errorf("fetch candidate videos: %w", err)
	}
	defer rows.Close()

	out := make([]models.RecommendedVideo, 0, limit)
	for rows.Next() {
		var v models.RecommendedVideo
		if err := rows.Scan(&v.ID, &v.Title, &v.Type, &v.TranscodedPath, &v.Duration); err != nil {
			return nil, fmt.Errorf("scan candidate: %w", err)
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *VideoRepository) FetchHotVideos(ctx context.Context, limit int) ([]models.RecommendedVideo, error) {
	rows, err := r.pool.Query(ctx, `
SELECT v.id, v.title, v.type, v.transcoded_path, v.duration_seconds
FROM videos v
LEFT JOIN LATERAL (
	SELECT COUNT(*)::int AS views
	FROM user_video_actions uva
	WHERE uva.video_id = v.id AND uva.action_type = 'view' AND uva.created_at >= NOW() - INTERVAL '30 days'
) s ON true
WHERE v.status = 'ready'
ORDER BY COALESCE(s.views,0) DESC, v.created_at DESC
LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("fetch hot videos: %w", err)
	}
	defer rows.Close()

	out := make([]models.RecommendedVideo, 0, limit)
	for rows.Next() {
		var v models.RecommendedVideo
		if err := rows.Scan(&v.ID, &v.Title, &v.Type, &v.TranscodedPath, &v.Duration); err != nil {
			return nil, fmt.Errorf("scan hot video: %w", err)
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *VideoRepository) InsertTranscodingJob(ctx context.Context, videoID uuid.UUID, userID *uuid.UUID, status string) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
INSERT INTO transcoding_jobs(video_id, user_id, status, started_at)
VALUES ($1,$2,$3,NOW()) RETURNING id`, videoID, userID, status).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert transcoding job: %w", err)
	}
	return id, nil
}

func (r *VideoRepository) FinishTranscodingJob(ctx context.Context, jobID int64, status, errMsg string) error {
	_, err := r.pool.Exec(ctx, `
UPDATE transcoding_jobs
SET status=$2, error_message=$3, finished_at=NOW()
WHERE id=$1`, jobID, status, errMsg)
	if err != nil {
		return fmt.Errorf("finish transcoding job: %w", err)
	}
	return nil
}

func (r *VideoRepository) UpsertSeries(ctx context.Context, tmdbID int, title, overview, poster, backdrop string, firstAirDate *time.Time, seasons, episodes int) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
INSERT INTO series(tmdb_id, title, overview, poster_path, backdrop_path, first_air_date, number_of_seasons, number_of_episodes)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
ON CONFLICT(tmdb_id)
DO UPDATE SET title=EXCLUDED.title, overview=EXCLUDED.overview, poster_path=EXCLUDED.poster_path,
              backdrop_path=EXCLUDED.backdrop_path, first_air_date=EXCLUDED.first_air_date,
              number_of_seasons=EXCLUDED.number_of_seasons, number_of_episodes=EXCLUDED.number_of_episodes
RETURNING id`, tmdbID, title, overview, poster, backdrop, firstAirDate, seasons, episodes).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("upsert series: %w", err)
	}
	return id, nil
}

func (r *VideoRepository) UpsertSeason(ctx context.Context, seriesID int64, seasonNumber int, name, overview, poster string, airDate *time.Time) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
INSERT INTO seasons(series_id, season_number, name, overview, poster_path, air_date)
VALUES ($1,$2,$3,$4,$5,$6)
ON CONFLICT(series_id, season_number)
DO UPDATE SET name=EXCLUDED.name, overview=EXCLUDED.overview, poster_path=EXCLUDED.poster_path, air_date=EXCLUDED.air_date
RETURNING id`, seriesID, seasonNumber, name, overview, poster, airDate).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("upsert season: %w", err)
	}
	return id, nil
}

func (r *VideoRepository) UpsertEpisode(ctx context.Context, seasonID int64, episodeNumber int, title, overview, still string, runtime int, airDate *time.Time, videoID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO episodes(season_id, episode_number, title, overview, still_path, runtime, air_date, video_id)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
ON CONFLICT(season_id, episode_number)
DO UPDATE SET title=EXCLUDED.title, overview=EXCLUDED.overview, still_path=EXCLUDED.still_path,
              runtime=EXCLUDED.runtime, air_date=EXCLUDED.air_date, video_id=EXCLUDED.video_id
`, seasonID, episodeNumber, title, overview, still, runtime, airDate, videoID)
	if err != nil {
		return fmt.Errorf("upsert episode: %w", err)
	}
	return nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
