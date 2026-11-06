package service

import (
	"context"
	"strings"

	"github.com/TimNikolaev/OKeeper/internal/models"
	"github.com/TimNikolaev/OKeeper/internal/repository"
)

type menuService struct {
	dishes repository.DishRepository
}

func (s *menuService) List(ctx context.Context, actor *models.User) ([]models.Dish, error) {
	if err := requireAuth(actor); err != nil {
		return nil, err
	}
	return s.dishes.GetAll(ctx)
}

func (s *menuService) ListAvailable(ctx context.Context, actor *models.User) ([]models.Dish, error) {
	dishes, err := s.List(ctx, actor)
	if err != nil {
		return nil, err
	}

	available := make([]models.Dish, 0, len(dishes))
	for _, d := range dishes {
		if d.IsAvailable == 1 {
			available = append(available, d)
		}
	}
	return available, nil
}

func (s *menuService) Create(ctx context.Context, actor *models.User, in DishInput) (*models.Dish, error) {
	if err := requireRole(actor, models.RoleAdmin); err != nil {
		return nil, err
	}

	dish := &models.Dish{}
	if err := applyDishInput(dish, in); err != nil {
		return nil, err
	}

	if err := s.dishes.Create(ctx, dish); err != nil {
		return nil, err
	}
	return dish, nil
}

func (s *menuService) Update(ctx context.Context, actor *models.User, id int64, in DishInput) (*models.Dish, error) {
	if err := requireRole(actor, models.RoleAdmin); err != nil {
		return nil, err
	}

	dish, err := s.dishes.GetByID(ctx, id)
	if err != nil {
		return nil, wrapNotFound(err, "dish")
	}

	if err := applyDishInput(dish, in); err != nil {
		return nil, err
	}

	if err := s.dishes.Update(ctx, dish); err != nil {
		return nil, err
	}
	return dish, nil
}

func (s *menuService) SetAvailability(ctx context.Context, actor *models.User, id int64, available bool) error {
	if err := requireRole(actor, models.RoleAdmin); err != nil {
		return err
	}

	if _, err := s.dishes.GetByID(ctx, id); err != nil {
		return wrapNotFound(err, "dish")
	}

	return s.dishes.SetAvailability(ctx, id, boolToInt(available))
}

// applyDishInput проверяет данные блюда и переносит их в модель.
func applyDishInput(dish *models.Dish, in DishInput) error {
	category := strings.TrimSpace(in.Category)
	name := strings.TrimSpace(in.Name)
	if category == "" {
		return invalid("category is required")
	}
	if name == "" {
		return invalid("name is required")
	}
	if in.Price < 0 {
		return invalid("price must not be negative")
	}

	dish.Category = category
	dish.Name = name
	dish.Description = normalizeComment(in.Description)
	dish.Price = in.Price
	dish.IsAvailable = boolToInt(in.IsAvailable)
	return nil
}
