package main

import (
	"log"

	"github.com/IgorAleksandroff/gophermart.git/internal/app"
	"github.com/IgorAleksandroff/gophermart.git/internal/config"
)

func main() {
	app, err := app.NewApp(config.GetConfig())

	if err != nil {
		log.Fatalf("Create app error: %s", err)
	}

	app.Run()
}
