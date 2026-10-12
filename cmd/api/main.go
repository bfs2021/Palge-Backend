package main

import (
	"log"

	"github.com/nongpal/Palge-Backend/internal/api"
	"github.com/nongpal/Palge-Backend/internal/configs"
)

func main() {
	var cfg configs.Config
	configs.NewConfig(&cfg)

	app, err := api.NewApplication(cfg)
	if err != nil {
		log.Fatal(err)
	}

	defer app.Close()

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
