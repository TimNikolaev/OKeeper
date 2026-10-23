package main

import (
	"log"

	"github.com/TimNikolaev/OKeeper/internal/repository"
)

func main() {
	repo, err := repository.New("./data/main.db")
	if err != nil {
		log.Fatal(err)
	}

	if err := repo.Init("./schema/schema.sql"); err != nil {
		log.Fatal(err)
	}

}
