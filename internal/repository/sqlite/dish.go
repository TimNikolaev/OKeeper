package sqlite

import (
	"context"
	"database/sql"

	"github.com/TimNikolaev/OKeeper/internal/models"
)

type dishRepo struct {
	db *sql.DB
}

func (r *dishRepo) GetAll(ctx context.Context) ([]models.Dish, error) {
	query := `SELECT id, category, name, description, price, is_available, created_at FROM dishes`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dishes []models.Dish
	for rows.Next() {
		var d models.Dish
		if err := rows.Scan(&d.ID, &d.Category, &d.Name, &d.Description, &d.Price, &d.IsAvailable, &d.CreatedAt); err != nil {
			return nil, err
		}
		dishes = append(dishes, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return dishes, nil
}

func (r *dishRepo) GetByID(ctx context.Context, id int64) (*models.Dish, error) {
	query := `SELECT id, category, name, description, price, is_available, created_at FROM dishes WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)
	var d models.Dish
	if err := row.Scan(&d.ID, &d.Category, &d.Name, &d.Description, &d.Price, &d.IsAvailable, &d.CreatedAt); err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *dishRepo) Create(ctx context.Context, dish *models.Dish) error {
	query := `INSERT INTO dishes (category, name, description, price, is_available) VALUES (?, ?, ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, query, dish.Category, dish.Name, dish.Description, dish.Price, dish.IsAvailable)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	dish.ID = id
	return nil
}

func (r *dishRepo) Update(ctx context.Context, dish *models.Dish) error {
	query := `UPDATE dishes SET category = ?, name = ?, description = ?, price = ?, is_available = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, dish.Category, dish.Name, dish.Description, dish.Price, dish.IsAvailable, dish.ID)
	return err
}

func (r *dishRepo) SetAvailability(ctx context.Context, id int64, isAvailable int) error {
	query := `UPDATE dishes SET is_available = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, isAvailable, id)
	return err
}
