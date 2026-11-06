package service

import (
	"strings"

	"github.com/TimNikolaev/OKeeper/internal/models"
	"github.com/TimNikolaev/OKeeper/internal/repository"
)

type services struct {
	auth   *authService
	users  *userService
	menu   *menuService
	orders *orderService
	bills  *billService
}

// Option настраивает сервисный слой.
type Option func(*options)

type options struct {
	passwordIterations int
}

// WithPasswordIterations задаёт число итераций PBKDF2 для новых паролей.
// Нужен в основном тестам, чтобы не ждать полноценного хеширования;
// уже сохранённые пароли проверяются с теми итерациями, с которыми созданы.
func WithPasswordIterations(n int) Option {
	return func(o *options) { o.passwordIterations = n }
}

func applyOptions(opts []Option) options {
	o := options{passwordIterations: hashIterations}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

func New(repo repository.Repository, opts ...Option) Service {
	o := applyOptions(opts)
	return &services{
		auth:   &authService{users: repo.Users()},
		users:  &userService{users: repo.Users(), iterations: o.passwordIterations},
		menu:   &menuService{dishes: repo.Dishes()},
		orders: &orderService{orders: repo.Orders(), dishes: repo.Dishes()},
		bills: &billService{
			bills:  repo.Bills(),
			orders: repo.Orders(),
			dishes: repo.Dishes(),
			users:  repo.Users(),
		},
	}
}

func (s *services) Auth() AuthService    { return s.auth }
func (s *services) Users() UserService   { return s.users }
func (s *services) Menu() MenuService    { return s.menu }
func (s *services) Orders() OrderService { return s.orders }
func (s *services) Bills() BillService   { return s.bills }

// requireAuth проверяет, что пользователь авторизован.
func requireAuth(actor *models.User) error {
	if actor == nil {
		return ErrUnauthorized
	}
	return nil
}

// requireRole проверяет, что пользователь авторизован и имеет одну из ролей.
func requireRole(actor *models.User, roles ...models.Role) error {
	if err := requireAuth(actor); err != nil {
		return err
	}
	for _, r := range roles {
		if actor.Role == r {
			return nil
		}
	}
	return ErrForbidden
}

// normalizeComment обрезает пробелы, пустой комментарий превращает в nil.
func normalizeComment(c *string) *string {
	if c == nil {
		return nil
	}
	v := strings.TrimSpace(*c)
	if v == "" {
		return nil
	}
	return &v
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
