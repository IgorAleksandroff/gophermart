package app

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/IgorAleksandroff/gophermart/internal/webapi"
	"github.com/go-chi/chi"

	"github.com/IgorAleksandroff/gophermart/internal/config"
	"github.com/IgorAleksandroff/gophermart/internal/hendler"
	"github.com/IgorAleksandroff/gophermart/internal/repository"
	"github.com/IgorAleksandroff/gophermart/internal/usecase"
	"github.com/IgorAleksandroff/gophermart/pkg/httpserver"
	"github.com/IgorAleksandroff/gophermart/pkg/logger"
)

type app struct {
	cfg    *config.Config
	router http.Handler
	l      *logger.Logger
}

func NewApp(cfg *config.Config) (*app, error) {
	l := logger.New(cfg.App.LogLevel)
	r := chi.NewRouter()

	repo := repository.NewMemoRepository()
	apiClient := webapi.NewClient(cfg.App.AccrualSystemAddress)
	ordersUsecase := usecase.NewOrders(repo, apiClient)
	auth := usecase.NewAuthorization(repo)

	h := hendler.New(ordersUsecase, auth)

	h.Register(r, http.MethodPost, "/api/user/register", h.HandleUserRegister)
	h.Register(r, http.MethodPost, "/api/user/login", h.HandleUserLogin)

	r.Group(func(r chi.Router) {
		r.Use(h.UserIdentity)

		h.Register(r, http.MethodPost, "/api/user/orders", h.HandlePostOrders)
		h.Register(r, http.MethodGet, "/api/user/orders", h.HandleGetOrders)

		h.Register(r, http.MethodGet, "/api/user/balance", h.HandleGetBalance)
		h.Register(r, http.MethodPost, "/api/user/balance/withdraw", h.HandlePostBalanceWithdraw)
		h.Register(r, http.MethodGet, "/api/user/withdrawals", h.HandleGetWithdrawals)
	})

	return &app{
		cfg:    cfg,
		router: r,
		l:      l,
	}, nil
}

func (a *app) Run() {
	// start http server
	httpServer := httpserver.New(a.router, httpserver.Addr(a.cfg.HTTPServer.ServerAddress))

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
