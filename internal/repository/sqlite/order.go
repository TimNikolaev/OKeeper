package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/TimNikolaev/OKeeper/internal/models"
)

type orderRepo struct {
	db *sql.DB
}

func (r *orderRepo) Create(ctx context.Context, order *models.Order, items []models.OrderDish) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	orderQuery := `INSERT INTO orders (table_num, waiter_id, status, comment) VALUES (?, ?, ?, ?)`
	res, err := tx.ExecContext(ctx, orderQuery, order.TableNum, order.WaiterID, order.Status, order.Comment)
	if err != nil {
		return err
	}

	orderID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	order.ID = orderID

	itemQuery := `INSERT INTO orders_dishes (order_id, dish_id, quantity, price, comment) VALUES (?, ?, ?, ?, ?)`
	for i := range items {
		items[i].OrderID = orderID
		resItem, err := tx.ExecContext(ctx, itemQuery, orderID, items[i].DishID, items[i].Quantity, items[i].Price, items[i].Comment)
		if err != nil {
			return err
		}
		itemID, err := resItem.LastInsertId()
		if err != nil {
			return err
		}
		items[i].ID = itemID
	}
	order.Items = items

	return tx.Commit()
}

func (r *orderRepo) GetByID(ctx context.Context, id int64) (*models.Order, error) {
	orderQuery := `SELECT id, table_num, waiter_id, status, comment, created_at FROM orders WHERE id = ?`
	row := r.db.QueryRowContext(ctx, orderQuery, id)
	var o models.Order
	if err := row.Scan(&o.ID, &o.TableNum, &o.WaiterID, &o.Status, &o.Comment, &o.CreatedAt); err != nil {
		return nil, err
	}

	itemsQuery := `SELECT id, order_id, dish_id, quantity, price, comment FROM orders_dishes WHERE order_id = ?`
	rows, err := r.db.QueryContext(ctx, itemsQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.OrderDish
	for rows.Next() {
		var item models.OrderDish
		if err := rows.Scan(&item.ID, &item.OrderID, &item.DishID, &item.Quantity, &item.Price, &item.Comment); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	o.Items = items

	return &o, nil
}

func (r *orderRepo) GetQueue(ctx context.Context, statuses []models.OrderStatus) ([]models.Order, error) {
	if len(statuses) == 0 {
		return nil, fmt.Errorf("statuses list cannot be empty")
	}

	query := `SELECT id, table_num, waiter_id, status, comment, created_at FROM orders WHERE status IN (`
	args := make([]interface{}, len(statuses))
	for i, s := range statuses {
		if i > 0 {
			query += ", "
		}
		query += "?"
		args[i] = s
	}
	query += `) ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.TableNum, &o.WaiterID, &o.Status, &o.Comment, &o.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *orderRepo) UpdateStatus(ctx context.Context, id int64, status models.OrderStatus) error {
	query := `UPDATE orders SET status = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}
