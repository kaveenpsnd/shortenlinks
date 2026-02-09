package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/kaveenpsnd/url-shortener/internal/core/domain"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Upsert(ctx context.Context, user domain.User) error {
	// This query does a "Do Nothing" on conflict, effectively creating the user
	// only the first time they login.
	query := `
		INSERT INTO users (id, email, role, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE 
		SET email = EXCLUDED.email;
	`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Email, user.Role, user.CreatedAt)
	return err
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (domain.User, error) {
	query := `SELECT id, email, role, created_at FROM users WHERE id = $1`
	var user domain.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Email, &user.Role, &user.CreatedAt)
	return user, err
}
func (r *PostgresUserRepository) GetAll(ctx context.Context) ([]domain.User, error) {
	query := `SELECT id, email, role, created_at FROM users ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}
func (r *PostgresUserRepository) Delete(ctx context.Context, id string) error {
	// 1. Delete links first (Manual Cascade if DB doesn't handle it)
	_, err := r.db.ExecContext(ctx, "DELETE FROM links WHERE user_id = $1", id)
	if err != nil {
		return err
	}

	// 2. Delete the user
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("user not found")
	}
	return nil
}
