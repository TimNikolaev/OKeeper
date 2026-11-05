package repository

import (
	"context"
	"time"

	"github.com/TimNikolaev/OKeeper/internal/models"
)

type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
}

type DishRepository interface {
	GetAll(ctx context.Context) ([]models.Dish, error)
	GetByID(ctx context.Context, id int64) (*models.Dish, error)
	Create(ctx context.Context, dish *models.Dish) error
	Update(ctx context.Context, dish *models.Dish) error
	SetAvailability(ctx context.Context, id int64, isAvailable int) error
}

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order, items []models.OrderDish) error
	GetByID(ctx context.Context, id int64) (*models.Order, error)
	GetQueue(ctx context.Context, statuses []models.OrderStatus) ([]models.Order, error)
	UpdateStatus(ctx context.Context, id int64, status models.OrderStatus) error
}

type BillRepository interface {
	Create(ctx context.Context, bill *models.Bill) error
	GetByOrderID(ctx context.Context, orderID int64) (*models.Bill, error)
	MarkAsPaid(ctx context.Context, id int64) error
	GetReports(ctx context.Context, dateFrom, dateTo time.Time) ([]models.Bill, error)
}

type Repository interface {
	Users() UserRepository
	Dishes() DishRepository
	Orders() OrderRepository
	Bills() BillRepository
}
