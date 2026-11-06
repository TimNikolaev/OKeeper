package presentation

import (
	"context"
	"fmt"
	"strconv"

	"github.com/TimNikolaev/OKeeper/internal/models"
)

// dishNames возвращает названия всех блюд (включая недоступные) по ID.
func (a *App) dishNames(ctx context.Context) (map[int64]string, error) {
	dishes, err := a.svc.Menu().List(ctx, a.user)
	if err != nil {
		return nil, err
	}
	names := make(map[int64]string, len(dishes))
	for _, d := range dishes {
		names[d.ID] = d.Name
	}
	return names, nil
}

func dishName(names map[int64]string, id int64) string {
	if n, ok := names[id]; ok {
		return n
	}
	return fmt.Sprintf("блюдо #%d", id)
}

// printDishes печатает меню таблицей.
func (a *App) printDishes(dishes []models.Dish, showAvailability bool) {
	header := []string{"ID", "Категория", "Название", "Цена"}
	if showAvailability {
		header = append(header, "В наличии")
	}

	rows := make([][]string, 0, len(dishes))
	for _, d := range dishes {
		row := []string{strconv.FormatInt(d.ID, 10), d.Category, d.Name, money(d.Price)}
		if showAvailability {
			if d.IsAvailable == 1 {
				row = append(row, "да")
			} else {
				row = append(row, "нет")
			}
		}
		rows = append(rows, row)
	}
	a.table(header, rows)
}

// printOrder печатает заказ с позициями, комментариями и итогом.
func (a *App) printOrder(o models.Order, names map[int64]string) {
	a.printf("%s\n", a.bold(fmt.Sprintf("Заказ № %d · столик %d · %s · %s",
		o.ID, o.TableNum, statusLabel(o.Status), fmtTime(o.CreatedAt))))
	for _, it := range o.Items {
		line := fmt.Sprintf("  %s × %d — %s", dishName(names, it.DishID), it.Quantity, money(it.Price*it.Quantity))
		if it.Comment != nil {
			line += "  (" + *it.Comment + ")"
		}
		a.println(line)
	}
	if o.Comment != nil {
		a.printf("  Комментарий: %s\n", *o.Comment)
	}
	a.printf("  Итого: %s\n", money(orderTotal(o.Items)))
}

// printOrdersTable печатает список заказов одной таблицей.
func (a *App) printOrdersTable(orders []models.Order) {
	rows := make([][]string, 0, len(orders))
	for _, o := range orders {
		rows = append(rows, []string{
			strconv.FormatInt(o.ID, 10),
			strconv.Itoa(o.TableNum),
			statusLabel(o.Status),
			strconv.Itoa(len(o.Items)),
			money(orderTotal(o.Items)),
			fmtTime(o.CreatedAt),
		})
	}
	a.table([]string{"№", "Столик", "Статус", "Позиций", "Сумма", "Создан"}, rows)
}
