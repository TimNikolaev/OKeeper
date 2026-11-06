package presentation

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/TimNikolaev/OKeeper/internal/models"
	"github.com/TimNikolaev/OKeeper/internal/service"
)

// App — TUI-приложение: экран входа и меню по ролям поверх сервисного слоя.
type App struct {
	ui
	svc  service.Service
	user *models.User
}

type Option func(*App)

// WithANSI включает цвета и очистку экрана.
func WithANSI(on bool) Option { return func(a *App) { a.ansi = on } }

// WithSecretHider задаёт функцию, скрывающую ввод пароля (см. StdinEchoOff).
func WithSecretHider(hide func() func()) Option { return func(a *App) { a.hide = hide } }

func New(svc service.Service, in io.Reader, out io.Writer, opts ...Option) *App {
	a := &App{ui: ui{in: bufio.NewReader(in), out: out}, svc: svc}
	for _, o := range opts {
		o(a)
	}
	return a
}

// Run крутит цикл «вход → меню роли → выход из учётной записи» до выхода
// из программы или конца ввода.
func (a *App) Run(ctx context.Context) error {
	for {
		if err := a.login(ctx); err != nil {
			return a.finish(err)
		}
		err := a.roleMenu(ctx)
		a.user = nil
		if err != nil {
			return a.finish(err)
		}
	}
}

// finish: выход пользователя и конец ввода — штатное завершение.
func (a *App) finish(err error) error {
	if errors.Is(err, errQuit) || errors.Is(err, io.EOF) {
		a.println()
		return nil
	}
	return err
}

func (a *App) header() string {
	return fmt.Sprintf("OKeeper · %s · %s", a.user.FullName, roleLabel(a.user.Role))
}

func (a *App) roleMenu(ctx context.Context) error {
	switch a.user.Role {
	case models.RoleWaiter:
		return a.waiterMenu(ctx)
	case models.RoleCook:
		return a.cookMenu(ctx)
	case models.RoleAdmin:
		return a.adminMenu(ctx)
	}
	a.println(a.red("Для этой роли нет интерфейса."))
	return nil
}

// login — блок-схема «Авторизация»: при неудаче показываем ошибку и
// возвращаемся к вводу. Пустой логин завершает программу.
func (a *App) login(ctx context.Context) error {
	a.clear()
	a.println(a.bold("OKeeper — учёт заказов в кафе"))
	a.println(a.dim("Пустой логин — выход из программы."))
	a.println()

	for {
		username, err := a.readLine("Логин: ")
		if err != nil {
			return err
		}
		if username == "" {
			return errQuit
		}
		password, err := a.readSecret("Пароль: ")
		if err != nil {
			return err
		}

		user, err := a.svc.Auth().Login(ctx, username, password)
		if err != nil {
			if isFatal(err) {
				return err
			}
			a.println(a.red(errText(err)))
			continue
		}
		a.user = user
		return nil
	}
}
