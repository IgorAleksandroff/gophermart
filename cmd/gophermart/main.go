package main

import (
	"context"
	"log"
	"time"

	"github.com/IgorAleksandroff/gophermart/internal/app"
	"github.com/IgorAleksandroff/gophermart/internal/config"
)

func main() {
	ctx, closeCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer closeCtx()

	app, err := app.NewApp(ctx, config.GetConfig())
	defer app.Cancel()

	if err != nil {
		log.Fatalf("Create app error: %s", err)
	}

	app.Run()
}
