package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"video-server/internal/models"
)

func scanPasswordVaultEntry(row rowScanner) (models.AdminPasswordVaultEntry, string, error) {
	var item models.AdminPasswordVaultEntry
	var ciphertext string
	if err := row.Scan(
		&item.ID,
		&item.Name,
		&item.Account,
		&ciphertext,
		&item.URL,
		&item.Note,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return models.AdminPasswordVaultEntry{}, "", err
	}
	return item, ciphertext, nil
}

func normalizePasswordVaultInput(in models.AdminPasswordVaultInput) models.AdminPasswordVaultInput {
	in.Name = strings.Join(strings.Fields(strings.TrimSpace(in.Name)), " ")
	in.Account = strings.TrimSpace(in.Account)
	in.URL = strings.TrimSpace(in.URL)
	in.Note = strings.TrimSpace(in.Note)
	return in
}

func (r *VideoRepository) ListPasswordVaultEntries(ctx context.Context, q string, page, pageSize int) ([]models.AdminPasswordVaultEntry, int, error) {
	where := []string{"1=1"}
	args := make([]any, 0, 4)
	next := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	if keyword := strings.ToLower(strings.TrimSpace(q)); keyword != "" {
		pattern := "%" + keyword + "%"
		placeholder := next(pattern)
		where = append(where, "(LOWER(name) LIKE "+placeholder+" OR LOWER(account) LIKE "+placeholder+" OR LOWER(url) LIKE "+placeholder+" OR LOWER(note) LIKE "+placeholder+")")
	}

	baseWhere := strings.Join(where, " AND ")
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM password_vault_entries WHERE "+baseWhere, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count password vault entries: %w", err)
	}

	args = append(args, pageSize, (page-1)*pageSize)
	sql := "SELECT id, name, account, password_ciphertext, url, note, created_at, updated_at FROM password_vault_entries WHERE " +
		baseWhere + " ORDER BY updated_at DESC, name ASC LIMIT $" + fmt.Sprintf("%d", len(args)-1) +
		" OFFSET $" + fmt.Sprintf("%d", len(args))
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list password vault entries: %w", err)
	}
	defer rows.Close()

	items := make([]models.AdminPasswordVaultEntry, 0, pageSize)
	for rows.Next() {
		item, _, scanErr := scanPasswordVaultEntry(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("scan password vault entry: %w", scanErr)
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *VideoRepository) CreatePasswordVaultEntry(ctx context.Context, input models.AdminPasswordVaultInput, passwordCiphertext string) (models.AdminPasswordVaultEntry, error) {
	input = normalizePasswordVaultInput(input)
	row := r.pool.QueryRow(ctx, `
INSERT INTO password_vault_entries (id, name, account, password_ciphertext, url, note)
VALUES ($1,$2,$3,$4,$5,$6)
RETURNING id, name, account, password_ciphertext, url, note, created_at, updated_at
`, uuid.New(), input.Name, input.Account, passwordCiphertext, input.URL, input.Note)
	item, _, err := scanPasswordVaultEntry(row)
	if err != nil {
		return models.AdminPasswordVaultEntry{}, fmt.Errorf("create password vault entry: %w", err)
	}
	return item, nil
}

func (r *VideoRepository) UpdatePasswordVaultEntry(ctx context.Context, entryID uuid.UUID, input models.AdminPasswordVaultInput, passwordCiphertext *string) (models.AdminPasswordVaultEntry, error) {
	input = normalizePasswordVaultInput(input)
	var row pgx.Row
	if passwordCiphertext != nil {
		row = r.pool.QueryRow(ctx, `
UPDATE password_vault_entries
SET name=$2, account=$3, password_ciphertext=$4, url=$5, note=$6, updated_at=NOW()
WHERE id=$1
RETURNING id, name, account, password_ciphertext, url, note, created_at, updated_at
`, entryID, input.Name, input.Account, *passwordCiphertext, input.URL, input.Note)
	} else {
		row = r.pool.QueryRow(ctx, `
UPDATE password_vault_entries
SET name=$2, account=$3, url=$4, note=$5, updated_at=NOW()
WHERE id=$1
RETURNING id, name, account, password_ciphertext, url, note, created_at, updated_at
`, entryID, input.Name, input.Account, input.URL, input.Note)
	}
	item, _, err := scanPasswordVaultEntry(row)
	if err != nil {
		return models.AdminPasswordVaultEntry{}, fmt.Errorf("update password vault entry: %w", err)
	}
	return item, nil
}

func (r *VideoRepository) DeletePasswordVaultEntry(ctx context.Context, entryID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM password_vault_entries WHERE id=$1`, entryID)
	if err != nil {
		return fmt.Errorf("delete password vault entry: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *VideoRepository) GetPasswordVaultCiphertext(ctx context.Context, entryID uuid.UUID) (string, error) {
	var ciphertext string
	err := r.pool.QueryRow(ctx, `SELECT password_ciphertext FROM password_vault_entries WHERE id=$1`, entryID).Scan(&ciphertext)
	if err != nil {
		return "", fmt.Errorf("get password vault ciphertext: %w", err)
	}
	return ciphertext, nil
}
