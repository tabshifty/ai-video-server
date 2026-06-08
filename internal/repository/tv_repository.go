package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"video-server/internal/models"
	"video-server/internal/utils"
)

type AdminTVSeriesFilter struct {
	Page        int
	PageSize    int
	Query       string
	Active      *bool
	HasPlayable *bool
}

func formatNullDate(value sql.NullTime) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format("2006-01-02")
}

func parseOptionalDate(raw string) (*time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, fmt.Errorf("invalid date %q", value)
	}
	return &parsed, nil
}

func boolPtr(value bool) *bool {
	return &value
}

func (r *VideoRepository) listTVSeriesSummaries(ctx context.Context, active *bool, q string, limit, offset int) ([]models.TvSeriesSummaryDto, int, error) {
	return r.listTVSeriesSummariesOrdered(
		ctx,
		active,
		q,
		limit,
		offset,
		"MAX(e.air_date) DESC NULLS LAST, s.title ASC",
	)
}

func (r *VideoRepository) listTVSeriesSummariesOrdered(
	ctx context.Context,
	active *bool,
	q string,
	limit,
	offset int,
	orderClause string,
) ([]models.TvSeriesSummaryDto, int, error) {
	where := []string{"1=1"}
	args := make([]any, 0, 8)
	next := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	if active != nil {
		where = append(where, "s.active = "+next(*active))
	}
	if keyword := strings.TrimSpace(strings.ToLower(q)); keyword != "" {
		where = append(where, "LOWER(COALESCE(s.title,'')) LIKE "+next("%"+keyword+"%"))
	}

	baseWhere := strings.Join(where, " AND ")
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM series s WHERE "+baseWhere, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count tv series: %w", err)
	}

	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx, `
SELECT
  s.id,
  COALESCE(s.title, ''),
  COALESCE(s.overview, ''),
  COALESCE(s.poster_path, ''),
  COALESCE(s.backdrop_path, ''),
  s.first_air_date,
  COALESCE(NULLIF(s.number_of_seasons, 0), COUNT(DISTINCT se.id)),
  COALESCE(NULLIF(s.number_of_episodes, 0), COUNT(e.id)),
  COUNT(e.id) FILTER (WHERE e.video_id IS NOT NULL AND COALESCE(v.status, '') = 'ready'),
  MAX(e.air_date)
FROM series s
LEFT JOIN seasons se ON se.series_id = s.id
LEFT JOIN episodes e ON e.season_id = se.id
LEFT JOIN videos v ON v.id = e.video_id
WHERE `+baseWhere+`
GROUP BY s.id, s.title, s.overview, s.poster_path, s.backdrop_path, s.first_air_date, s.number_of_seasons, s.number_of_episodes
ORDER BY `+orderClause+`
LIMIT $`+fmt.Sprintf("%d", len(args)-1)+` OFFSET $`+fmt.Sprintf("%d", len(args)), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list tv series: %w", err)
	}
	defer rows.Close()

	items := make([]models.TvSeriesSummaryDto, 0, limit)
	for rows.Next() {
		var item models.TvSeriesSummaryDto
		var firstAirDate sql.NullTime
		var latestAirDate sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Overview,
			&item.PosterURL,
			&item.BackdropURL,
			&firstAirDate,
			&item.TotalSeasons,
			&item.TotalEpisodes,
			&item.PlayableEpisodes,
			&latestAirDate,
		); err != nil {
			return nil, 0, fmt.Errorf("scan tv series summary: %w", err)
		}
		item.FirstAirDate = formatNullDate(firstAirDate)
		item.LatestEpisodeAirDate = formatNullDate(latestAirDate)
		item.PosterURL = resolveTVSeriesPosterURL(item.ID, item.PosterURL)
		item.BackdropURL = resolveTVSeriesBackdropURL(item.ID, item.BackdropURL)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate tv series summary: %w", err)
	}
	return items, total, nil
}

func (r *VideoRepository) ListActiveTVSeriesSummaries(ctx context.Context, limit, offset int) ([]models.TvSeriesSummaryDto, int, error) {
	return r.listTVSeriesSummaries(ctx, boolPtr(true), "", limit, offset)
}

func (r *VideoRepository) ListTVSeriesSummariesOrdered(ctx context.Context, orderClause string, limit, offset int) ([]models.TvSeriesSummaryDto, int, error) {
	return r.listTVSeriesSummariesOrdered(ctx, boolPtr(true), "", limit, offset, orderClause)
}

func (r *VideoRepository) ListBingeTVSeriesSummaries(ctx context.Context, limit, offset int) ([]models.TvSeriesSummaryDto, int, error) {
	return r.listTVSeriesSummariesOrdered(
		ctx,
		boolPtr(true),
		"",
		limit,
		offset,
		"COUNT(e.id) FILTER (WHERE e.video_id IS NOT NULL AND COALESCE(v.status, '') = 'ready') DESC, MAX(e.air_date) DESC NULLS LAST, s.title ASC",
	)
}

