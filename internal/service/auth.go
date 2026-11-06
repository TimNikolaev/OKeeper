package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/TimNikolaev/OKeeper/internal/models"
	"github.com/TimNikolaev/OKeeper/internal/repository"
)

type authService struct {
	users repository.UserRepository
}

func (s *authService) Login(ctx context.Context, username, password string) (*models.User, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !verifyPassword(password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	user.PasswordHash = ""
	return user, nil
}
