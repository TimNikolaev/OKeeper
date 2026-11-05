package sqlite

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/TimNikolaev/OKeeper/internal/repository"
	_ "modernc.org/sqlite"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(initBasePath string) (*Repository, error) {
	db, err := sql.Open("sqlite", initBasePath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &Repository{db: db}, nil
}

func (r *Repository) Init(schemaPath string) error {
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", schemaPath, err)
	}

	if _, err := r.db.Exec(string(content)); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}

	return nil
}

func (r *Repository) Users() repository.UserRepository {
	return &userRepo{db: r.db}
}

func (r *Repository) Dishes() repository.DishRepository {
	return &dishRepo{db: r.db}
}

func (r *Repository) Orders() repository.OrderRepository {
	return &orderRepo{db: r.db}
}

func (r *Repository) Bills() repository.BillRepository {
	return &billRepo{db: r.db}
}
