package repo

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"phsio_track_backend/internal/core"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (core.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, username, password_hash, created_time
		FROM users
		WHERE username = :1
	`, username)
	var u core.User
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedTime); err != nil {
		if err == sql.ErrNoRows {
			return u, ErrNotFound
		}
		return u, err
	}
	return u, nil
}

func (r *UserRepo) UpsertUser(ctx context.Context, user core.User) error {
	if user.ID == "" {
		user.ID = uuid.NewString()
	}
	_, err := r.db.ExecContext(ctx, `
		MERGE INTO users u
		USING (SELECT :1 AS username, :2 AS password_hash, :3 AS id FROM dual) s
		ON (u.username = s.username)
		WHEN MATCHED THEN UPDATE SET u.password_hash = s.password_hash
		WHEN NOT MATCHED THEN INSERT (id, username, password_hash, created_time)
		VALUES (s.id, s.username, s.password_hash, SYSTIMESTAMP)
	`, user.Username, user.PasswordHash, user.ID)
	return err
}
