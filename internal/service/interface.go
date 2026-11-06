package service

import (
	"context"
	"time"

	"github.com/TimNikolaev/OKeeper/internal/models"
)

type AuthService interface {
	Login(ctx context.Context, username, password string) (*models.User, error)
}

type CreateUserInput struct {
	Username string
	Password string
	FullName string
	Role     models.Role
}

type UserService interface {
	Create(ctx context.Context, actor *models.User, in CreateUserInput) (*models.User, error)
	GetByID(ctx context.Context, actor *models.User, id int64) (*models.User, error)
}

type DishInput struct {
	Category    string
	Name        string
	Description *string
	Price       int
	IsAvailable bool
}

type MenuService interface {
	List(ctx context.Context, actor *models.User) ([]models.Dish, error)
	ListAvailable(ctx context.Context, actor *models.User) ([]models.Dish, error)
	Create(ctx context.Context, actor *models.User, in DishInput) (*models.Dish, error)
	Update(ctx context.Context, actor *models.User, id int64, in DishInput) (*models.Dish, error)
	SetAvailability(ctx context.Context, actor *models.User, id int64, available bool) error
}

type OrderItemInput struct {
	DishID   int64
	Quantity int
	Comment  *string
}

type CreateOrderInput struct {
	TableNum int
	Comment  *string
	Items    []OrderItemInput
}

// OrderFilter: пустые поля не участвуют в фильтрации.
type OrderFilter struct {
	Statuses []models.OrderStatus
	TableNum *int
	WaiterID *int64
}

type OrderService interface {
	Create(ctx context.Context, actor *models.User, in CreateOrderInput) (*models.Order, error)
	GetByID(ctx context.Context, actor *models.User, id int64) (*models.Order, error)
	List(ctx context.Context, actor *models.User, filter OrderFilter) ([]models.Order, error)
	CookQueue(ctx context.Context, actor *models.User) ([]models.Order, error)
	ChangeStatus(ctx context.Context, actor *models.User, orderID int64, status models.OrderStatus) (*models.Order, error)
}

type ReceiptLine struct {
	DishName string
	Quantity int
	Price    int
	Sum      int
}

type Receipt struct {
	BillID   int64
	OrderID  int64
	TableNum int
	Waiter   string
	ClosedAt *time.Time
	Lines    []ReceiptLine
	Total    int
}

// ReportFilter: From и To включительно, Paid == nil означает "все счета".
type ReportFilter struct {
	From time.Time
	To   time.Time
	Paid *bool
}

type Report struct {
	From       time.Time
	To         time.Time
	Bills      []models.Bill
	PaidCount  int
	OpenCount  int
	PaidAmount int
	OpenAmount int
}

type BillService interface {
	CalculateTotal(ctx context.Context, actor *models.User, orderID int64) (int, error)
	Create(ctx context.Context, actor *models.User, orderID int64) (*models.Bill, error)
	Close(ctx context.Context, actor *models.User, orderID int64) (*models.Bill, error)
	Receipt(ctx context.Context, actor *models.User, orderID int64) (*Receipt, error)
	Report(ctx context.Context, actor *models.User, filter ReportFilter) (*Report, error)
}

type Service interface {
	Auth() AuthService
	Users() UserService
	Menu() MenuService
	Orders() OrderService
	Bills() BillService
}
