package presentation

import (
	"context"
	"strconv"
	"time"

	"github.com/TimNikolaev/OKeeper/internal/models"
	"github.com/TimNikolaev/OKeeper/internal/service"
)

func (a *App) adminMenu(ctx context.Context) error {
	return a.menu(a.header(), nil, []menuItem{
		{"Редактирование меню", func() error { return a.menuEditor(ctx) }},
		{"Создать пользователя", func() error { return a.createUser(ctx) }},
		{"Отчёт по счетам", func() error { return a.report(ctx) }},
	}, "Выйти из учётной записи")
}

func (a *App) menuEditor(ctx context.Context) error {
	before := func() error {
		dishes, err := a.svc.Menu().List(ctx, a.user)
		if err != nil {
			return err
		}
		if len(dishes) == 0 {
			a.println(a.dim("Меню пока пустое."))
			return nil
		}
		a.printDishes(dishes, true)
		return nil
	}
	return a.menu("Редактирование меню", before, []menuItem{
		{"Добавить блюдо", func() error { return a.addDish(ctx) }},
		{"Изменить блюдо", func() error { return a.editDish(ctx) }},
		{"Включить/выключить блюдо", func() error { return a.toggleDish(ctx) }},
	}, "Назад")
}

func (a *App) addDish(ctx context.Context) error {
	category, ok, err := a.askRequired("Категория (пусто — отмена): ")
	if err != nil || !ok {
		return err
	}
	name, ok, err := a.askRequired("Название (пусто — отмена): ")
	if err != nil || !ok {
		return err
	}
	description, err := a.readLine("Описание (Enter — без описания): ")
	if err != nil {
		return err
	}
	price, ok, err := a.askInt("Цена (пусто — отмена): ", 0, noMax)
	if err != nil || !ok {
		return err
	}

	in := service.DishInput{Category: category, Name: name, Price: price, IsAvailable: true}
	if description != "" {
		in.Description = &description
	}
	dish, err := a.svc.Menu().Create(ctx, a.user, in)
	if err != nil {
		return err
	}
	a.setNotice(a.green("Блюдо добавлено: " + dish.Name + " (ID " + strconv.FormatInt(dish.ID, 10) + ")."))
	return nil
}

// findDish ищет блюдо в полном меню, включая недоступные.
func (a *App) findDish(ctx context.Context, id int64) (*models.Dish, error) {
	dishes, err := a.svc.Menu().List(ctx, a.user)
	if err != nil {
		return nil, err
	}
	for i := range dishes {
		if dishes[i].ID == id {
			return &dishes[i], nil
		}
	}
	return nil, service.ErrNotFound
}

// editDish: пустой ввод в поле оставляет текущее значение.
func (a *App) editDish(ctx context.Context) error {
	id, ok, err := a.askInt("ID блюда (пусто — отмена): ", 1, noMax)
	if err != nil || !ok {
		return err
	}
	dish, err := a.findDish(ctx, int64(id))
	if err != nil {
		return err
	}

	in := service.DishInput{
		Category:    dish.Category,
		Name:        dish.Name,
		Description: dish.Description,
		Price:       dish.Price,
		IsAvailable: dish.IsAvailable == 1,
	}

	a.println(a.dim("Enter — оставить текущее значение."))
	if s, err := a.readLine("Категория [" + dish.Category + "]: "); err != nil {
		return err
	} else if s != "" {
		in.Category = s
	}
	if s, err := a.readLine("Название [" + dish.Name + "]: "); err != nil {
		return err
	} else if s != "" {
		in.Name = s
	}
	if s, err := a.readLine("Описание: "); err != nil {
		return err
	} else if s != "" {
		in.Description = &s
	}
	price, ok, err := a.askInt("Цена ["+strconv.Itoa(dish.Price)+"]: ", 0, noMax)
	if err != nil {
		return err
	}
	if ok {
		in.Price = price
	}

	updated, err := a.svc.Menu().Update(ctx, a.user, dish.ID, in)
	if err != nil {
		return err
	}
	a.setNotice(a.green("Блюдо обновлено: " + updated.Name + "."))
	return nil
}

