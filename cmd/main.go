package main

import (
	"log"

	"github.com/TimNikolaev/OKeeper/internal/repository/sqlite"
)

func main() {
	repo, err := sqlite.NewRepository("./data/main.db")
	if err != nil {
		log.Fatal(err)
	}

	if err := repo.Init("./schema/schema.sql"); err != nil {
		log.Fatal(err)
	}

}