func (r *VideoRepository) ListClassicTVSeriesSummaries(ctx context.Context, limit, offset int) ([]models.TvSeriesSummaryDto, int, error) {
	return r.listTVSeriesSummariesOrdered(
		ctx,
		boolPtr(true),
		"",
		limit,
		offset,
		"s.first_air_date ASC NULLS LAST, s.title ASC",
	)
}

func (r *VideoRepository) SearchTVSeriesSummaries(ctx context.Context, q string, limit, offset int) ([]models.TvSeriesSummaryDto, int, error) {
	return r.listTVSeriesSummaries(ctx, boolPtr(true), q, limit, offset)
}

func (r *VideoRepository) GetTVContinueWatching(ctx context.Context, userID uuid.UUID) (*models.TvContinueWatchingDto, error) {
	var item models.TvContinueWatchingDto
	var durationSeconds sql.NullInt32
	err := r.pool.QueryRow(ctx, `
SELECT
  s.id,
  COALESCE(s.title, ''),
  se.season_number,
  e.episode_number,
  COALESCE(e.title, ''),
  COALESCE(s.poster_path, ''),
  COALESCE(s.backdrop_path, ''),
  a.watch_seconds,
  COALESCE(v.duration_seconds, e.runtime, 0)
FROM user_video_actions a
JOIN episodes e ON e.video_id = a.video_id
JOIN seasons se ON se.id = e.season_id
JOIN series s ON s.id = se.series_id
LEFT JOIN videos v ON v.id = e.video_id
WHERE a.user_id = $1
  AND a.action_type = 'view'
  AND a.watch_seconds > 0
  AND s.active = TRUE
ORDER BY a.updated_at DESC
LIMIT 1
`, userID).Scan(
		&item.SeriesID,
		&item.SeriesTitle,
		&item.SeasonNumber,
		&item.EpisodeNumber,
		&item.EpisodeTitle,
		&item.PosterURL,
		&item.BackdropURL,
		&item.WatchSeconds,
		&durationSeconds,
	)
	if err != nil {
		if IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get tv continue watching: %w", err)
	}
	item.DurationSeconds = nullInt32ToInt(durationSeconds)
	item.PosterURL = resolveTVSeriesPosterURL(item.SeriesID, item.PosterURL)
	item.BackdropURL = resolveTVSeriesBackdropURL(item.SeriesID, item.BackdropURL)
	if item.DurationSeconds > 0 {
		item.ProgressPercent = (item.WatchSeconds * 100) / item.DurationSeconds
		if item.ProgressPercent > 100 {
			item.ProgressPercent = 100
		}
	}
	return &item, nil
}

