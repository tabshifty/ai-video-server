package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"video-server/internal/models"
)

func normalizeActorName(raw string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(raw)), " "))
}

func normalizeActorAliases(aliases []string, actorName string) []string {
	out := make([]string, 0, len(aliases))
	seen := map[string]struct{}{}
	nameNorm := normalizeActorName(actorName)
	for _, alias := range aliases {
		cleaned := strings.Join(strings.Fields(strings.TrimSpace(alias)), " ")
		if cleaned == "" {
			continue
		}
		norm := normalizeActorName(cleaned)
		if norm == "" || norm == nameNorm {
			continue
		}
		if _, ok := seen[norm]; ok {
			continue
		}
		seen[norm] = struct{}{}
		out = append(out, cleaned)
	}
	return out
}

func parseActorBirthDate(raw string) (*time.Time, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return nil, fmt.Errorf("出生日期格式错误，应为 YYYY-MM-DD")
	}
	return &t, nil
}

func normalizeActorInput(in models.AdminActorInput) (models.AdminActorInput, error) {
	in.Name = strings.Join(strings.Fields(strings.TrimSpace(in.Name)), " ")
	if in.Name == "" {
		return in, fmt.Errorf("演员姓名不能为空")
	}
	in.Aliases = normalizeActorAliases(in.Aliases, in.Name)
	in.Gender = strings.TrimSpace(in.Gender)
	in.Country = strings.TrimSpace(in.Country)
	in.BirthDate = strings.TrimSpace(in.BirthDate)
	in.AvatarURL = strings.TrimSpace(in.AvatarURL)
	in.Source = strings.ToLower(strings.TrimSpace(in.Source))
	in.ExternalID = strings.TrimSpace(in.ExternalID)
	in.Notes = strings.TrimSpace(in.Notes)
	if in.Source == "" {
		in.Source = "manual"
	}
	return in, nil
}

