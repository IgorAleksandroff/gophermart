package main

import (
	"context"
	"log"
	"time"

	"github.com/IgorAleksandroff/gophermart/internal/app"
	"github.com/IgorAleksandroff/gophermart/internal/config"
	"github.com/IgorAleksandroff/gophermart/internal/worker"
)

func main() {
	ctx, closeCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer closeCtx()

	cfg := config.GetConfig()

	app, err := app.NewApp(ctx, cfg)
	if err != nil {
		log.Fatalf("Create app error: %s", err)
	}
	defer app.Cancel()

	w := worker.NewUpdater(context.Background(), cfg)

	// start worker for update statuses of orders
	go w.Run()

	app.Run()
}