func (r *VideoRepository) GetTVSeriesDetail(ctx context.Context, userID uuid.UUID, seriesID int64) (models.TvSeriesDetailDto, error) {
	var detail models.TvSeriesDetailDto
	var firstAirDate sql.NullTime
	err := r.pool.QueryRow(ctx, `
SELECT
  s.id,
  COALESCE(s.title, ''),
  COALESCE(s.overview, ''),
  COALESCE(s.poster_path, ''),
  COALESCE(s.backdrop_path, ''),
  s.first_air_date,
  COALESCE(NULLIF(s.number_of_seasons, 0), (SELECT COUNT(*) FROM seasons se WHERE se.series_id = s.id)),
  COALESCE(NULLIF(s.number_of_episodes, 0), (
    SELECT COUNT(*)
    FROM episodes e
    JOIN seasons se ON se.id = e.season_id
    WHERE se.series_id = s.id
  )),
  (
    SELECT COUNT(*)
    FROM episodes e
    JOIN seasons se ON se.id = e.season_id
    LEFT JOIN videos v ON v.id = e.video_id
    WHERE se.series_id = s.id
      AND e.video_id IS NOT NULL
      AND COALESCE(v.status, '') = 'ready'
  )
FROM series s
WHERE s.id = $1 AND s.active = TRUE
`, seriesID).Scan(
		&detail.ID,
		&detail.Title,
		&detail.Overview,
		&detail.PosterURL,
		&detail.BackdropURL,
		&firstAirDate,
		&detail.TotalSeasons,
		&detail.TotalEpisodes,
		&detail.PlayableEpisodes,
	)
	if err != nil {
		return models.TvSeriesDetailDto{}, fmt.Errorf("get tv series detail: %w", err)
	}
	detail.FirstAirDate = formatNullDate(firstAirDate)
	detail.PosterURL = resolveTVSeriesPosterURL(detail.ID, detail.PosterURL)
	detail.BackdropURL = resolveTVSeriesBackdropURL(detail.ID, detail.BackdropURL)

	tagRows, err := r.pool.Query(ctx, `
SELECT DISTINCT vt.tag
FROM video_tags vt
JOIN episodes e ON e.video_id = vt.video_id
JOIN seasons se ON se.id = e.season_id
WHERE se.series_id = $1
ORDER BY vt.tag
LIMIT 12
`, seriesID)
	if err != nil {
		return models.TvSeriesDetailDto{}, fmt.Errorf("query tv tags: %w", err)
	}
	for tagRows.Next() {
		var tag string
		if err := tagRows.Scan(&tag); err != nil {
			tagRows.Close()
			return models.TvSeriesDetailDto{}, fmt.Errorf("scan tv tag: %w", err)
		}
		detail.Tags = append(detail.Tags, tag)
	}
	tagRows.Close()

	castRows, err := r.pool.Query(ctx, `
SELECT DISTINCT a.name
FROM actors a
JOIN video_actors va ON va.actor_id = a.id
JOIN episodes e ON e.video_id = va.video_id
JOIN seasons se ON se.id = e.season_id
WHERE se.series_id = $1
ORDER BY a.name
LIMIT 12
`, seriesID)
	if err != nil {
		return models.TvSeriesDetailDto{}, fmt.Errorf("query tv cast: %w", err)
	}
	for castRows.Next() {
		var name string
		if err := castRows.Scan(&name); err != nil {
			castRows.Close()
			return models.TvSeriesDetailDto{}, fmt.Errorf("scan tv cast: %w", err)
		}
		detail.Cast = append(detail.Cast, name)
	}
	castRows.Close()

	seasonRows, err := r.pool.Query(ctx, `
SELECT id, season_number, COALESCE(name, ''), COALESCE(overview, ''), COALESCE(poster_path, ''), air_date
FROM seasons
WHERE series_id = $1
ORDER BY season_number ASC
`, seriesID)
	if err != nil {
		return models.TvSeriesDetailDto{}, fmt.Errorf("query tv seasons: %w", err)
	}
	seasonMap := make(map[int64]int)
	for seasonRows.Next() {
		var season models.TvSeasonDto
		var airDate sql.NullTime
		if err := seasonRows.Scan(&season.ID, &season.SeasonNumber, &season.Title, &season.Overview, &season.PosterURL, &airDate); err != nil {
			seasonRows.Close()
			return models.TvSeriesDetailDto{}, fmt.Errorf("scan tv season: %w", err)
		}
		season.AirDate = formatNullDate(airDate)
		seasonMap[season.ID] = len(detail.Seasons)
		detail.Seasons = append(detail.Seasons, season)
	}
	seasonRows.Close()

	episodeRows, err := r.pool.Query(ctx, `
SELECT
  e.id,
  e.season_id,
  se.season_number,
  e.episode_number,
  COALESCE(e.title, ''),
  COALESCE(e.overview, ''),
  COALESCE(e.runtime, v.duration_seconds, 0),
  e.air_date,
  COALESCE(e.still_path, ''),
  COALESCE(e.video_id::text, ''),
  COALESCE(v.title, ''),
  COALESCE(v.status, ''),
  COALESCE(a.watch_seconds, 0),
  a.updated_at,
  COALESCE(v.metadata, '{}'::jsonb)
FROM episodes e
LEFT JOIN videos v ON v.id = e.video_id
LEFT JOIN user_video_actions a
  ON a.video_id = e.video_id
 AND a.user_id = $2
 AND a.action_type = 'view'
JOIN seasons se ON se.id = e.season_id
WHERE se.series_id = $1
ORDER BY se.season_number ASC, e.episode_number ASC
`, seriesID, userID)
	if err != nil {
		return models.TvSeriesDetailDto{}, fmt.Errorf("query tv episodes: %w", err)
	}
	for episodeRows.Next() {
		var episode models.TvEpisodeDto
		var seasonID int64
		var seasonNumber int
		var airDate sql.NullTime
		var lastWatchedAt sql.NullTime
		if err := episodeRows.Scan(
			&episode.ID,
			&seasonID,
			&seasonNumber,
			&episode.EpisodeNumber,
			&episode.Title,
			&episode.Overview,
			&episode.Runtime,
			&airDate,
			&episode.StillURL,
			&episode.VideoID,
			&episode.VideoTitle,
			&episode.VideoStatus,
			&episode.WatchSeconds,
			&lastWatchedAt,
			&episode.Metadata,
		); err != nil {
			episodeRows.Close()
			return models.TvSeriesDetailDto{}, fmt.Errorf("scan tv episode: %w", err)
		}
		episode.StillURL = resolveTVEpisodeStillURL(detail.ID, seasonNumber, episode.EpisodeNumber, episode.StillURL)
		episode.AirDate = formatNullDate(airDate)
		if lastWatchedAt.Valid {
			episode.LastWatchedAt = lastWatchedAt.Time.Format(time.RFC3339)
		}
		episode.Playable = episode.VideoID != "" && episode.VideoStatus == "ready"
		if episode.Runtime > 0 {
			episode.ProgressPercent = (episode.WatchSeconds * 100) / episode.Runtime
			if episode.ProgressPercent > 100 {
				episode.ProgressPercent = 100
			}
		}
		index, ok := seasonMap[seasonID]
		if !ok {
			continue
		}
		detail.Seasons[index].Episodes = append(detail.Seasons[index].Episodes, episode)
	}
	episodeRows.Close()

	videoIDs := make([]uuid.UUID, 0, len(detail.Seasons)*4)
	for _, season := range detail.Seasons {
		for _, episode := range season.Episodes {
			if strings.TrimSpace(episode.VideoID) == "" {
				continue
			}
			videoID, err := uuid.Parse(episode.VideoID)
			if err != nil {
				continue
			}
			videoIDs = append(videoIDs, videoID)
		}
	}
	subtitlesByVideoID, err := r.ListVideoSubtitlesByVideoIDs(ctx, videoIDs)
	if err != nil {
		return models.TvSeriesDetailDto{}, fmt.Errorf("query tv subtitles: %w", err)
	}
	for seasonIndex := range detail.Seasons {
		for episodeIndex := range detail.Seasons[seasonIndex].Episodes {
			videoIDRaw := strings.TrimSpace(detail.Seasons[seasonIndex].Episodes[episodeIndex].VideoID)
			if videoIDRaw == "" {
				continue
			}
			videoID, err := uuid.Parse(videoIDRaw)
			if err != nil {
				continue
			}
			subtitles := subtitlesByVideoID[videoID]
			detail.Seasons[seasonIndex].Episodes[episodeIndex].SubtitleTracks = make([]models.SubtitleTrack, 0, len(subtitles))
			for _, subtitle := range subtitles {
				detail.Seasons[seasonIndex].Episodes[episodeIndex].SubtitleTracks = append(detail.Seasons[seasonIndex].Episodes[episodeIndex].SubtitleTracks, BuildAppSubtitleTrack(subtitle))
			}
		}
	}
	return detail, nil
}

