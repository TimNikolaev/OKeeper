package service

import (
	"context"
	"fmt"

	"github.com/TimNikolaev/OKeeper/internal/models"
	"github.com/TimNikolaev/OKeeper/internal/repository"
)

type orderService struct {
	orders repository.OrderRepository
	dishes repository.DishRepository
}

var allStatuses = []models.OrderStatus{
	models.StatusNew,
	models.StatusCooking,
	models.StatusReady,
	models.StatusServed,
	models.StatusPaid,
	models.StatusCancelled,
}

// Допустимые переходы статусов по ролям (блок-схемы «Обработка заказа поваром»
// и «Процесс закрытия счёта»). Статус paid выставляется только при закрытии
// счёта в BillService и вручную не меняется.
var statusTransitions = map[models.Role]map[models.OrderStatus][]models.OrderStatus{
	models.RoleCook: {
		models.StatusNew:     {models.StatusCooking},
		models.StatusCooking: {models.StatusReady},
	},
	models.RoleWaiter: {
		models.StatusReady: {models.StatusServed},
		models.StatusNew:   {models.StatusCancelled},
	},
}

func validStatus(s models.OrderStatus) bool {
	for _, v := range allStatuses {
		if v == s {
			return true
		}
	}
	return false
}

// Create формирует заказ и передаёт его на кухню: заказ создаётся со статусом
// new и сразу попадает в очередь повара. Цены позиций фиксируются на момент
// создания заказа.
func (s *orderService) Create(ctx context.Context, actor *models.User, in CreateOrderInput) (*models.Order, error) {
	if err := requireRole(actor, models.RoleWaiter); err != nil {
		return nil, err
	}

	if in.TableNum <= 0 {
		return nil, invalid("table number must be positive")
	}
	if len(in.Items) == 0 {
		return nil, invalid("order must contain at least one item")
	}

	items := make([]models.OrderDish, 0, len(in.Items))
	for _, it := range in.Items {
		if it.Quantity <= 0 {
			return nil, invalid("quantity must be positive")
		}

		dish, err := s.dishes.GetByID(ctx, it.DishID)
		if err != nil {
			return nil, wrapNotFound(err, fmt.Sprintf("dish %d", it.DishID))
		}
		if dish.IsAvailable != 1 {
			return nil, fmt.Errorf("%w: %s", ErrDishUnavailable, dish.Name)
		}

		items = append(items, models.OrderDish{
			DishID:   dish.ID,
			Quantity: it.Quantity,
			Price:    dish.Price,
			Comment:  normalizeComment(it.Comment),
		})
	}

	order := &models.Order{
		TableNum: in.TableNum,
		WaiterID: actor.ID,
		Status:   models.StatusNew,
		Comment:  normalizeComment(in.Comment),
	}
	if err := s.orders.Create(ctx, order, items); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *orderService) GetByID(ctx context.Context, actor *models.User, id int64) (*models.Order, error) {
	if err := requireAuth(actor); err != nil {
		return nil, err
	}

	order, err := s.orders.GetByID(ctx, id)
	if err != nil {
		return nil, wrapNotFound(err, "order")
	}
	return order, nil
}

// List возвращает заказы с фильтрацией по статусам, столику и официанту.
// Без фильтра по статусам возвращаются заказы во всех статусах.
func (s *orderService) List(ctx context.Context, actor *models.User, filter OrderFilter) ([]models.Order, error) {
	if err := requireRole(actor, models.RoleWaiter, models.RoleAdmin); err != nil {
		return nil, err
	}

	statuses := filter.Statuses
	if len(statuses) == 0 {
		statuses = allStatuses
	}
	for _, st := range statuses {
		if !validStatus(st) {
			return nil, invalid("unknown status %q", st)
		}
	}

	orders, err := s.orders.GetQueue(ctx, statuses)
	if err != nil {
		return nil, err
	}

	result := make([]models.Order, 0, len(orders))
	for _, o := range orders {
		if filter.TableNum != nil && o.TableNum != *filter.TableNum {
			continue
		}
		if filter.WaiterID != nil && o.WaiterID != *filter.WaiterID {
			continue
		}
		result = append(result, o)
	}

	return s.withItems(ctx, result)
}

// CookQueue возвращает очередь повара: новые заказы и заказы в приготовлении,
// от старых к новым, вместе с позициями.
func (s *orderService) CookQueue(ctx context.Context, actor *models.User) ([]models.Order, error) {
	if err := requireRole(actor, models.RoleCook); err != nil {
		return nil, err
	}

	orders, err := s.orders.GetQueue(ctx, []models.OrderStatus{models.StatusNew, models.StatusCooking})
	if err != nil {
		return nil, err
	}
	return s.withItems(ctx, orders)
}

// ChangeStatus меняет статус заказа с проверкой роли и допустимости перехода.
func (s *orderService) ChangeStatus(ctx context.Context, actor *models.User, orderID int64, status models.OrderStatus) (*models.Order, error) {
	if err := requireRole(actor, models.RoleCook, models.RoleWaiter); err != nil {
		return nil, err
	}

	order, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return nil, wrapNotFound(err, "order")
	}

	allowed := false
	for _, to := range statusTransitions[actor.Role][order.Status] {
		if to == status {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, fmt.Errorf("%w: %s -> %s", ErrInvalidStatusTransition, order.Status, status)
	}

	if err := s.orders.UpdateStatus(ctx, orderID, status); err != nil {
		return nil, err
	}
	order.Status = status
	return order, nil
}

// withItems дозагружает позиции: GetQueue их не возвращает.
func (s *orderService) withItems(ctx context.Context, orders []models.Order) ([]models.Order, error) {
	for i := range orders {
		full, err := s.orders.GetByID(ctx, orders[i].ID)
		if err != nil {
			return nil, wrapNotFound(err, "order")
		}
		orders[i].Items = full.Items
	}
	return orders, nil
}
