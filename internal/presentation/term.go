package presentation

import (
	"os"
	"os/exec"
)

// StdinEchoOn возвращает эхо ввода. Вызывается при прерывании программы
// по Ctrl+C, чтобы терминал не остался со скрытым вводом.
func StdinEchoOn() {
	on := exec.Command("stty", "echo")
	on.Stdin = os.Stdin
	_ = on.Run()
}

// StdinEchoOff отключает эхо вводимых символов через stty (Linux, macOS)
// и возвращает функцию, которая его восстанавливает. Если stty недоступен
// (например, в Windows), ввод остаётся видимым, а возвращаемая функция
// ничего не делает.
func StdinEchoOff() func() {
	off := exec.Command("stty", "-echo")
	off.Stdin = os.Stdin
	if err := off.Run(); err != nil {
		return func() {}
	}
	return func() {
		on := exec.Command("stty", "echo")
		on.Stdin = os.Stdin
		_ = on.Run()
	}
}