func (r *VideoRepository) GetTVSeriesArtworkPaths(ctx context.Context, seriesID int64) (posterPath, backdropPath string, err error) {
	err = r.pool.QueryRow(ctx, `
SELECT COALESCE(poster_path, ''), COALESCE(backdrop_path, '')
FROM series
WHERE id = $1
`, seriesID).Scan(&posterPath, &backdropPath)
	if err != nil {
		return "", "", fmt.Errorf("get tv series artwork: %w", err)
	}
	return posterPath, backdropPath, nil
}

func (r *VideoRepository) GetTVEpisodeStillPath(ctx context.Context, seriesID int64, seasonNumber, episodeNumber int) (string, error) {
	var stillPath string
	err := r.pool.QueryRow(ctx, `
SELECT COALESCE(e.still_path, '')
FROM episodes e
JOIN seasons se ON se.id = e.season_id
WHERE se.series_id = $1
  AND se.season_number = $2
  AND e.episode_number = $3
`, seriesID, seasonNumber, episodeNumber).Scan(&stillPath)
	if err != nil {
		return "", fmt.Errorf("get tv episode still: %w", err)
	}
	return stillPath, nil
}

func resolveTVSeriesPosterURL(seriesID int64, raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	return utils.TVSeriesPosterURL(seriesID)
}

func resolveTVSeriesBackdropURL(seriesID int64, raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	return utils.TVSeriesBackdropURL(seriesID)
}

func resolveTVEpisodeStillURL(seriesID int64, seasonNumber, episodeNumber int, raw string) string {
	if strings.TrimSpace(raw) == "" || seriesID <= 0 || seasonNumber <= 0 || episodeNumber <= 0 {
		return ""
	}
	return utils.TVEpisodeStillURL(seriesID, seasonNumber, episodeNumber)
}

