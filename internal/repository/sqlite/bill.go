package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/TimNikolaev/OKeeper/internal/models"
)

type billRepo struct {
	db *sql.DB
}

func (r *billRepo) Create(ctx context.Context, bill *models.Bill) error {
	query := `INSERT INTO bills (order_id, total_amount, is_paid) VALUES (?, ?, ?)`
	res, err := r.db.ExecContext(ctx, query, bill.OrderID, bill.TotalAmount, bill.IsPaid)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	bill.ID = id
	return nil
}

func (r *billRepo) GetByOrderID(ctx context.Context, orderID int64) (*models.Bill, error) {
	query := `SELECT id, order_id, total_amount, is_paid, created_at, closed_at FROM bills WHERE order_id = ?`
	row := r.db.QueryRowContext(ctx, query, orderID)
	var b models.Bill
	var closedAt sql.NullTime
	if err := row.Scan(&b.ID, &b.OrderID, &b.TotalAmount, &b.IsPaid, &b.CreatedAt, &closedAt); err != nil {
		return nil, err
	}
	if closedAt.Valid {
		b.ClosedAt = &closedAt.Time
	}
	return &b, nil
}

func (r *billRepo) MarkAsPaid(ctx context.Context, id int64) error {
	query := `UPDATE bills SET is_paid = 1, closed_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *billRepo) GetReports(ctx context.Context, dateFrom, dateTo time.Time) ([]models.Bill, error) {
	query := `SELECT id, order_id, total_amount, is_paid, created_at, closed_at FROM bills WHERE created_at BETWEEN ? AND ?`
	rows, err := r.db.QueryContext(ctx, query, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bills []models.Bill
	for rows.Next() {
		var b models.Bill
		var closedAt sql.NullTime
		if err := rows.Scan(&b.ID, &b.OrderID, &b.TotalAmount, &b.IsPaid, &b.CreatedAt, &closedAt); err != nil {
			return nil, err
		}
		if closedAt.Valid {
			b.ClosedAt = &closedAt.Time
		}
		bills = append(bills, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return bills, nil
}
