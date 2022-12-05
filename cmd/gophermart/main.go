package main

import (
	"context"
	"log"
	"time"

	"github.com/IgorAleksandroff/gophermart/internal/app"
	"github.com/IgorAleksandroff/gophermart/internal/config"
)

func main() {
	log.Println("debug: start main")

	ctx, closeCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer closeCtx()

	cfg := config.GetConfig()

	app, err := app.NewApp(ctx, cfg)
	if err != nil {
		log.Fatalf("Create app error: %s", err)
	}
	defer app.Cancel()

	app.Run()
}
