package service

import (
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrUnauthorized            = errors.New("authorization required")
	ErrForbidden               = errors.New("operation is not allowed for this role")
	ErrInvalidCredentials      = errors.New("invalid username or password")
	ErrInvalidInput            = errors.New("invalid input")
	ErrNotFound                = errors.New("not found")
	ErrUserExists              = errors.New("user already exists")
	ErrDishUnavailable         = errors.New("dish is not available")
	ErrInvalidStatusTransition = errors.New("invalid order status transition")
	ErrOrderNotServed          = errors.New("order must be served before billing")
	ErrBillAlreadyPaid         = errors.New("bill is already paid")
	ErrBillNotPaid             = errors.New("bill is not paid")
)

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidInput, fmt.Sprintf(format, args...))
}

func wrapNotFound(err error, what string) error {
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: %s", ErrNotFound, what)
	}
	return err
}
