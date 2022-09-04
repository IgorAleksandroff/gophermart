package app

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/IgorAleksandroff/gophermart.git/pkg/logger"
	"github.com/go-chi/chi"

	"github.com/IgorAleksandroff/gophermart.git/internal/config"
	"github.com/IgorAleksandroff/gophermart.git/pkg/httpserver"
)

type app struct {
	cfg     *config.Config
	handler http.Handler
	l       *logger.Logger
}

func NewApp(cfg *config.Config) (*app, error) {
	l := logger.New(cfg.App.LogLevel)
	r := chi.NewRouter()

	return &app{
		cfg:     cfg,
		handler: r,
		l:       l,
	}, nil
}

func (a *app) Run() {
	// start http server
	httpServer := httpserver.New(a.handler, httpserver.Addr(a.cfg.HTTPServer.ServerAddress))

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case s := <-interrupt:
		a.l.Info("app - Run - signal: " + s.String())
	case err := <-httpServer.Notify():
		a.l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	err := httpServer.Shutdown()
	if err != nil {
		a.l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}
