package main

import (
	"log"

	"github.com/IgorAleksandroff/gophermart/internal/app"
	"github.com/IgorAleksandroff/gophermart/internal/config"
)

func main() {
	log.Println("debug: start main")
	app, err := app.NewApp(config.GetConfig())
	defer app.Cancel()

	if err != nil {
		log.Fatalf("Create app error: %s", err)
	}

	app.Run()
}
