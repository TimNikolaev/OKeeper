package presentation

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"
)

// errQuit сигнализирует штатный выход из программы.
var errQuit = errors.New("quit")

// isFatal — ошибки, которые прерывают работу, а не показываются пользователю.
func isFatal(err error) bool {
	return errors.Is(err, errQuit) ||
		errors.Is(err, io.EOF) ||
		errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded)
}

// ui — низкоуровневый ввод-вывод терминала: строки, меню, таблицы.
type ui struct {
	in     *bufio.Reader
	out    io.Writer
	ansi   bool
	hide   func() func() // отключает эхо ввода и возвращает функцию восстановления
	notice string        // сообщение, которое покажет следующий экран меню
}

func (u *ui) printf(format string, a ...any) { fmt.Fprintf(u.out, format, a...) }
func (u *ui) println(a ...any)               { fmt.Fprintln(u.out, a...) }

func (u *ui) style(code, s string) string {
	if !u.ansi {
		return s
	}
	return "\033[" + code + "m" + s + "\033[0m"
}

func (u *ui) bold(s string) string  { return u.style("1", s) }
func (u *ui) dim(s string) string   { return u.style("2", s) }
func (u *ui) red(s string) string   { return u.style("31", s) }
func (u *ui) green(s string) string { return u.style("32", s) }

func (u *ui) clear() {
	if u.ansi {
		u.printf("\033[H\033[2J")
		return
	}
	u.println()
}

// readRaw читает строку без обрезки пробелов.
func (u *ui) readRaw(prompt string) (string, error) {
	u.printf("%s", prompt)
	line, err := u.in.ReadString('\n')
	if err != nil && (err != io.EOF || line == "") {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func (u *ui) readLine(prompt string) (string, error) {
	s, err := u.readRaw(prompt)
	return strings.TrimSpace(s), err
}

// readSecret читает пароль, по возможности скрывая ввод.
func (u *ui) readSecret(prompt string) (string, error) {
	if u.hide == nil {
		return u.readRaw(prompt)
	}
	restore := u.hide()
	defer restore()
	s, err := u.readRaw(prompt)
	u.println()
	return s, err
}

// noMax передаётся в askInt, когда верхней границы нет.
const noMax = -1

// askInt запрашивает целое число от min до max (max == noMax — без верхней
// границы). Пустой ввод — отмена (ok == false).
func (u *ui) askInt(prompt string, min, max int) (n int, ok bool, err error) {
	for {
		s, err := u.readLine(prompt)
		if err != nil {
			return 0, false, err
		}
		if s == "" {
			return 0, false, nil
		}
		v, convErr := strconv.Atoi(s)
		if convErr != nil || v < min || (max != noMax && v > max) {
			if max != noMax {
				u.println(u.red(fmt.Sprintf("Введите число от %d до %d (пустая строка — отмена).", min, max)))
			} else {
				u.println(u.red(fmt.Sprintf("Введите целое число не меньше %d (пустая строка — отмена).", min)))
			}
			continue
		}
		return v, true, nil
	}
}

// askRequired запрашивает непустую строку. Пустой ввод — отмена.
func (u *ui) askRequired(prompt string) (string, bool, error) {
	s, err := u.readLine(prompt)
	if err != nil || s == "" {
		return "", false, err
	}
	return s, true, nil
}

// confirm задаёт вопрос «да/нет». Пустой ответ считается «нет».
func (u *ui) confirm(prompt string) (bool, error) {
	for {
		s, err := u.readLine(prompt + " (д/н): ")
		if err != nil {
			return false, err
		}
		switch strings.ToLower(s) {
		case "д", "да", "y", "yes":
			return true, nil
		case "", "н", "нет", "n", "no":
			return false, nil
		}
		u.println(u.red("Ответьте «д» или «н»."))
	}
}

func (u *ui) pause() error {
	_, err := u.readLine(u.dim("Нажмите Enter, чтобы продолжить…"))
	return err
}

func (u *ui) setNotice(s string) { u.notice = s }

type menuItem struct {
	label string
	run   func() error
}

// menu показывает пронумерованное меню и выполняет выбранный пункт, пока
// пользователь не выберет «0». before выводит данные над меню на каждом
// проходе (например, очередь заказов). Ошибки действий показываются на
// следующем экране, фатальные — возвращаются.
func (u *ui) menu(title string, before func() error, items []menuItem, back string) error {
	for {
		u.clear()
		u.println(u.bold(title))
		u.println()
		if u.notice != "" {
			u.println(u.notice)
			u.println()
			u.notice = ""
		}
		if before != nil {
			if err := before(); err != nil {
				if isFatal(err) {
					return err
				}
				u.println(u.red(errText(err)))
			}
			u.println()
		}
		for i, it := range items {
			u.printf("  %d. %s\n", i+1, it.label)
		}
		u.printf("  0. %s\n\n", back)

		choice, err := u.readLine("Выбор: ")
		if err != nil {
			return err
		}
		if choice == "0" {
			return nil
		}
		n, convErr := strconv.Atoi(choice)
		if convErr != nil || n < 1 || n > len(items) {
			u.notice = u.red("Нет такого пункта меню.")
			continue
		}
		if err := items[n-1].run(); err != nil {
			if isFatal(err) {
				return err
			}
			u.notice = u.red(errText(err))
		}
	}
}

// table печатает выровненную таблицу. Цвет внутри ячеек не используется,
// чтобы не ломать выравнивание.
func (u *ui) table(header []string, rows [][]string) {
	tw := tabwriter.NewWriter(u.out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, strings.Join(header, "\t"))
	for _, r := range rows {
		fmt.Fprintln(tw, strings.Join(r, "\t"))
	}
	tw.Flush()
}
