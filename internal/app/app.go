package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/IgorAleksandroff/gophermart/internal/config"
	"github.com/IgorAleksandroff/gophermart/internal/hendler"
	"github.com/IgorAleksandroff/gophermart/internal/repository"
	"github.com/IgorAleksandroff/gophermart/internal/usecase"
	"github.com/IgorAleksandroff/gophermart/internal/webapi"
	"github.com/IgorAleksandroff/gophermart/internal/worker"
	"github.com/IgorAleksandroff/gophermart/pkg/httpserver"
	"github.com/IgorAleksandroff/gophermart/pkg/logger"
	"github.com/go-chi/chi"
)

type app struct {
	cfg    *config.Config
	router http.Handler
	worker *worker.Updater
	l      *logger.Logger
	Cancel cancelFunc
}

type cancelFunc func()

func NewApp(ctx context.Context, cfg *config.Config) (*app, error) {
	l := logger.New(cfg.App.LogLevel)
	r := chi.NewRouter()

	l.Debug("start NewApp")

	var repo usecase.OrdersRepository
	var authRepo usecase.UserRepository
	var statusesRepo usecase.StatusesRepository
	if cfg.App.DataBaseURI != "" {
		pgRepo := repository.NewPGRepository(ctx, l, cfg.App.DataBaseURI)
		repo, authRepo, statusesRepo = pgRepo, pgRepo, pgRepo
	} else {
		inMemoRepo := repository.NewMemoRepository(ctx, l)
		repo, authRepo, statusesRepo = inMemoRepo, inMemoRepo, inMemoRepo
	}

	apiClient := webapi.NewClient(cfg.App.AccrualSystemAddress)
	ordersUsecase := usecase.NewOrders(repo, apiClient)
	auth := usecase.NewAuthorization(authRepo)
	statusesUsecase := usecase.NewStatuses(ordersUsecase, statusesRepo)

	w := worker.NewUpdater(ctx, statusesUsecase, l)

	h := hendler.New(ordersUsecase, auth, l)

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
		worker: w,
		l:      l,
		Cancel: repo.Close,
	}, nil
}

func (a *app) Run() {
	// start http server
	httpServer := httpserver.New(a.router, httpserver.Addr(a.cfg.HTTPServer.ServerAddress))

	// start worker for update statuses of orders
	go a.worker.Run()

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
