package models

import "time"

type Role string

const (
	RoleWaiter Role = "waiter"
	RoleCook   Role = "cook"
	RoleAdmin  Role = "admin"
)

type OrderStatus string

const (
	StatusNew       OrderStatus = "new"
	StatusCooking   OrderStatus = "cooking"
	StatusReady     OrderStatus = "ready"
	StatusServed    OrderStatus = "served"
	StatusPaid      OrderStatus = "paid"
	StatusCancelled OrderStatus = "cancelled"
)

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	FullName     string
	Role         Role
	CreatedAt    time.Time
}

type Dish struct {
	ID          int64
	Category    string
	Name        string
	Description *string
	Price       int
	IsAvailable int
	CreatedAt   time.Time
}

type Order struct {
	ID        int64
	TableNum  int
	WaiterID  int64
	Status    OrderStatus
	Comment   *string
	CreatedAt time.Time
	Items     []OrderDish
}

type OrderDish struct {
	ID       int64
	OrderID  int64
	DishID   int64
	Quantity int
	Price    int
	Comment  *string
}

type Bill struct {
	ID          int64
	OrderID     int64
	TotalAmount int
	IsPaid      int
	CreatedAt   time.Time
	ClosedAt    *time.Time
}
