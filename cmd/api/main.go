package main

import (
	"errors"
	"log"
	"os"

	"github.com/nongpal/Palge-Backend/internal/api"
	"github.com/nongpal/Palge-Backend/internal/configs"
)

func main() {
	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func run() (err error) {
	var cfg configs.Config
	configs.NewConfig(&cfg)

	app, err := api.NewApplication(cfg)
	if err != nil {
		return err
	}

	defer func() {
		if closeErr := app.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()

	return app.Run()
}