func (r *VideoRepository) ListAdminTVSeries(ctx context.Context, filter AdminTVSeriesFilter) ([]models.AdminTvSeriesListItem, int, error) {
	where := []string{"1=1"}
	args := make([]any, 0, 8)
	next := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}
	if keyword := strings.TrimSpace(strings.ToLower(filter.Query)); keyword != "" {
		where = append(where, "LOWER(COALESCE(s.title,'')) LIKE "+next("%"+keyword+"%"))
	}
	if filter.Active != nil {
		where = append(where, "s.active = "+next(*filter.Active))
	}
	if filter.HasPlayable != nil {
		existsSQL := `EXISTS (
  SELECT 1
  FROM episodes e
  JOIN seasons se ON se.id = e.season_id
  LEFT JOIN videos v ON v.id = e.video_id
  WHERE se.series_id = s.id
    AND e.video_id IS NOT NULL
    AND COALESCE(v.status, '') = 'ready'
)`
		if *filter.HasPlayable {
			where = append(where, existsSQL)
		} else {
			where = append(where, "NOT "+existsSQL)
		}
	}
	baseWhere := strings.Join(where, " AND ")

	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM series s WHERE "+baseWhere, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count admin tv series: %w", err)
	}

	args = append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)
	rows, err := r.pool.Query(ctx, `
SELECT
  s.id,
  COALESCE(s.title, ''),
  COALESCE(s.overview, ''),
  COALESCE(s.poster_path, ''),
  COALESCE(s.backdrop_path, ''),
  s.first_air_date,
  COALESCE(NULLIF(s.number_of_seasons, 0), COUNT(DISTINCT se.id)),
  COALESCE(NULLIF(s.number_of_episodes, 0), COUNT(e.id)),
  COUNT(e.id) FILTER (WHERE e.video_id IS NOT NULL AND COALESCE(v.status, '') = 'ready'),
  s.active
FROM series s
LEFT JOIN seasons se ON se.series_id = s.id
LEFT JOIN episodes e ON e.season_id = se.id
LEFT JOIN videos v ON v.id = e.video_id
WHERE `+baseWhere+`
GROUP BY s.id, s.title, s.overview, s.poster_path, s.backdrop_path, s.first_air_date, s.number_of_seasons, s.number_of_episodes, s.active
ORDER BY s.active DESC, MAX(e.air_date) DESC NULLS LAST, s.title ASC
LIMIT $`+fmt.Sprintf("%d", len(args)-1)+` OFFSET $`+fmt.Sprintf("%d", len(args)), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list admin tv series: %w", err)
	}
	defer rows.Close()

	items := make([]models.AdminTvSeriesListItem, 0, filter.PageSize)
	for rows.Next() {
		var item models.AdminTvSeriesListItem
		var firstAirDate sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Overview,
			&item.PosterURL,
			&item.BackdropURL,
			&firstAirDate,
			&item.TotalSeasons,
			&item.TotalEpisodes,
			&item.PlayableEpisodes,
			&item.Active,
		); err != nil {
			return nil, 0, fmt.Errorf("scan admin tv series: %w", err)
		}
		item.FirstAirDate = formatNullDate(firstAirDate)
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *VideoRepository) GetAdminTVSeriesDetail(ctx context.Context, seriesID int64) (models.AdminTvSeriesDetail, error) {
	var detail models.AdminTvSeriesDetail
	var firstAirDate sql.NullTime
	err := r.pool.QueryRow(ctx, `
SELECT
  s.id,
  COALESCE(s.title, ''),
  COALESCE(s.overview, ''),
  COALESCE(s.poster_path, ''),
  COALESCE(s.backdrop_path, ''),
  s.first_air_date,
  COALESCE(NULLIF(s.number_of_seasons, 0), (SELECT COUNT(*) FROM seasons se WHERE se.series_id = s.id)),
  COALESCE(NULLIF(s.number_of_episodes, 0), (
    SELECT COUNT(*)
    FROM episodes e
    JOIN seasons se ON se.id = e.season_id
    WHERE se.series_id = s.id
  )),
  (
    SELECT COUNT(*)
    FROM episodes e
    JOIN seasons se ON se.id = e.season_id
    LEFT JOIN videos v ON v.id = e.video_id
    WHERE se.series_id = s.id
      AND e.video_id IS NOT NULL
      AND COALESCE(v.status, '') = 'ready'
  ),
  s.active
FROM series s
WHERE s.id = $1
`, seriesID).Scan(
		&detail.ID,
		&detail.Title,
		&detail.Overview,
		&detail.PosterURL,
		&detail.BackdropURL,
		&firstAirDate,
		&detail.TotalSeasons,
		&detail.TotalEpisodes,
		&detail.PlayableEpisodes,
		&detail.Active,
	)
	if err != nil {
		return models.AdminTvSeriesDetail{}, fmt.Errorf("get admin tv detail: %w", err)
	}
	detail.FirstAirDate = formatNullDate(firstAirDate)

	seasonRows, err := r.pool.Query(ctx, `
SELECT id, season_number, COALESCE(name, ''), COALESCE(overview, ''), COALESCE(poster_path, ''), air_date
FROM seasons
WHERE series_id = $1
ORDER BY season_number ASC
`, seriesID)
	if err != nil {
		return models.AdminTvSeriesDetail{}, fmt.Errorf("query admin tv seasons: %w", err)
	}
	seasonMap := make(map[int64]int)
	for seasonRows.Next() {
		var season models.AdminTvSeasonDetail
		var airDate sql.NullTime
		if err := seasonRows.Scan(&season.ID, &season.SeasonNumber, &season.Title, &season.Overview, &season.PosterURL, &airDate); err != nil {
			seasonRows.Close()
			return models.AdminTvSeriesDetail{}, fmt.Errorf("scan admin tv season: %w", err)
		}
		season.SeriesID = seriesID
		season.AirDate = formatNullDate(airDate)
		seasonMap[season.ID] = len(detail.Seasons)
		detail.Seasons = append(detail.Seasons, season)
	}
	seasonRows.Close()

	episodeRows, err := r.pool.Query(ctx, `
SELECT
  e.id,
  e.season_id,
  e.episode_number,
  COALESCE(e.title, ''),
  COALESCE(e.overview, ''),
  COALESCE(e.runtime, v.duration_seconds, 0),
  e.air_date,
  COALESCE(e.still_path, ''),
  COALESCE(e.video_id::text, ''),
  COALESCE(v.title, ''),
  COALESCE(v.status, '')
FROM episodes e
LEFT JOIN videos v ON v.id = e.video_id
JOIN seasons se ON se.id = e.season_id
WHERE se.series_id = $1
ORDER BY se.season_number ASC, e.episode_number ASC
`, seriesID)
	if err != nil {
		return models.AdminTvSeriesDetail{}, fmt.Errorf("query admin tv episodes: %w", err)
	}
	for episodeRows.Next() {
		var episode models.AdminTvEpisodeDetail
		var seasonID int64
		var airDate sql.NullTime
		if err := episodeRows.Scan(
			&episode.ID,
			&seasonID,
			&episode.EpisodeNumber,
			&episode.Title,
			&episode.Overview,
			&episode.Runtime,
			&airDate,
			&episode.StillURL,
			&episode.VideoID,
			&episode.VideoTitle,
			&episode.VideoStatus,
		); err != nil {
			episodeRows.Close()
			return models.AdminTvSeriesDetail{}, fmt.Errorf("scan admin tv episode: %w", err)
		}
		episode.SeasonID = seasonID
		episode.AirDate = formatNullDate(airDate)
		episode.Playable = episode.VideoID != "" && episode.VideoStatus == "ready"
		index, ok := seasonMap[seasonID]
		if ok {
			detail.Seasons[index].Episodes = append(detail.Seasons[index].Episodes, episode)
		}
	}
	episodeRows.Close()

	return detail, nil
}

