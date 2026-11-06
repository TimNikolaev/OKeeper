package presentation

import (
	"context"
	"errors"
	"strconv"

	"github.com/TimNikolaev/OKeeper/internal/models"
	"github.com/TimNikolaev/OKeeper/internal/service"
)

func (a *App) waiterMenu(ctx context.Context) error {
	return a.menu(a.header(), nil, []menuItem{
		{"Создать заказ", func() error { return a.createOrder(ctx) }},
		{"Заказы: просмотр и фильтр", func() error { return a.listOrders(ctx) }},
		{"Изменить статус заказа", func() error { return a.changeOrderStatus(ctx) }},
		{"Счёт и оплата", func() error { return a.billing(ctx) }},
		{"Печать чека", func() error { return a.printReceipt(ctx) }},
	}, "Выйти из учётной записи")
}

// createOrder — блок-схема «Формирование заказа»: столик → позиции (в цикле)
// → комментарий → передача на кухню.
func (a *App) createOrder(ctx context.Context) error {
	dishes, err := a.svc.Menu().ListAvailable(ctx, a.user)
	if err != nil {
		return err
	}
	if len(dishes) == 0 {
		a.setNotice(a.red("В меню нет доступных блюд."))
		return nil
	}

	table, ok, err := a.askInt("Номер столика (пусто — отмена): ", 1, noMax)
	if err != nil || !ok {
		return err
	}

	a.println()
	a.printDishes(dishes, false)
	a.println()

	available := make(map[int64]bool, len(dishes))
	names := make(map[int64]string, len(dishes))
	for _, d := range dishes {
		available[d.ID] = true
		names[d.ID] = d.Name
	}

	var items []service.OrderItemInput
	for {
		id, ok, err := a.askInt("ID блюда (пусто — отмена): ", 1, noMax)
		if err != nil {
			return err
		}
		if !ok {
			if len(items) == 0 {
				return nil
			}
			break
		}
		if !available[int64(id)] {
			a.println(a.red("Такого блюда нет в списке доступных."))
			continue
		}

		qty, ok, err := a.askInt("Количество (пусто — пропустить позицию): ", 1, noMax)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}

		comment, err := a.readLine("Комментарий к позиции (Enter — без комментария): ")
		if err != nil {
			return err
		}
		item := service.OrderItemInput{DishID: int64(id), Quantity: qty}
		if comment != "" {
			item.Comment = &comment
		}
		items = append(items, item)
		a.println(a.green("Позиция добавлена: " + dishName(names, int64(id)) + " × " + strconv.Itoa(qty)))

		more, err := a.confirm("Добавить ещё позицию?")
		if err != nil {
			return err
		}
		if !more {
			break
		}
	}

	input := service.CreateOrderInput{TableNum: table, Items: items}

	need, err := a.confirm("Нужен комментарий к заказу?")
	if err != nil {
		return err
	}
	if need {
		comment, err := a.readLine("Комментарий: ")
		if err != nil {
			return err
		}
		if comment != "" {
			input.Comment = &comment
		}
	}

	send, err := a.confirm("Передать заказ на кухню?")
	if err != nil {
		return err
	}
	if !send {
		a.setNotice("Заказ не создан.")
		return nil
	}

	order, err := a.svc.Orders().Create(ctx, a.user, input)
	if err != nil {
		return err
	}
	a.setNotice(a.green("Заказ № " + strconv.FormatInt(order.ID, 10) + " передан на кухню."))
	return nil
}

// listOrders — просмотр заказов с фильтрацией и проверкой статусов.
func (a *App) listOrders(ctx context.Context) error {
	filter := service.OrderFilter{}

	a.println("Статусы: 1 — новый, 2 — готовится, 3 — готов, 4 — подан, 5 — оплачен, 6 — отменён")
	n, ok, err := a.askInt("Фильтр по статусу (пусто — все): ", 1, len(orderStatuses))
	if err != nil {
		return err
	}
	if ok {
		filter.Statuses = []models.OrderStatus{orderStatuses[n-1]}
	}

	table, ok, err := a.askInt("Фильтр по столику (пусто — любой): ", 1, noMax)
	if err != nil {
		return err
	}
	if ok {
		filter.TableNum = &table
	}

	mine, err := a.confirm("Только мои заказы?")
	if err != nil {
		return err
	}
	if mine {
		filter.WaiterID = &a.user.ID
	}

	orders, err := a.svc.Orders().List(ctx, a.user, filter)
	if err != nil {
		return err
	}
	a.println()
	if len(orders) == 0 {
		a.println("Заказов по этому фильтру нет.")
		return a.pause()
	}
	a.printOrdersTable(orders)
	a.println()

	id, ok, err := a.askInt("Номер заказа для подробностей (пусто — назад): ", 1, noMax)
	if err != nil || !ok {
		return err
	}
	order, err := a.svc.Orders().GetByID(ctx, a.user, int64(id))
	if err != nil {
		return err
	}
	names, err := a.dishNames(ctx)
	if err != nil {
		return err
	}
	a.println()
	a.printOrder(*order, names)
	a.println()
	return a.pause()
}