func (a *App) toggleDish(ctx context.Context) error {
	id, ok, err := a.askInt("ID блюда (пусто — отмена): ", 1, noMax)
	if err != nil || !ok {
		return err
	}
	dish, err := a.findDish(ctx, int64(id))
	if err != nil {
		return err
	}

	enable := dish.IsAvailable != 1
	if err := a.svc.Menu().SetAvailability(ctx, a.user, dish.ID, enable); err != nil {
		return err
	}
	if enable {
		a.setNotice(a.green(dish.Name + ": теперь в наличии."))
	} else {
		a.setNotice(a.green(dish.Name + ": снято с продажи."))
	}
	return nil
}

func (a *App) createUser(ctx context.Context) error {
	username, ok, err := a.askRequired("Логин (пусто — отмена): ")
	if err != nil || !ok {
		return err
	}
	fullName, ok, err := a.askRequired("ФИО (пусто — отмена): ")
	if err != nil || !ok {
		return err
	}
	password, err := a.readSecret("Пароль (не короче 6 символов): ")
	if err != nil {
		return err
	}

	a.println("Роль: 1 — официант, 2 — повар, 3 — администратор")
	n, ok, err := a.askInt("Роль (пусто — отмена): ", 1, 3)
	if err != nil || !ok {
		return err
	}
	roles := []models.Role{models.RoleWaiter, models.RoleCook, models.RoleAdmin}

	user, err := a.svc.Users().Create(ctx, a.user, service.CreateUserInput{
		Username: username,
		Password: password,
		FullName: fullName,
		Role:     roles[n-1],
	})
	if err != nil {
		return err
	}
	a.setNotice(a.green("Пользователь создан: " + user.Username + " (" + roleLabel(user.Role) + ")."))
	return nil
}

// report — отчёт по счетам за период с фильтром по оплате.
func (a *App) report(ctx context.Context) error {
	today := time.Now()

	from, err := a.askDate("Начало периода ДД.ММ.ГГГГ (Enter — сегодня): ", today)
	if err != nil {
		return err
	}
	to, err := a.askDate("Конец периода ДД.ММ.ГГГГ (Enter — сегодня): ", today)
	if err != nil {
		return err
	}

	a.println("Счета: 1 — все, 2 — оплаченные, 3 — открытые")
	choice, ok, err := a.askInt("Фильтр (Enter — все): ", 1, 3)
	if err != nil {
		return err
	}
	filter := service.ReportFilter{
		From: time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.Local).UTC(),
		To:   time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 0, time.Local).UTC(),
	}
	if ok && choice != 1 {
		paid := choice == 2
		filter.Paid = &paid
	}

	rep, err := a.svc.Bills().Report(ctx, a.user, filter)
	if err != nil {
		return err
	}

	a.println()
	a.printf("%s\n", a.bold("Отчёт по счетам: "+from.Format(dateLayout)+" — "+to.Format(dateLayout)))
	if len(rep.Bills) == 0 {
		a.println("Счетов за этот период нет.")
		return a.pause()
	}

	rows := make([][]string, 0, len(rep.Bills))
	for _, b := range rep.Bills {
		state := "открыт"
		if b.IsPaid == 1 {
			state = "оплачен"
		}
		rows = append(rows, []string{
			strconv.FormatInt(b.ID, 10),
			strconv.FormatInt(b.OrderID, 10),
			money(b.TotalAmount),
			state,
			fmtTime(b.CreatedAt),
			fmtTimePtr(b.ClosedAt),
		})
	}
	a.table([]string{"Счёт", "Заказ", "Сумма", "Статус", "Создан", "Закрыт"}, rows)
	a.println()
	a.printf("Оплачено: %d на %s\n", rep.PaidCount, money(rep.PaidAmount))
	a.printf("Открыто:  %d на %s\n\n", rep.OpenCount, money(rep.OpenAmount))
	return a.pause()
}

// askDate запрашивает дату; пустой ввод возвращает def.
func (a *App) askDate(prompt string, def time.Time) (time.Time, error) {
	for {
		s, err := a.readLine(prompt)
		if err != nil {
			return time.Time{}, err
		}
		if s == "" {
			return def, nil
		}
		t, parseErr := parseDate(s)
		if parseErr != nil {
			a.println(a.red("Неверный формат даты, нужно ДД.ММ.ГГГГ."))
			continue
		}
		return t, nil
	}
}
