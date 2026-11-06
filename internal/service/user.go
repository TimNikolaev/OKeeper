package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/TimNikolaev/OKeeper/internal/models"
	"github.com/TimNikolaev/OKeeper/internal/repository"
)

type userService struct {
	users      repository.UserRepository
	iterations int
}

func validRole(r models.Role) bool {
	switch r {
	case models.RoleWaiter, models.RoleCook, models.RoleAdmin:
		return true
	}
	return false
}

func (s *userService) Create(ctx context.Context, actor *models.User, in CreateUserInput) (*models.User, error) {
	if err := requireRole(actor, models.RoleAdmin); err != nil {
		return nil, err
	}
	return s.create(ctx, in)
}

// CreateAdmin создаёт администратора без авторизованного пользователя.
// Нужен для первого запуска: в пустой базе некому создавать учётные записи.
func CreateAdmin(ctx context.Context, repo repository.Repository, username, fullName, password string, opts ...Option) (*models.User, error) {
	s := &userService{users: repo.Users(), iterations: applyOptions(opts).passwordIterations}
	return s.create(ctx, CreateUserInput{
		Username: username,
		Password: password,
		FullName: fullName,
		Role:     models.RoleAdmin,
	})
}

func (s *userService) create(ctx context.Context, in CreateUserInput) (*models.User, error) {
	username := strings.TrimSpace(in.Username)
	fullName := strings.TrimSpace(in.FullName)
	if username == "" {
		return nil, invalid("username is required")
	}
	if fullName == "" {
		return nil, invalid("full name is required")
	}
	if len(in.Password) < minPasswordLen {
		return nil, invalid("password must be at least %d characters", minPasswordLen)
	}
	if !validRole(in.Role) {
		return nil, invalid("unknown role %q", in.Role)
	}

	_, err := s.users.GetByUsername(ctx, username)
	if err == nil {
		return nil, ErrUserExists
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	hash, err := hashPasswordIter(in.Password, s.iterations)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:     username,
		PasswordHash: hash,
		FullName:     fullName,
		Role:         in.Role,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *userService) GetByID(ctx context.Context, actor *models.User, id int64) (*models.User, error) {
	if err := requireRole(actor, models.RoleAdmin); err != nil {
		return nil, err
	}

	user, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, wrapNotFound(err, "user")
	}

	user.PasswordHash = ""
	return user, nil
}
