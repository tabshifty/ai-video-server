package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"video-server/internal/models"
)

func (r *VideoRepository) CreateUser(ctx context.Context, user models.User) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO users (id, username, email, password_hash, role)
VALUES ($1,$2,$3,$4,$5)
`, user.ID, strings.ToLower(strings.TrimSpace(user.Username)), strings.ToLower(strings.TrimSpace(user.Email)), user.PasswordHash, user.Role)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *VideoRepository) ExistsUsernameOrEmail(ctx context.Context, username, email string) (bool, error) {
	var found bool
	err := r.pool.QueryRow(ctx, `
SELECT EXISTS(
	SELECT 1
	FROM users
	WHERE username = $1 OR email = $2
)`, strings.ToLower(strings.TrimSpace(username)), strings.ToLower(strings.TrimSpace(email))).Scan(&found)
	if err != nil {
		return false, fmt.Errorf("check user exists: %w", err)
	}
	return found, nil
}

func (r *VideoRepository) GetUserByUsernameOrEmail(ctx context.Context, identity string) (models.User, error) {
	identity = strings.ToLower(strings.TrimSpace(identity))
	var user models.User
	err := r.pool.QueryRow(ctx, `
SELECT id, username, email, password_hash, role, created_at, updated_at
FROM users
WHERE username = $1 OR email = $1
LIMIT 1`, identity).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return models.User{}, fmt.Errorf("get user by identity: %w", err)
	}
	return user, nil
}

func (r *VideoRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (models.User, error) {
	var user models.User
	err := r.pool.QueryRow(ctx, `
SELECT id, username, email, password_hash, role, created_at, updated_at
FROM users
WHERE id = $1`, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return models.User{}, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func IsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "23505"
}
