package sqlite

import (
	"context"
	"database/sql"

	"github.com/TimNikolaev/OKeeper/internal/models"
)

type userRepo struct {
	db *sql.DB
}

func (r *userRepo) GetByID(ctx context.Context, id int64) (*models.User, error) {
	query := `SELECT id, username, password_hash, full_name, role, created_at FROM users WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)
	var u models.User
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.FullName, &u.Role, &u.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `SELECT id, username, password_hash, full_name, role, created_at FROM users WHERE username = ?`
	row := r.db.QueryRowContext(ctx, query, username)
	var u models.User
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.FullName, &u.Role, &u.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepo) Create(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (username, password_hash, full_name, role) VALUES (?, ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, query, user.Username, user.PasswordHash, user.FullName, user.Role)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = id
	return nil
}
