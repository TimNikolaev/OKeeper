package presentation

import (
	"context"
)

// CreateAdminFunc создаёт учётную запись администратора (service.CreateAdmin).
type CreateAdminFunc func(ctx context.Context, username, fullName, password string) error

// Setup — первичная настройка: создание учётной записи администратора.
// Пустой логин пропускает настройку.
func (a *App) Setup(ctx context.Context, create CreateAdminFunc) error {
	a.clear()
	a.println(a.bold("OKeeper — первичная настройка"))
	a.println("Создайте учётную запись администратора. Через неё добавляются остальные пользователи и меню.")
	a.println(a.dim("Пустой логин — пропустить (позже: okeeper -setup)."))
	a.println()

	for {
		username, err := a.readLine("Логин администратора: ")
		if err != nil {
			return a.finish(err)
		}
		if username == "" {
			return nil
		}
		fullName, err := a.readLine("ФИО: ")
		if err != nil {
			return a.finish(err)
		}
		password, err := a.readSecret("Пароль (не короче 6 символов): ")
		if err != nil {
			return a.finish(err)
		}
		repeat, err := a.readSecret("Повторите пароль: ")
		if err != nil {
			return a.finish(err)
		}
		if password != repeat {
			a.println(a.red("Пароли не совпадают."))
			continue
		}

		if err := create(ctx, username, fullName, password); err != nil {
			if isFatal(err) {
				return a.finish(err)
			}
			a.println(a.red(errText(err)))
			continue
		}
		a.println(a.green("Администратор создан."))
		a.println()
		return nil
	}
}