func (r *VideoRepository) validateEpisodeVideoBinding(ctx context.Context, raw string) (*uuid.UUID, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, nil
	}
	videoID, err := uuid.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("invalid video_id")
	}
	var videoType string
	if err := r.pool.QueryRow(ctx, `SELECT type FROM videos WHERE id = $1`, videoID).Scan(&videoType); err != nil {
		if IsNotFound(err) {
			return nil, fmt.Errorf("video not found")
		}
		return nil, fmt.Errorf("query video binding: %w", err)
	}
	if videoType != "episode" {
		return nil, fmt.Errorf("只能绑定 type=episode 的视频")
	}
	return &videoID, nil
}

func (r *VideoRepository) syncSeriesCounts(ctx context.Context, seriesID int64) error {
	_, err := r.pool.Exec(ctx, `
UPDATE series
SET
  number_of_seasons = COALESCE(src.season_count, 0),
  number_of_episodes = COALESCE(src.episode_count, 0)
FROM (
  SELECT
    s.id,
    COUNT(DISTINCT se.id) AS season_count,
    COUNT(e.id) AS episode_count
  FROM series s
  LEFT JOIN seasons se ON se.series_id = s.id
  LEFT JOIN episodes e ON e.season_id = se.id
  WHERE s.id = $1
  GROUP BY s.id
) AS src
WHERE series.id = src.id
`, seriesID)
	if err != nil {
		return fmt.Errorf("sync series counts: %w", err)
	}
	return nil
}

func (r *VideoRepository) CreateAdminTVSeries(ctx context.Context, input models.AdminTvSeriesInput) (models.AdminTvSeriesDetail, error) {
	firstAirDate, err := parseOptionalDate(input.FirstAirDate)
	if err != nil {
		return models.AdminTvSeriesDetail{}, err
	}
	var seriesID int64
	err = r.pool.QueryRow(ctx, `
INSERT INTO series(title, overview, poster_path, backdrop_path, first_air_date, number_of_seasons, number_of_episodes, active)
VALUES ($1, $2, $3, $4, $5, 0, 0, $6)
RETURNING id
`, strings.TrimSpace(input.Title), strings.TrimSpace(input.Overview), strings.TrimSpace(input.PosterURL), strings.TrimSpace(input.BackdropURL), firstAirDate, input.Active).Scan(&seriesID)
	if err != nil {
		return models.AdminTvSeriesDetail{}, fmt.Errorf("create admin tv series: %w", err)
	}
	return r.GetAdminTVSeriesDetail(ctx, seriesID)
}

