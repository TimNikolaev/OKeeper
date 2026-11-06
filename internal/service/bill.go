package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/TimNikolaev/OKeeper/internal/models"
	"github.com/TimNikolaev/OKeeper/internal/repository"
)

type billService struct {
	bills  repository.BillRepository
	orders repository.OrderRepository
	dishes repository.DishRepository
	users  repository.UserRepository
}

func orderTotal(order *models.Order) int {
	total := 0
	for _, it := range order.Items {
		total += it.Price * it.Quantity
	}
	return total
}

// CalculateTotal считает сумму заказа по зафиксированным ценам позиций.
func (s *billService) CalculateTotal(ctx context.Context, actor *models.User, orderID int64) (int, error) {
	if err := requireRole(actor, models.RoleWaiter); err != nil {
		return 0, err
	}

	order, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return 0, wrapNotFound(err, "order")
	}
	return orderTotal(order), nil
}

// Create составляет счёт по поданному заказу. Если счёт уже составлен
// (в том числе открытый и неоплаченный), возвращается существующий.
func (s *billService) Create(ctx context.Context, actor *models.User, orderID int64) (*models.Bill, error) {
	if err := requireRole(actor, models.RoleWaiter); err != nil {
		return nil, err
	}

	order, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return nil, wrapNotFound(err, "order")
	}

	existing, err := s.bills.GetByOrderID(ctx, orderID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if order.Status != models.StatusServed {
		return nil, fmt.Errorf("%w: current status is %s", ErrOrderNotServed, order.Status)
	}

	bill := &models.Bill{
		OrderID:     orderID,
		TotalAmount: orderTotal(order),
		IsPaid:      0,
	}
	if err := s.bills.Create(ctx, bill); err != nil {
		return nil, err
	}
	return s.bills.GetByOrderID(ctx, orderID)
}

func (s *billService) Close(ctx context.Context, actor *models.User, orderID int64) (*models.Bill, error) {
	if err := requireRole(actor, models.RoleWaiter); err != nil {
		return nil, err
	}

	bill, err := s.bills.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, wrapNotFound(err, "bill")
	}
	if bill.IsPaid == 1 {
		return nil, ErrBillAlreadyPaid
	}

	if err := s.bills.MarkAsPaid(ctx, bill.ID); err != nil {
		return nil, err
	}
	if err := s.orders.UpdateStatus(ctx, orderID, models.StatusPaid); err != nil {
		return nil, fmt.Errorf("bill %d is paid, but order status update failed: %w", bill.ID, err)
	}

	return s.bills.GetByOrderID(ctx, orderID)
}

func (s *billService) Receipt(ctx context.Context, actor *models.User, orderID int64) (*Receipt, error) {
	if err := requireRole(actor, models.RoleWaiter); err != nil {
		return nil, err
	}

	bill, err := s.bills.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, wrapNotFound(err, "bill")
	}
	if bill.IsPaid != 1 {
		return nil, ErrBillNotPaid
	}

	order, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return nil, wrapNotFound(err, "order")
	}

	waiter, err := s.users.GetByID(ctx, order.WaiterID)
	if err != nil {
		return nil, wrapNotFound(err, "waiter")
	}

	lines := make([]ReceiptLine, 0, len(order.Items))
	for _, it := range order.Items {
		dish, err := s.dishes.GetByID(ctx, it.DishID)
		if err != nil {
			return nil, wrapNotFound(err, fmt.Sprintf("dish %d", it.DishID))
		}
		lines = append(lines, ReceiptLine{
			DishName: dish.Name,
			Quantity: it.Quantity,
			Price:    it.Price,
			Sum:      it.Price * it.Quantity,
		})
	}

	return &Receipt{
		BillID:   bill.ID,
		OrderID:  order.ID,
		TableNum: order.TableNum,
		Waiter:   waiter.FullName,
		ClosedAt: bill.ClosedAt,
		Lines:    lines,
		Total:    bill.TotalAmount,
	}, nil
}

func (r *Receipt) Text() string {
	var b strings.Builder
	line := strings.Repeat("-", 40)

	fmt.Fprintf(&b, "Чек № %d\n", r.BillID)
	fmt.Fprintf(&b, "Заказ № %d, столик %d\n", r.OrderID, r.TableNum)
	fmt.Fprintf(&b, "Официант: %s\n", r.Waiter)
	if r.ClosedAt != nil {
		fmt.Fprintf(&b, "Дата: %s\n", r.ClosedAt.Format("02.01.2006 15:04"))
	}
	b.WriteString(line + "\n")
	for _, l := range r.Lines {
		fmt.Fprintf(&b, "%s\n  %d x %d = %d\n", l.DishName, l.Quantity, l.Price, l.Sum)
	}
	b.WriteString(line + "\n")
	fmt.Fprintf(&b, "ИТОГО: %d\n", r.Total)
	return b.String()
}

func (s *billService) Report(ctx context.Context, actor *models.User, filter ReportFilter) (*Report, error) {
	if err := requireRole(actor, models.RoleAdmin); err != nil {
		return nil, err
	}

	if filter.From.IsZero() || filter.To.IsZero() {
		return nil, invalid("report period is required")
	}
	if filter.To.Before(filter.From) {
		return nil, invalid("end of period is before its start")
	}

	bills, err := s.bills.GetReports(ctx, filter.From, filter.To)
	if err != nil {
		return nil, err
	}

	report := &Report{From: filter.From, To: filter.To, Bills: make([]models.Bill, 0, len(bills))}
	for _, b := range bills {
		paid := b.IsPaid == 1
		if filter.Paid != nil && *filter.Paid != paid {
			continue
		}
		report.Bills = append(report.Bills, b)
		if paid {
			report.PaidCount++
			report.PaidAmount += b.TotalAmount
		} else {
			report.OpenCount++
			report.OpenAmount += b.TotalAmount
		}
	}
	return report, nil
}