func scanAdminActor(rows rowScanner) (models.AdminActor, error) {
	var out models.AdminActor
	if err := rows.Scan(
		&out.ID,
		&out.Name,
		&out.Aliases,
		&out.Gender,
		&out.Country,
		&out.BirthDate,
		&out.AvatarURL,
		&out.Source,
		&out.ExternalID,
		&out.Notes,
		&out.Active,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		return models.AdminActor{}, err
	}
	return out, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func (r *VideoRepository) ListActors(ctx context.Context, q string, active *bool, page, pageSize int) ([]models.AdminActor, int, error) {
	where := []string{"1=1"}
	args := make([]any, 0, 6)
	next := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	keyword := strings.ToLower(strings.TrimSpace(q))
	if keyword != "" {
		pattern := "%" + keyword + "%"
		placeholder := next(pattern)
		where = append(where, "(LOWER(name) LIKE "+placeholder+" OR EXISTS (SELECT 1 FROM unnest(aliases) alias WHERE LOWER(alias) LIKE "+placeholder+"))")
	}
	if active != nil {
		where = append(where, "active = "+next(*active))
	}

	baseWhere := strings.Join(where, " AND ")
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM actors WHERE "+baseWhere, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count actors: %w", err)
	}

	args = append(args, pageSize, (page-1)*pageSize)
	sql := "SELECT id, name, aliases, COALESCE(gender,''), COALESCE(country,''), COALESCE(to_char(birth_date, 'YYYY-MM-DD'), ''), COALESCE(avatar_url,''), COALESCE(source,''), COALESCE(external_id,''), COALESCE(notes,''), active, created_at, updated_at FROM actors WHERE " + baseWhere + " ORDER BY active DESC, updated_at DESC, name ASC LIMIT $" + fmt.Sprintf("%d", len(args)-1) + " OFFSET $" + fmt.Sprintf("%d", len(args))
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list actors: %w", err)
	}
	defer rows.Close()

	items := make([]models.AdminActor, 0, pageSize)
	for rows.Next() {
		item, scanErr := scanAdminActor(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("scan actor: %w", scanErr)
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *VideoRepository) CreateActor(ctx context.Context, input models.AdminActorInput) (models.AdminActor, error) {
	input, err := normalizeActorInput(input)
	if err != nil {
		return models.AdminActor{}, err
	}
	birthDate, err := parseActorBirthDate(input.BirthDate)
	if err != nil {
		return models.AdminActor{}, err
	}

	row := r.pool.QueryRow(ctx, `
INSERT INTO actors (
  id, name, normalized_name, aliases, gender, country, birth_date, avatar_url, source, external_id, notes, active
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
RETURNING
  id, name, aliases, COALESCE(gender,''), COALESCE(country,''), COALESCE(to_char(birth_date, 'YYYY-MM-DD'), ''),
  COALESCE(avatar_url,''), COALESCE(source,''), COALESCE(external_id,''), COALESCE(notes,''), active, created_at, updated_at
`, uuid.New(), input.Name, normalizeActorName(input.Name), input.Aliases, input.Gender, input.Country, birthDate, input.AvatarURL, input.Source, input.ExternalID, input.Notes, input.Active)
	out, err := scanAdminActor(row)
	if err != nil {
		return models.AdminActor{}, fmt.Errorf("create actor: %w", err)
	}
	return out, nil
}

func (r *VideoRepository) UpdateActor(ctx context.Context, actorID uuid.UUID, input models.AdminActorInput) (models.AdminActor, error) {
	input, err := normalizeActorInput(input)
	if err != nil {
		return models.AdminActor{}, err
	}
	birthDate, err := parseActorBirthDate(input.BirthDate)
	if err != nil {
		return models.AdminActor{}, err
	}

	row := r.pool.QueryRow(ctx, `
UPDATE actors
SET
  name=$2,
  normalized_name=$3,
  aliases=$4,
  gender=$5,
  country=$6,
  birth_date=$7,
  avatar_url=$8,
  source=$9,
  external_id=$10,
  notes=$11,
  active=$12,
  updated_at=NOW()
WHERE id=$1
RETURNING
  id, name, aliases, COALESCE(gender,''), COALESCE(country,''), COALESCE(to_char(birth_date, 'YYYY-MM-DD'), ''),
  COALESCE(avatar_url,''), COALESCE(source,''), COALESCE(external_id,''), COALESCE(notes,''), active, created_at, updated_at
`, actorID, input.Name, normalizeActorName(input.Name), input.Aliases, input.Gender, input.Country, birthDate, input.AvatarURL, input.Source, input.ExternalID, input.Notes, input.Active)
	out, err := scanAdminActor(row)
	if err != nil {
		return models.AdminActor{}, fmt.Errorf("update actor: %w", err)
	}
	return out, nil
}

func (r *VideoRepository) ListVideoActors(ctx context.Context, videoID uuid.UUID) ([]models.AdminVideoActor, error) {
	rows, err := r.pool.Query(ctx, `
SELECT a.id, a.name, COALESCE(a.avatar_url,''), a.active, COALESCE(va.source,'')
FROM video_actors va
JOIN actors a ON a.id = va.actor_id
WHERE va.video_id=$1
ORDER BY va.created_at ASC, a.name ASC
`, videoID)
	if err != nil {
		return nil, fmt.Errorf("list video actors: %w", err)
	}
	defer rows.Close()

	out := make([]models.AdminVideoActor, 0, 8)
	for rows.Next() {
		var item models.AdminVideoActor
		if err := rows.Scan(&item.ID, &item.Name, &item.AvatarURL, &item.Active, &item.BindSource); err != nil {
			return nil, fmt.Errorf("scan video actor: %w", err)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *VideoRepository) ResolveActorIDs(ctx context.Context, actorIDs []uuid.UUID, actorNames []string, source string) ([]uuid.UUID, error) {
	if strings.TrimSpace(source) == "" {
		source = "manual"
	}
	out := make([]uuid.UUID, 0, len(actorIDs)+len(actorNames))
	seen := map[uuid.UUID]struct{}{}

	if len(actorIDs) > 0 {
		rows, err := r.pool.Query(ctx, `SELECT id FROM actors WHERE id = ANY($1)`, actorIDs)
		if err != nil {
			return nil, fmt.Errorf("query actor ids: %w", err)
		}
		defer rows.Close()
		existing := map[uuid.UUID]struct{}{}
		for rows.Next() {
			var id uuid.UUID
			if err := rows.Scan(&id); err != nil {
				return nil, fmt.Errorf("scan actor id: %w", err)
			}
			existing[id] = struct{}{}
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		for _, id := range actorIDs {
			if _, ok := existing[id]; !ok {
				return nil, fmt.Errorf("演员不存在: %s", id.String())
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}

	nameSeen := map[string]struct{}{}
	for _, raw := range actorNames {
		name := strings.Join(strings.Fields(strings.TrimSpace(raw)), " ")
		if name == "" {
			continue
		}
		nameKey := normalizeActorName(name)
		if nameKey == "" {
			continue
		}
		if _, ok := nameSeen[nameKey]; ok {
			continue
		}
		nameSeen[nameKey] = struct{}{}
		actorID, err := r.upsertActorByName(ctx, name, source)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[actorID]; ok {
			continue
		}
		seen[actorID] = struct{}{}
		out = append(out, actorID)
	}

	return out, nil
}

func (r *VideoRepository) upsertActorByName(ctx context.Context, name, source string) (uuid.UUID, error) {
	cleanedName := strings.Join(strings.Fields(strings.TrimSpace(name)), " ")
	if cleanedName == "" {
		return uuid.Nil, fmt.Errorf("演员姓名不能为空")
	}
	if strings.TrimSpace(source) == "" {
		source = "manual"
	}
	var actorID uuid.UUID
	if err := r.pool.QueryRow(ctx, `
INSERT INTO actors(id, name, normalized_name, source, active)
VALUES ($1, $2, $3, $4, TRUE)
ON CONFLICT(normalized_name) DO UPDATE
SET
  updated_at=NOW(),
  aliases = CASE
    WHEN EXCLUDED.name <> actors.name
      AND NOT (EXCLUDED.name = ANY(actors.aliases))
    THEN array_append(actors.aliases, EXCLUDED.name)
    ELSE actors.aliases
  END
RETURNING id
`, uuid.New(), cleanedName, normalizeActorName(cleanedName), source).Scan(&actorID); err != nil {
		return uuid.Nil, fmt.Errorf("upsert actor by name: %w", err)
	}
	return actorID, nil
}

func dedupeActorIDs(ids []uuid.UUID) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(ids))
	seen := map[uuid.UUID]struct{}{}
	for _, id := range ids {
		if id == uuid.Nil {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func (r *VideoRepository) AddVideoActors(ctx context.Context, videoID uuid.UUID, actorIDs []uuid.UUID, source string) error {
	if strings.TrimSpace(source) == "" {
		source = "manual"
	}
	actorIDs = dedupeActorIDs(actorIDs)
	if len(actorIDs) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for _, actorID := range actorIDs {
		batch.Queue(
			`INSERT INTO video_actors(video_id, actor_id, source) VALUES ($1,$2,$3)
ON CONFLICT(video_id, actor_id) DO UPDATE SET source = EXCLUDED.source`,
			videoID, actorID, source,
		)
	}
	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()
	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("add video actor: %w", err)
		}
	}
	return nil
}

func (r *VideoRepository) ReplaceVideoActors(ctx context.Context, videoID uuid.UUID, actorIDs []uuid.UUID, source string) error {
	if strings.TrimSpace(source) == "" {
		source = "manual"
	}
	actorIDs = dedupeActorIDs(actorIDs)
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx replace video actors: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM video_actors WHERE video_id=$1`, videoID); err != nil {
		return fmt.Errorf("clear video actors: %w", err)
	}
	for _, actorID := range actorIDs {
		if _, err := tx.Exec(ctx, `INSERT INTO video_actors(video_id, actor_id, source) VALUES ($1,$2,$3)`, videoID, actorID, source); err != nil {
			return fmt.Errorf("insert video actor: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit replace video actors: %w", err)
	}
	return nil
}

func (r *VideoRepository) ReplaceVideoActorsByInput(ctx context.Context, videoID uuid.UUID, actorIDs []uuid.UUID, actorNames []string, source string) error {
	resolved, err := r.ResolveActorIDs(ctx, actorIDs, actorNames, source)
	if err != nil {
		return err
	}
	return r.ReplaceVideoActors(ctx, videoID, resolved, source)
}