func (r *VideoRepository) UpdateAdminTVSeries(ctx context.Context, seriesID int64, input models.AdminTvSeriesInput) (models.AdminTvSeriesDetail, error) {
	firstAirDate, err := parseOptionalDate(input.FirstAirDate)
	if err != nil {
		return models.AdminTvSeriesDetail{}, err
	}
	tag, err := r.pool.Exec(ctx, `
UPDATE series
SET title = $2,
    overview = $3,
    poster_path = $4,
    backdrop_path = $5,
    first_air_date = $6,
    active = $7
WHERE id = $1
`, seriesID, strings.TrimSpace(input.Title), strings.TrimSpace(input.Overview), strings.TrimSpace(input.PosterURL), strings.TrimSpace(input.BackdropURL), firstAirDate, input.Active)
	if err != nil {
		return models.AdminTvSeriesDetail{}, fmt.Errorf("update admin tv series: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.AdminTvSeriesDetail{}, pgx.ErrNoRows
	}
	return r.GetAdminTVSeriesDetail(ctx, seriesID)
}

func (r *VideoRepository) DeleteAdminTVSeries(ctx context.Context, seriesID int64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM series WHERE id = $1`, seriesID)
	if err != nil {
		return fmt.Errorf("delete admin tv series: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *VideoRepository) CreateAdminTVSeason(ctx context.Context, seriesID int64, input models.AdminTvSeasonInput) (models.AdminTvSeasonDetail, error) {
	airDate, err := parseOptionalDate(input.AirDate)
	if err != nil {
		return models.AdminTvSeasonDetail{}, err
	}
	var seasonID int64
	err = r.pool.QueryRow(ctx, `
INSERT INTO seasons(series_id, season_number, name, overview, poster_path, air_date)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id
`, seriesID, input.SeasonNumber, strings.TrimSpace(input.Title), strings.TrimSpace(input.Overview), strings.TrimSpace(input.PosterURL), airDate).Scan(&seasonID)
	if err != nil {
		return models.AdminTvSeasonDetail{}, fmt.Errorf("create admin tv season: %w", err)
	}
	if err := r.syncSeriesCounts(ctx, seriesID); err != nil {
		return models.AdminTvSeasonDetail{}, err
	}
	detail, err := r.GetAdminTVSeriesDetail(ctx, seriesID)
	if err != nil {
		return models.AdminTvSeasonDetail{}, err
	}
	for _, season := range detail.Seasons {
		if season.ID == seasonID {
			return season, nil
		}
	}
	return models.AdminTvSeasonDetail{}, pgx.ErrNoRows
}

func (r *VideoRepository) UpdateAdminTVSeason(ctx context.Context, seasonID int64, input models.AdminTvSeasonInput) (models.AdminTvSeasonDetail, error) {
	airDate, err := parseOptionalDate(input.AirDate)
	if err != nil {
		return models.AdminTvSeasonDetail{}, err
	}
	var seriesID int64
	err = r.pool.QueryRow(ctx, `
UPDATE seasons
SET season_number = $2,
    name = $3,
    overview = $4,
    poster_path = $5,
    air_date = $6
WHERE id = $1
RETURNING series_id
`, seasonID, input.SeasonNumber, strings.TrimSpace(input.Title), strings.TrimSpace(input.Overview), strings.TrimSpace(input.PosterURL), airDate).Scan(&seriesID)
	if err != nil {
		return models.AdminTvSeasonDetail{}, fmt.Errorf("update admin tv season: %w", err)
	}
	if err := r.syncSeriesCounts(ctx, seriesID); err != nil {
		return models.AdminTvSeasonDetail{}, err
	}
	detail, err := r.GetAdminTVSeriesDetail(ctx, seriesID)
	if err != nil {
		return models.AdminTvSeasonDetail{}, err
	}
	for _, season := range detail.Seasons {
		if season.ID == seasonID {
			return season, nil
		}
	}
	return models.AdminTvSeasonDetail{}, pgx.ErrNoRows
}

func (r *VideoRepository) DeleteAdminTVSeason(ctx context.Context, seasonID int64) error {
	var seriesID int64
	if err := r.pool.QueryRow(ctx, `SELECT series_id FROM seasons WHERE id = $1`, seasonID).Scan(&seriesID); err != nil {
		return fmt.Errorf("query season before delete: %w", err)
	}
	if _, err := r.pool.Exec(ctx, `DELETE FROM seasons WHERE id = $1`, seasonID); err != nil {
		return fmt.Errorf("delete admin tv season: %w", err)
	}
	return r.syncSeriesCounts(ctx, seriesID)
}

func (r *VideoRepository) CreateAdminTVEpisode(ctx context.Context, seasonID int64, input models.AdminTvEpisodeInput) (models.AdminTvEpisodeDetail, error) {
	videoID, err := r.validateEpisodeVideoBinding(ctx, input.VideoID)
	if err != nil {
		return models.AdminTvEpisodeDetail{}, err
	}
	airDate, err := parseOptionalDate(input.AirDate)
	if err != nil {
		return models.AdminTvEpisodeDetail{}, err
	}
	var episodeID int64
	err = r.pool.QueryRow(ctx, `
INSERT INTO episodes(season_id, episode_number, title, overview, still_path, runtime, air_date, video_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id
`, seasonID, input.EpisodeNumber, strings.TrimSpace(input.Title), strings.TrimSpace(input.Overview), strings.TrimSpace(input.StillURL), input.Runtime, airDate, videoID).Scan(&episodeID)
	if err != nil {
		return models.AdminTvEpisodeDetail{}, fmt.Errorf("create admin tv episode: %w", err)
	}
	var seriesID int64
	if err := r.pool.QueryRow(ctx, `SELECT series_id FROM seasons WHERE id = $1`, seasonID).Scan(&seriesID); err != nil {
		return models.AdminTvEpisodeDetail{}, fmt.Errorf("query series for episode: %w", err)
	}
	if err := r.syncSeriesCounts(ctx, seriesID); err != nil {
		return models.AdminTvEpisodeDetail{}, err
	}
	detail, err := r.GetAdminTVSeriesDetail(ctx, seriesID)
	if err != nil {
		return models.AdminTvEpisodeDetail{}, err
	}
	for _, season := range detail.Seasons {
		for _, episode := range season.Episodes {
			if episode.ID == episodeID {
				return episode, nil
			}
		}
	}
	return models.AdminTvEpisodeDetail{}, pgx.ErrNoRows
}

func (r *VideoRepository) UpdateAdminTVEpisode(ctx context.Context, episodeID int64, input models.AdminTvEpisodeInput) (models.AdminTvEpisodeDetail, error) {
	videoID, err := r.validateEpisodeVideoBinding(ctx, input.VideoID)
	if err != nil {
		return models.AdminTvEpisodeDetail{}, err
	}
	airDate, err := parseOptionalDate(input.AirDate)
	if err != nil {
		return models.AdminTvEpisodeDetail{}, err
	}
	var seasonID int64
	err = r.pool.QueryRow(ctx, `
UPDATE episodes
SET episode_number = $2,
    title = $3,
    overview = $4,
    still_path = $5,
    runtime = $6,
    air_date = $7,
    video_id = $8
WHERE id = $1
RETURNING season_id
`, episodeID, input.EpisodeNumber, strings.TrimSpace(input.Title), strings.TrimSpace(input.Overview), strings.TrimSpace(input.StillURL), input.Runtime, airDate, videoID).Scan(&seasonID)
	if err != nil {
		return models.AdminTvEpisodeDetail{}, fmt.Errorf("update admin tv episode: %w", err)
	}
	var seriesID int64
	if err := r.pool.QueryRow(ctx, `SELECT series_id FROM seasons WHERE id = $1`, seasonID).Scan(&seriesID); err != nil {
		return models.AdminTvEpisodeDetail{}, fmt.Errorf("query series for updated episode: %w", err)
	}
	if err := r.syncSeriesCounts(ctx, seriesID); err != nil {
		return models.AdminTvEpisodeDetail{}, err
	}
	detail, err := r.GetAdminTVSeriesDetail(ctx, seriesID)
	if err != nil {
		return models.AdminTvEpisodeDetail{}, err
	}
	for _, season := range detail.Seasons {
		for _, episode := range season.Episodes {
			if episode.ID == episodeID {
				return episode, nil
			}
		}
	}
	return models.AdminTvEpisodeDetail{}, pgx.ErrNoRows
}

func (r *VideoRepository) DeleteAdminTVEpisode(ctx context.Context, episodeID int64) error {
	var seasonID int64
	if err := r.pool.QueryRow(ctx, `SELECT season_id FROM episodes WHERE id = $1`, episodeID).Scan(&seasonID); err != nil {
		return fmt.Errorf("query episode before delete: %w", err)
	}
	var seriesID int64
	if err := r.pool.QueryRow(ctx, `SELECT series_id FROM seasons WHERE id = $1`, seasonID).Scan(&seriesID); err != nil {
		return fmt.Errorf("query series before delete episode: %w", err)
	}
	if _, err := r.pool.Exec(ctx, `DELETE FROM episodes WHERE id = $1`, episodeID); err != nil {
		return fmt.Errorf("delete admin tv episode: %w", err)
	}
	return r.syncSeriesCounts(ctx, seriesID)
}

func NormalizeTVSeriesSearchResults(items []models.TvSeriesSummaryDto) []models.TvSeriesSummaryDto {
	out := append([]models.TvSeriesSummaryDto(nil), items...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].PlayableEpisodes == out[j].PlayableEpisodes {
			if out[i].LatestEpisodeAirDate == out[j].LatestEpisodeAirDate {
				return out[i].Title < out[j].Title
			}
			return out[i].LatestEpisodeAirDate > out[j].LatestEpisodeAirDate
		}
		return out[i].PlayableEpisodes > out[j].PlayableEpisodes
	})
	return out
}
