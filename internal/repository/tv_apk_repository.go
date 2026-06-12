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
	"github.com/jackc/pgx/v5"

	"video-server/internal/models"
)

type tvRowScanner interface {
	Scan(dest ...any) error
}

func (r *VideoRepository) AdminListTVAppReleases(ctx context.Context, f models.AdminTvAppReleaseFilter) ([]models.AdminTvAppReleaseListItem, int, error) {
	where := []string{"1=1"}
	args := make([]any, 0, 10)
	next := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	if keyword := strings.ToLower(strings.TrimSpace(f.Keyword)); keyword != "" {
		like := "%" + keyword + "%"
		p1 := next(like)
		p2 := next(parseVersionKeyword(keyword))
		where = append(where, "(LOWER(r.version_name) LIKE "+p1+" OR r.version_code::text LIKE "+p2+" OR to_char(r.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS') LIKE "+p1+" OR to_char(COALESCE(r.published_at, r.created_at) AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS') LIKE "+p1+")")
	}
	if status := strings.TrimSpace(f.Status); status != "" {
		where = append(where, "r.publish_status = "+next(status))
	}
	switch strings.TrimSpace(f.ABICompleteness) {
	case "complete":
		where = append(where, "(SELECT COUNT(*) FROM tv_app_release_apks a WHERE a.release_id = r.id) >= 2")
	case "missing":
		where = append(where, "(SELECT COUNT(*) FROM tv_app_release_apks a WHERE a.release_id = r.id) BETWEEN 1 AND 1")
	case "empty":
		where = append(where, "(SELECT COUNT(*) FROM tv_app_release_apks a WHERE a.release_id = r.id) = 0")
	}
	if f.CurrentPublished {
		where = append(where, "r.publish_status IN ('published_complete', 'published_missing_abi')")
	}

	baseWhere := strings.Join(where, " AND ")
	var total int
	countSQL := "SELECT COUNT(*) FROM tv_app_releases r WHERE " + baseWhere
	if err := r.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count tv app releases: %w", err)
	}

	args = append(args, f.PageSize, (f.Page-1)*f.PageSize)
	listSQL := `
SELECT
  r.id,
  r.package_name,
  r.version_code,
  r.version_name,
  COALESCE(r.release_notes, ''),
  COALESCE(r.remarks, ''),
  r.publish_status,
  r.published_at,
  r.last_status_changed_at,
  r.created_at,
  r.updated_at,
  r.latest_recommended
FROM tv_app_releases r
WHERE ` + baseWhere + `
ORDER BY
  CASE WHEN r.publish_status IN ('published_complete', 'published_missing_abi') THEN 0 ELSE 1 END,
  r.version_code DESC,
  r.created_at DESC
LIMIT $` + fmt.Sprintf("%d", len(args)-1) + ` OFFSET $` + fmt.Sprintf("%d", len(args))

	rows, err := r.pool.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list tv app releases: %w", err)
	}
	defer rows.Close()

	releases := make([]models.TVAppReleaseRecord, 0, f.PageSize)
	for rows.Next() {
		item, scanErr := scanTVAppReleaseRecord(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		releases = append(releases, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if err := r.attachTVAppReleaseABIs(ctx, releases); err != nil {
		return nil, 0, err
	}

	items := make([]models.AdminTvAppReleaseListItem, 0, len(releases))
	for _, release := range releases {
		items = append(items, buildAdminTVAppReleaseListItem(release))
	}
	return items, total, nil
}

func (r *VideoRepository) GetAdminTVAppReleaseDetail(ctx context.Context, releaseID int64) (models.AdminTvAppReleaseDetail, error) {
	row := r.pool.QueryRow(ctx, `
SELECT
  id,
  package_name,
  version_code,
  version_name,
  COALESCE(release_notes, ''),
  COALESCE(remarks, ''),
  publish_status,
  published_at,
  last_status_changed_at,
  created_at,
  updated_at,
  latest_recommended
FROM tv_app_releases
WHERE id = $1
`, releaseID)
	release, err := scanTVAppReleaseRecord(row)
	if err != nil {
		return models.AdminTvAppReleaseDetail{}, err
	}
	releases := []models.TVAppReleaseRecord{release}
	if err := r.attachTVAppReleaseABIs(ctx, releases); err != nil {
		return models.AdminTvAppReleaseDetail{}, err
	}
	return buildAdminTVAppReleaseListItem(releases[0]), nil
}

func (r *VideoRepository) CreateOrUpdateTVAppDraftReleaseWithABI(
	ctx context.Context,
	meta models.TVAppAPKParsedMetadata,
	storedPath string,
	userID *uuid.UUID,
	username string,
	replaceExisting bool,
) (models.TVAppReleaseRecord, models.TVAppReleaseABIInfo, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return models.TVAppReleaseRecord{}, models.TVAppReleaseABIInfo{}, fmt.Errorf("begin tv app draft upsert: %w", err)
	}
	defer tx.Rollback(ctx)

	release, err := upsertTVAppReleaseRecord(ctx, tx, meta.PackageName, meta.VersionCode, meta.VersionName)
	if err != nil {
		return models.TVAppReleaseRecord{}, models.TVAppReleaseABIInfo{}, err
	}

	existingABI, hasExisting, err := getTVAppReleaseABIByABI(ctx, tx, release.ID, meta.ABI)
	if err != nil {
		return models.TVAppReleaseRecord{}, models.TVAppReleaseABIInfo{}, err
	}
	if hasExisting && !replaceExisting {
		return models.TVAppReleaseRecord{}, models.TVAppReleaseABIInfo{}, models.NewTVAPKDomainError(models.TVAPKErrorABIAlreadyExists, "该版本的同 ABI 安装包已存在，请使用替换动作")
	}
	if hasExisting && release.PublishStatus != models.TVReleaseStatusDraft && release.PublishStatus != models.TVReleaseStatusOffline {
		return models.TVAppReleaseRecord{}, models.TVAppReleaseABIInfo{}, models.NewTVAPKDomainError(models.TVAPKErrorReplaceNeedsOffline, "已发布记录替换 ABI 前必须先下线")
	}

	abiInfo, err := saveTVAppReleaseABI(ctx, tx, release, existingABI, hasExisting, meta, storedPath, userID, username)
	if err != nil {
		return models.TVAppReleaseRecord{}, models.TVAppReleaseABIInfo{}, err
	}

	releaseWithABIs, err := loadTVAppReleaseRecordTx(ctx, tx, release.ID)
	if err != nil {
		return models.TVAppReleaseRecord{}, models.TVAppReleaseABIInfo{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return models.TVAppReleaseRecord{}, models.TVAppReleaseABIInfo{}, fmt.Errorf("commit tv app draft upsert: %w", err)
	}
	return releaseWithABIs, abiInfo, nil
}

func (r *VideoRepository) UpdateTVAppReleaseNotes(ctx context.Context, releaseID int64, releaseNotes, remarks string) (models.TVAppReleaseRecord, error) {
	_, err := r.pool.Exec(ctx, `
UPDATE tv_app_releases
SET release_notes = $2,
    remarks = $3,
    updated_at = NOW()
WHERE id = $1
`, releaseID, strings.TrimSpace(releaseNotes), strings.TrimSpace(remarks))
	if err != nil {
		return models.TVAppReleaseRecord{}, fmt.Errorf("update tv app release notes: %w", err)
	}
	return r.GetTVAppReleaseRecord(ctx, releaseID)
}

func (r *VideoRepository) PublishTVAppRelease(ctx context.Context, releaseID int64, releaseNotes, remarks string) (models.TVAppReleaseRecord, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return models.TVAppReleaseRecord{}, fmt.Errorf("begin publish tv app release: %w", err)
	}
	defer tx.Rollback(ctx)

	release, err := loadTVAppReleaseRecordTx(ctx, tx, releaseID)
	if err != nil {
		return models.TVAppReleaseRecord{}, err
	}
	if len(release.ABIItems) == 0 {
		return models.TVAppReleaseRecord{}, models.NewTVAPKDomainError(models.TVAPKErrorReleaseNotPublishable, "至少上传一个 ABI 安装包后才能发布")
	}

	now := time.Now().UTC()
	publishedAt := release.PublishedAt
	if publishedAt == nil {
		publishedAt = &now
	}
	status := models.TVReleaseStatusForVisibility(true, collectReleaseABIStrings(release.ABIItems))
	_, err = tx.Exec(ctx, `
UPDATE tv_app_releases
SET release_notes = $2,
    remarks = $3,
    publish_status = $4,
    published_at = $5,
    last_status_changed_at = NOW(),
    updated_at = NOW()
WHERE id = $1
`, releaseID, strings.TrimSpace(releaseNotes), strings.TrimSpace(remarks), status, publishedAt)
	if err != nil {
		return models.TVAppReleaseRecord{}, fmt.Errorf("publish tv app release: %w", err)
	}
	if err := recalcTVAppRecommendationsTx(ctx, tx); err != nil {
		return models.TVAppReleaseRecord{}, err
	}
	release, err = loadTVAppReleaseRecordTx(ctx, tx, releaseID)
	if err != nil {
		return models.TVAppReleaseRecord{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return models.TVAppReleaseRecord{}, fmt.Errorf("commit publish tv app release: %w", err)
	}
	return release, nil
}

func (r *VideoRepository) OfflineTVAppRelease(ctx context.Context, releaseID int64) (models.TVAppReleaseRecord, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return models.TVAppReleaseRecord{}, fmt.Errorf("begin offline tv app release: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
UPDATE tv_app_releases
SET publish_status = $2,
    latest_recommended = FALSE,
    last_status_changed_at = NOW(),
    updated_at = NOW()
WHERE id = $1
`, releaseID, models.TVReleaseStatusOffline)
	if err != nil {
		return models.TVAppReleaseRecord{}, fmt.Errorf("offline tv app release: %w", err)
	}
	if err := recalcTVAppRecommendationsTx(ctx, tx); err != nil {
		return models.TVAppReleaseRecord{}, err
	}
	release, err := loadTVAppReleaseRecordTx(ctx, tx, releaseID)
	if err != nil {
		return models.TVAppReleaseRecord{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return models.TVAppReleaseRecord{}, fmt.Errorf("commit offline tv app release: %w", err)
	}
	return release, nil
}

func (r *VideoRepository) RestoreTVAppRelease(ctx context.Context, releaseID int64) (models.TVAppReleaseRecord, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return models.TVAppReleaseRecord{}, fmt.Errorf("begin restore tv app release: %w", err)
	}
	defer tx.Rollback(ctx)

	release, err := loadTVAppReleaseRecordTx(ctx, tx, releaseID)
	if err != nil {
		return models.TVAppReleaseRecord{}, err
	}
	if len(release.ABIItems) == 0 {
		return models.TVAppReleaseRecord{}, models.NewTVAPKDomainError(models.TVAPKErrorReleaseNotPublishable, "没有可下载 ABI 的记录不能恢复发布")
	}
	status := models.TVReleaseStatusForVisibility(true, collectReleaseABIStrings(release.ABIItems))
	_, err = tx.Exec(ctx, `
UPDATE tv_app_releases
SET publish_status = $2,
    last_status_changed_at = NOW(),
    updated_at = NOW()
WHERE id = $1
`, releaseID, status)
	if err != nil {
		return models.TVAppReleaseRecord{}, fmt.Errorf("restore tv app release: %w", err)
	}
	if err := recalcTVAppRecommendationsTx(ctx, tx); err != nil {
		return models.TVAppReleaseRecord{}, err
	}
	release, err = loadTVAppReleaseRecordTx(ctx, tx, releaseID)
	if err != nil {
		return models.TVAppReleaseRecord{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return models.TVAppReleaseRecord{}, fmt.Errorf("commit restore tv app release: %w", err)
	}
	return release, nil
}

func (r *VideoRepository) DeleteTVAppDraftRelease(ctx context.Context, releaseID int64) error {
	tag, err := r.pool.Exec(ctx, `
DELETE FROM tv_app_releases
WHERE id = $1 AND publish_status = $2
`, releaseID, models.TVReleaseStatusDraft)
	if err != nil {
		return fmt.Errorf("delete tv app draft release: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *VideoRepository) GetTVAppReleaseRecord(ctx context.Context, releaseID int64) (models.TVAppReleaseRecord, error) {
	row := r.pool.QueryRow(ctx, `
SELECT
  id,
  package_name,
  version_code,
  version_name,
  COALESCE(release_notes, ''),
  COALESCE(remarks, ''),
  publish_status,
  published_at,
  last_status_changed_at,
  created_at,
  updated_at,
  latest_recommended
FROM tv_app_releases
WHERE id = $1
`, releaseID)
	release, err := scanTVAppReleaseRecord(row)
	if err != nil {
		return models.TVAppReleaseRecord{}, err
	}
	releases := []models.TVAppReleaseRecord{release}
	if err := r.attachTVAppReleaseABIs(ctx, releases); err != nil {
		return models.TVAppReleaseRecord{}, err
	}
	return releases[0], nil
}

func (r *VideoRepository) ListTVAppFamilyReleases(ctx context.Context, limit int) ([]models.TVAppFamilyRelease, error) {
	if limit <= 0 {
		limit = 3
	}
	rows, err := r.pool.Query(ctx, `
SELECT
  id,
  package_name,
  version_code,
  version_name,
  COALESCE(release_notes, ''),
  published_at,
  latest_recommended
FROM tv_app_releases
WHERE publish_status IN ('published_complete', 'published_missing_abi')
ORDER BY version_code DESC, created_at DESC
LIMIT $1
`, limit)
	if err != nil {
		return nil, fmt.Errorf("list family tv app releases: %w", err)
	}
	defer rows.Close()

	records := make([]models.TVAppReleaseRecord, 0, limit)
	for rows.Next() {
		var release models.TVAppReleaseRecord
		var publishedAt sql.NullTime
		if err := rows.Scan(
			&release.ID,
			&release.PackageName,
			&release.VersionCode,
			&release.VersionName,
			&release.ReleaseNotes,
			&publishedAt,
			&release.LatestRecommended,
		); err != nil {
			return nil, fmt.Errorf("scan family tv app release: %w", err)
		}
		release.PublishStatus = models.TVReleaseStatusPublishedComplete
		release.PublishedAt = nullTimePtr(publishedAt)
		records = append(records, release)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return []models.TVAppFamilyRelease{}, nil
	}
	if err := r.attachTVAppReleaseABIs(ctx, records); err != nil {
		return nil, err
	}

	out := make([]models.TVAppFamilyRelease, 0, len(records))
	for _, release := range records {
		abiItems := make([]models.TVAppFamilyReleaseABI, 0, len(release.ABIItems))
		for _, abi := range release.ABIItems {
			updatedAt := abi.UploadedAt
			if abi.ReplacedAt != nil {
				updatedAt = *abi.ReplacedAt
			}
			abiItems = append(abiItems, models.TVAppFamilyReleaseABI{
				ABI:       abi.ABI,
				FileName:  abi.FileName,
				FileSize:  abi.FileSize,
				MIMEType:  abi.MIMEType,
				UpdatedAt: updatedAt,
			})
		}
		out = append(out, models.TVAppFamilyRelease{
			ID:                release.ID,
			PackageName:       release.PackageName,
			VersionCode:       release.VersionCode,
			VersionName:       release.VersionName,
			ReleaseNotes:      release.ReleaseNotes,
			PublishedAt:       release.PublishedAt,
			LatestRecommended: release.LatestRecommended,
			UploadedABIs:      models.TVUploadedABIs(collectReleaseABIStrings(release.ABIItems)),
			MissingABIs:       models.TVMissingABIs(collectReleaseABIStrings(release.ABIItems)),
			ABIItems:          abiItems,
		})
	}
	return out, nil
}

func (r *VideoRepository) GetTVAppReleaseAPKByABI(ctx context.Context, releaseID int64, abi string) (models.TVAppReleaseABIInfo, error) {
	row := r.pool.QueryRow(ctx, `
SELECT
  a.id,
  a.release_id,
  a.abi,
  a.file_name,
  a.stored_path,
  a.file_size,
  a.mime_type,
  a.sha256,
  a.is_debuggable,
  a.uploaded_at,
  a.replaced_at,
  COALESCE(a.upload_user_id::text, ''),
  COALESCE(u.username, ''),
  COALESCE(a.metadata, '{}'::jsonb)
FROM tv_app_release_apks a
LEFT JOIN users u ON u.id = a.upload_user_id
WHERE a.release_id = $1 AND a.abi = $2
`, releaseID, abi)
	return scanTVAppReleaseABI(row)
}

func scanTVAppReleaseRecord(row tvRowScanner) (models.TVAppReleaseRecord, error) {
	var (
		release             models.TVAppReleaseRecord
		publishedAt         sql.NullTime
		lastStatusChangedAt sql.NullTime
	)
	if err := row.Scan(
		&release.ID,
		&release.PackageName,
		&release.VersionCode,
		&release.VersionName,
		&release.ReleaseNotes,
		&release.Remarks,
		&release.PublishStatus,
		&publishedAt,
		&lastStatusChangedAt,
		&release.CreatedAt,
		&release.UpdatedAt,
		&release.LatestRecommended,
	); err != nil {
		return models.TVAppReleaseRecord{}, fmt.Errorf("scan tv app release record: %w", err)
	}
	release.PublishedAt = nullTimePtr(publishedAt)
	if lastStatusChangedAt.Valid {
		release.LastStatusChangedAt = lastStatusChangedAt.Time
	} else {
		release.LastStatusChangedAt = release.UpdatedAt
	}
	return release, nil
}

func scanTVAppReleaseABI(row tvRowScanner) (models.TVAppReleaseABIInfo, error) {
	var (
		item         models.TVAppReleaseABIInfo
		replacedAt   sql.NullTime
		uploadUserID string
	)
	if err := row.Scan(
		&item.ID,
		&item.ReleaseID,
		&item.ABI,
		&item.FileName,
		&item.StoredPath,
		&item.FileSize,
		&item.MIMEType,
		&item.SHA256,
		&item.IsDebuggable,
		&item.UploadedAt,
		&replacedAt,
		&uploadUserID,
		&item.UploadUser,
		&item.Metadata,
	); err != nil {
		return models.TVAppReleaseABIInfo{}, fmt.Errorf("scan tv app release abi: %w", err)
	}
	item.ReplacedAt = nullTimePtr(replacedAt)
	item.UploadUserID = parseNullableUUIDText(uploadUserID)
	return item, nil
}

func buildAdminTVAppReleaseListItem(release models.TVAppReleaseRecord) models.AdminTvAppReleaseListItem {
	abiItems := make([]models.AdminTvAppReleaseABIItem, 0, len(release.ABIItems))
	for _, item := range release.ABIItems {
		abiItems = append(abiItems, models.AdminTvAppReleaseABIItem{
			ID:           item.ID,
			ReleaseID:    item.ReleaseID,
			ABI:          item.ABI,
			FileName:     item.FileName,
			StoredPath:   item.StoredPath,
			FileSize:     item.FileSize,
			MIMEType:     item.MIMEType,
			SHA256:       item.SHA256,
			IsDebuggable: item.IsDebuggable,
			UploadedAt:   item.UploadedAt,
			ReplacedAt:   item.ReplacedAt,
			UploadUserID: item.UploadUserID,
			UploadUser:   item.UploadUser,
			Metadata:     item.Metadata,
		})
	}
	uploaded := models.TVUploadedABIs(collectReleaseABIStrings(release.ABIItems))
	missing := models.TVMissingABIs(uploaded)
	return models.AdminTvAppReleaseListItem{
		ID:                  release.ID,
		PackageName:         release.PackageName,
		VersionCode:         release.VersionCode,
		VersionName:         release.VersionName,
		ReleaseNotes:        release.ReleaseNotes,
		Remarks:             release.Remarks,
		PublishStatus:       release.PublishStatus,
		PublishedAt:         release.PublishedAt,
		LastStatusChangedAt: release.LastStatusChangedAt,
		OriginalUploadedAt:  release.CreatedAt,
		CreatedAt:           release.CreatedAt,
		UpdatedAt:           release.UpdatedAt,
		LatestRecommended:   release.LatestRecommended,
		VisibleToFamily:     models.TVReleaseVisibleToFamily(release.PublishStatus),
		MissingABIs:         missing,
		UploadedABIs:        uploaded,
		ABIComplete:         len(missing) == 0,
		ABIItems:            abiItems,
	}
}

func (r *VideoRepository) attachTVAppReleaseABIs(ctx context.Context, releases []models.TVAppReleaseRecord) error {
	if len(releases) == 0 {
		return nil
	}
	indexByID := make(map[int64]int, len(releases))
	ids := make([]int64, 0, len(releases))
	for i := range releases {
		indexByID[releases[i].ID] = i
		ids = append(ids, releases[i].ID)
		releases[i].ABIItems = nil
	}

	rows, err := r.pool.Query(ctx, `
SELECT
  a.id,
  a.release_id,
  a.abi,
  a.file_name,
  a.stored_path,
  a.file_size,
  a.mime_type,
  a.sha256,
  a.is_debuggable,
  a.uploaded_at,
  a.replaced_at,
  COALESCE(a.upload_user_id::text, ''),
  COALESCE(u.username, ''),
  COALESCE(a.metadata, '{}'::jsonb)
FROM tv_app_release_apks a
LEFT JOIN users u ON u.id = a.upload_user_id
WHERE a.release_id = ANY($1)
ORDER BY a.release_id DESC, a.abi ASC, a.uploaded_at DESC
`, ids)
	if err != nil {
		return fmt.Errorf("query tv app release abis: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		item, scanErr := scanTVAppReleaseABI(rows)
		if scanErr != nil {
			return scanErr
		}
		idx, ok := indexByID[item.ReleaseID]
		if !ok {
			continue
		}
		releases[idx].ABIItems = append(releases[idx].ABIItems, item)
	}
	return rows.Err()
}

func upsertTVAppReleaseRecord(ctx context.Context, tx pgx.Tx, packageName string, versionCode int64, versionName string) (models.TVAppReleaseRecord, error) {
	row := tx.QueryRow(ctx, `
INSERT INTO tv_app_releases (package_name, version_code, version_name)
VALUES ($1, $2, $3)
ON CONFLICT (package_name, version_code)
DO UPDATE SET version_name = EXCLUDED.version_name, updated_at = NOW()
RETURNING
  id,
  package_name,
  version_code,
  version_name,
  COALESCE(release_notes, ''),
  COALESCE(remarks, ''),
  publish_status,
  published_at,
  last_status_changed_at,
  created_at,
  updated_at,
  latest_recommended
`, packageName, versionCode, versionName)
	release, err := scanTVAppReleaseRecord(row)
	if err != nil {
		return models.TVAppReleaseRecord{}, err
	}
	return release, nil
}

func getTVAppReleaseABIByABI(ctx context.Context, tx pgx.Tx, releaseID int64, abi string) (models.TVAppReleaseABIInfo, bool, error) {
	row := tx.QueryRow(ctx, `
SELECT
  a.id,
  a.release_id,
  a.abi,
  a.file_name,
  a.stored_path,
  a.file_size,
  a.mime_type,
  a.sha256,
  a.is_debuggable,
  a.uploaded_at,
  a.replaced_at,
  COALESCE(a.upload_user_id::text, ''),
  COALESCE(u.username, ''),
  COALESCE(a.metadata, '{}'::jsonb)
FROM tv_app_release_apks a
LEFT JOIN users u ON u.id = a.upload_user_id
WHERE a.release_id = $1 AND a.abi = $2
`, releaseID, abi)
	item, err := scanTVAppReleaseABI(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || IsNotFound(err) {
			return models.TVAppReleaseABIInfo{}, false, nil
		}
		return models.TVAppReleaseABIInfo{}, false, err
	}
	return item, true, nil
}

func saveTVAppReleaseABI(
	ctx context.Context,
	tx pgx.Tx,
	release models.TVAppReleaseRecord,
	existing models.TVAppReleaseABIInfo,
	hasExisting bool,
	meta models.TVAppAPKParsedMetadata,
	storedPath string,
	userID *uuid.UUID,
	username string,
) (models.TVAppReleaseABIInfo, error) {
	metadataRaw, err := json.Marshal(map[string]any{
		"package_name":  meta.PackageName,
		"version_code":  meta.VersionCode,
		"version_name":  meta.VersionName,
		"abi":           meta.ABI,
		"is_debuggable": meta.IsDebuggable,
		"sha256":        meta.SHA256,
		"parsed_at":     meta.ParsedAt,
		"raw_manifest":  json.RawMessage(meta.RawManifest),
	})
	if err != nil {
		return models.TVAppReleaseABIInfo{}, fmt.Errorf("marshal tv app abi metadata: %w", err)
	}

	if hasExisting {
		_, err = tx.Exec(ctx, `
UPDATE tv_app_release_apks
SET file_name = $2,
    stored_path = $3,
    file_size = $4,
    mime_type = $5,
    sha256 = $6,
    upload_user_id = $7,
    is_debuggable = $8,
    replaced_at = NOW(),
    uploaded_at = NOW(),
    metadata = $9
WHERE id = $1
`, existing.ID, meta.FileName, storedPath, meta.FileSize, meta.MIMEType, meta.SHA256, userID, meta.IsDebuggable, metadataRaw)
		if err != nil {
			return models.TVAppReleaseABIInfo{}, fmt.Errorf("replace tv app release abi: %w", err)
		}
		if _, err := tx.Exec(ctx, `
UPDATE tv_app_releases
SET last_status_changed_at = NOW(),
    updated_at = NOW()
WHERE id = $1
`, release.ID); err != nil {
			return models.TVAppReleaseABIInfo{}, fmt.Errorf("touch tv app release after abi replace: %w", err)
		}
		row := tx.QueryRow(ctx, `
SELECT
  a.id,
  a.release_id,
  a.abi,
  a.file_name,
  a.stored_path,
  a.file_size,
  a.mime_type,
  a.sha256,
  a.is_debuggable,
  a.uploaded_at,
  a.replaced_at,
  COALESCE(a.upload_user_id::text, ''),
  $2,
  COALESCE(a.metadata, '{}'::jsonb)
FROM tv_app_release_apks a
WHERE a.id = $1
`, existing.ID, username)
		return scanTVAppReleaseABI(row)
	}

	row := tx.QueryRow(ctx, `
INSERT INTO tv_app_release_apks (
  release_id,
  abi,
  file_name,
  stored_path,
  file_size,
  mime_type,
  sha256,
  upload_user_id,
  is_debuggable,
  metadata
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
RETURNING
  id,
  release_id,
  abi,
  file_name,
  stored_path,
  file_size,
  mime_type,
  sha256,
  is_debuggable,
  uploaded_at,
  replaced_at,
  COALESCE(upload_user_id::text, ''),
  $11,
  COALESCE(metadata, '{}'::jsonb)
`, release.ID, meta.ABI, meta.FileName, storedPath, meta.FileSize, meta.MIMEType, meta.SHA256, userID, meta.IsDebuggable, metadataRaw, username)
	item, err := scanTVAppReleaseABI(row)
	if err != nil {
		return models.TVAppReleaseABIInfo{}, fmt.Errorf("insert tv app release abi: %w", err)
	}

	visible := models.TVReleaseVisibleToFamily(release.PublishStatus)
	if visible {
		release.ABIItems = append(release.ABIItems, item)
		status := models.TVReleaseStatusForVisibility(true, collectReleaseABIStrings(release.ABIItems))
		if _, err := tx.Exec(ctx, `
UPDATE tv_app_releases
SET publish_status = $2,
    last_status_changed_at = NOW(),
    updated_at = NOW()
WHERE id = $1
`, release.ID, status); err != nil {
			return models.TVAppReleaseABIInfo{}, fmt.Errorf("update tv app release status after abi insert: %w", err)
		}
	}
	return item, nil
}

func loadTVAppReleaseRecordTx(ctx context.Context, tx pgx.Tx, releaseID int64) (models.TVAppReleaseRecord, error) {
	row := tx.QueryRow(ctx, `
SELECT
  id,
  package_name,
  version_code,
  version_name,
  COALESCE(release_notes, ''),
  COALESCE(remarks, ''),
  publish_status,
  published_at,
  last_status_changed_at,
  created_at,
  updated_at,
  latest_recommended
FROM tv_app_releases
WHERE id = $1
`, releaseID)
	release, err := scanTVAppReleaseRecord(row)
	if err != nil {
		return models.TVAppReleaseRecord{}, err
	}
	rows, err := tx.Query(ctx, `
SELECT
  a.id,
  a.release_id,
  a.abi,
  a.file_name,
  a.stored_path,
  a.file_size,
  a.mime_type,
  a.sha256,
  a.is_debuggable,
  a.uploaded_at,
  a.replaced_at,
  COALESCE(a.upload_user_id::text, ''),
  COALESCE(u.username, ''),
  COALESCE(a.metadata, '{}'::jsonb)
FROM tv_app_release_apks a
LEFT JOIN users u ON u.id = a.upload_user_id
WHERE a.release_id = $1
ORDER BY a.abi ASC, a.uploaded_at DESC
`, releaseID)
	if err != nil {
		return models.TVAppReleaseRecord{}, fmt.Errorf("query tv app release abis tx: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		item, scanErr := scanTVAppReleaseABI(rows)
		if scanErr != nil {
			return models.TVAppReleaseRecord{}, scanErr
		}
		release.ABIItems = append(release.ABIItems, item)
	}
	if err := rows.Err(); err != nil {
		return models.TVAppReleaseRecord{}, err
	}
	return release, nil
}

func recalcTVAppRecommendationsTx(ctx context.Context, tx pgx.Tx) error {
	if _, err := tx.Exec(ctx, `UPDATE tv_app_releases SET latest_recommended = FALSE WHERE latest_recommended = TRUE`); err != nil {
		return fmt.Errorf("reset tv app recommendations: %w", err)
	}
	if _, err := tx.Exec(ctx, `
UPDATE tv_app_releases
SET latest_recommended = TRUE
WHERE id = (
  SELECT id
  FROM tv_app_releases
  WHERE publish_status IN ('published_complete', 'published_missing_abi')
  ORDER BY version_code DESC, created_at DESC
  LIMIT 1
)
`); err != nil {
		return fmt.Errorf("recalculate tv app recommendation: %w", err)
	}
	return nil
}

func collectReleaseABIStrings(items []models.TVAppReleaseABIInfo) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.ABI)
	}
	return out
}

func parseVersionKeyword(keyword string) string {
	digits := make([]rune, 0, len(keyword))
	for _, r := range keyword {
		if r >= '0' && r <= '9' {
			digits = append(digits, r)
		}
	}
	if len(digits) == 0 {
		return "%"
	}
	return "%" + string(digits) + "%"
}
