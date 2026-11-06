package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/TimNikolaev/OKeeper/internal/presentation"
	"github.com/TimNikolaev/OKeeper/internal/repository/sqlite"
	"github.com/TimNikolaev/OKeeper/internal/service"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "okeeper:", err)
		os.Exit(1)
	}
}

func run() error {
	dbPath := flag.String("db", "./data/main.db", "путь к файлу базы данных SQLite")
	schemaPath := flag.String("schema", "./schema/schema.sql", "путь к SQL-схеме")
	setup := flag.Bool("setup", false, "создать учётную запись администратора перед входом")
	noColor := flag.Bool("no-color", false, "отключить цвета и очистку экрана")
	flag.Parse()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigs
		presentation.StdinEchoOn()
		fmt.Fprintln(os.Stderr)
		os.Exit(130)
	}()

	info, statErr := os.Stat(*dbPath)
	fresh := errors.Is(statErr, os.ErrNotExist) || (statErr == nil && info.Size() == 0)

	if err := os.MkdirAll(filepath.Dir(*dbPath), 0o755); err != nil {
		return fmt.Errorf("создать каталог базы данных: %w", err)
	}

	repo, err := sqlite.NewRepository(dsn(*dbPath))
	if err != nil {
		return fmt.Errorf("открыть базу %s: %w", *dbPath, err)
	}
	if err := repo.Init(*schemaPath); err != nil {
		return fmt.Errorf("%w (запускайте из корня проекта или укажите -schema)", err)
	}

	svc := service.New(repo)

	opts := []presentation.Option{
		presentation.WithANSI(colorEnabled(*noColor)),
		presentation.WithSecretHider(presentation.StdinEchoOff),
	}
	app := presentation.New(svc, os.Stdin, os.Stdout, opts...)

	ctx := context.Background()

	if fresh || *setup {
		err := app.Setup(ctx, func(ctx context.Context, username, fullName, password string) error {
			_, err := service.CreateAdmin(ctx, repo, username, fullName, password)
			return err
		})
		if err != nil {
			return err
		}
	}

	return app.Run(ctx)
}

func dsn(path string) string {
	return path + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_time_format=sqlite"
}

func colorEnabled(disabled bool) bool {
	if disabled || runtime.GOOS == "windows" || os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	fi, err := os.Stdout.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}
