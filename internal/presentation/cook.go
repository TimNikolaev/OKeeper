package presentation

import (
	"context"
	"strconv"

	"github.com/TimNikolaev/OKeeper/internal/models"
)

// cookMenu — блок-схема «Обработка заказа поваром»: очередь показывается
// над меню и обновляется на каждом проходе.
func (a *App) cookMenu(ctx context.Context) error {
	return a.menu(a.header(), func() error { return a.showQueue(ctx) }, []menuItem{
		{"Обновить очередь", func() error { return nil }},
		{"Начать приготовление", func() error { return a.moveOrder(ctx, models.StatusCooking) }},
		{"Заказ готов", func() error { return a.moveOrder(ctx, models.StatusReady) }},
	}, "Выйти из учётной записи")
}

func (a *App) showQueue(ctx context.Context) error {
	orders, err := a.svc.Orders().CookQueue(ctx, a.user)
	if err != nil {
		return err
	}
	if len(orders) == 0 {
		a.println(a.dim("Очередь пуста — заказов нет."))
		return nil
	}

	names, err := a.dishNames(ctx)
	if err != nil {
		return err
	}
	a.println(a.bold("Очередь заказов:"))
	a.println()
	for _, o := range orders {
		a.printOrder(o, names)
		a.println()
	}
	return nil
}

func (a *App) moveOrder(ctx context.Context, status models.OrderStatus) error {
	id, ok, err := a.askInt("Номер заказа (пусто — отмена): ", 1, noMax)
	if err != nil || !ok {
		return err
	}
	order, err := a.svc.Orders().ChangeStatus(ctx, a.user, int64(id), status)
	if err != nil {
		return err
	}
	a.setNotice(a.green("Заказ № " + strconv.FormatInt(order.ID, 10) + ": " + statusLabel(order.Status) + "."))
	return nil
}