// changeOrderStatus: официант отмечает заказ поданным или отменяет новый.
func (a *App) changeOrderStatus(ctx context.Context) error {
	id, ok, err := a.askInt("Номер заказа (пусто — отмена): ", 1, noMax)
	if err != nil || !ok {
		return err
	}
	order, err := a.svc.Orders().GetByID(ctx, a.user, int64(id))
	if err != nil {
		return err
	}
	names, err := a.dishNames(ctx)
	if err != nil {
		return err
	}
	a.println()
	a.printOrder(*order, names)
	a.println()

	a.println("1 — отметить поданным, 2 — отменить заказ")
	action, ok, err := a.askInt("Действие (пусто — отмена): ", 1, 2)
	if err != nil || !ok {
		return err
	}
	target := models.StatusServed
	if action == 2 {
		target = models.StatusCancelled
	}

	updated, err := a.svc.Orders().ChangeStatus(ctx, a.user, order.ID, target)
	if err != nil {
		return err
	}
	a.setNotice(a.green("Заказ № " + strconv.FormatInt(updated.ID, 10) + ": " + statusLabel(updated.Status) + "."))
	return nil
}

// billing — блок-схема «Процесс закрытия счёта»: выбрать заказ стола →
// составить счёт → оплата → закрыть счёт → чек.
func (a *App) billing(ctx context.Context) error {
	table, ok, err := a.askInt("Номер столика (пусто — отмена): ", 1, noMax)
	if err != nil || !ok {
		return err
	}

	orders, err := a.svc.Orders().List(ctx, a.user, service.OrderFilter{
		Statuses: []models.OrderStatus{models.StatusServed},
		TableNum: &table,
	})
	if err != nil {
		return err
	}
	if len(orders) == 0 {
		a.setNotice(a.red("У столика нет поданных заказов: счёт составляется после подачи."))
		return nil
	}

	order := orders[0]
	if len(orders) > 1 {
		a.println()
		a.printOrdersTable(orders)
		a.println()
		id, ok, err := a.askInt("Номер заказа (пусто — отмена): ", 1, noMax)
		if err != nil || !ok {
			return err
		}
		found := false
		for _, o := range orders {
			if o.ID == int64(id) {
				order, found = o, true
				break
			}
		}
		if !found {
			return service.ErrNotFound
		}
	}

	bill, err := a.svc.Bills().Create(ctx, a.user, order.ID)
	if err != nil {
		return err
	}
	names, err := a.dishNames(ctx)
	if err != nil {
		return err
	}
	a.println()
	a.printOrder(order, names)
	a.printf("  К оплате: %s\n\n", a.bold(money(bill.TotalAmount)))

	if bill.IsPaid == 1 {
		a.setNotice("Счёт уже оплачен.")
		return nil
	}

	paid, err := a.confirm("Гость оплатил?")
	if err != nil {
		return err
	}
	if !paid {
		a.setNotice("Счёт остаётся открытым.")
		return nil
	}

	if _, err := a.svc.Bills().Close(ctx, a.user, order.ID); err != nil {
		return err
	}
	a.println(a.green("Счёт закрыт."))

	wantReceipt, err := a.confirm("Распечатать чек?")
	if err != nil {
		return err
	}
	if !wantReceipt {
		a.setNotice(a.green("Счёт по заказу № " + strconv.FormatInt(order.ID, 10) + " закрыт."))
		return nil
	}
	return a.showReceipt(ctx, order.ID)
}

// printReceipt печатает чек по уже оплаченному заказу.
func (a *App) printReceipt(ctx context.Context) error {
	id, ok, err := a.askInt("Номер оплаченного заказа (пусто — отмена): ", 1, noMax)
	if err != nil || !ok {
		return err
	}
	return a.showReceipt(ctx, int64(id))
}

func (a *App) showReceipt(ctx context.Context, orderID int64) error {
	receipt, err := a.svc.Bills().Receipt(ctx, a.user, orderID)
	if err != nil {
		if errors.Is(err, service.ErrBillNotPaid) || errors.Is(err, service.ErrNotFound) {
			a.setNotice(a.red(errText(err)))
			return nil
		}
		return err
	}
	a.println()
	a.println(receipt.Text())
	return a.pause()
}
