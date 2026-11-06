package presentation

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/TimNikolaev/OKeeper/internal/models"
	"github.com/TimNikolaev/OKeeper/internal/service"
)

// currency — символ валюты в суммах. Цены в БД хранятся целыми числами.
const currency = "₽"

const dateLayout = "02.01.2006"

// money форматирует сумму с разделением разрядов: 1 200 ₽.
func money(v int) string {
	s := strconv.Itoa(v)
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
	}
	out := b.String()
	if neg {
		out = "-" + out
	}
	return out + " " + currency
}

func fmtTime(t time.Time) string { return t.Local().Format("02.01 15:04") }

func fmtTimePtr(t *time.Time) string {
	if t == nil {
		return "—"
	}
	return fmtTime(*t)
}

// orderStatuses — статусы в порядке, в котором они показываются в фильтре.
var orderStatuses = []models.OrderStatus{
	models.StatusNew,
	models.StatusCooking,
	models.StatusReady,
	models.StatusServed,
	models.StatusPaid,
	models.StatusCancelled,
}

func statusLabel(s models.OrderStatus) string {
	switch s {
	case models.StatusNew:
		return "Новый"
	case models.StatusCooking:
		return "Готовится"
	case models.StatusReady:
		return "Готов"
	case models.StatusServed:
		return "Подан"
	case models.StatusPaid:
		return "Оплачен"
	case models.StatusCancelled:
		return "Отменён"
	}
	return string(s)
}

func roleLabel(r models.Role) string {
	switch r {
	case models.RoleWaiter:
		return "Официант"
	case models.RoleCook:
		return "Повар"
	case models.RoleAdmin:
		return "Администратор"
	}
	return string(r)
}

func orderTotal(items []models.OrderDish) int {
	total := 0
	for _, it := range items {
		total += it.Price * it.Quantity
	}
	return total
}

// parseDate разбирает дату ДД.ММ.ГГГГ в локальном часовом поясе.
func parseDate(s string) (time.Time, error) {
	return time.ParseInLocation(dateLayout, s, time.Local)
}

// entityRU переводит названия сущностей из сообщений сервисного слоя.
var entityRU = strings.NewReplacer(
	"order", "заказ",
	"dish", "блюдо",
	"bill", "счёт",
	"waiter", "официант",
	"user", "пользователь",
)

// errText переводит ошибки сервисного слоя в понятные сообщения.
func errText(err error) string {
	detail := func(prefix string) string {
		d := strings.TrimPrefix(err.Error(), prefix)
		if d == "" || d == err.Error() {
			return ""
		}
		return " (" + d + ")"
	}

	switch {
	case errors.Is(err, service.ErrInvalidCredentials):
		return "Неверный логин или пароль."
	case errors.Is(err, service.ErrUnauthorized):
		return "Требуется авторизация."
	case errors.Is(err, service.ErrForbidden):
		return "Недостаточно прав для этой операции."
	case errors.Is(err, service.ErrNotFound):
		return "Не найдено" + entityRU.Replace(detail("not found: ")) + "."
	case errors.Is(err, service.ErrInvalidInput):
		return "Некорректные данные" + detail("invalid input: ") + "."
	case errors.Is(err, service.ErrUserExists):
		return "Пользователь с таким логином уже существует."
	case errors.Is(err, service.ErrDishUnavailable):
		return "Блюдо недоступно" + detail("dish is not available: ") + "."
	case errors.Is(err, service.ErrInvalidStatusTransition):
		return "Недопустимая смена статуса" + detail("invalid order status transition: ") + "."
	case errors.Is(err, service.ErrOrderNotServed):
		return "Счёт можно составить только по поданному заказу."
	case errors.Is(err, service.ErrBillAlreadyPaid):
		return "Счёт уже оплачен."
	case errors.Is(err, service.ErrBillNotPaid):
		return "Счёт ещё не оплачен, чек недоступен."
	}
	return fmt.Sprintf("Ошибка: %v", err)
}
