package worker

import (
	"context"
	"time"

	"github.com/IgorAleksandroff/gophermart/internal/config"
	"github.com/IgorAleksandroff/gophermart/internal/repository"
	"github.com/IgorAleksandroff/gophermart/internal/usecase"
	"github.com/IgorAleksandroff/gophermart/internal/webapi"
	"github.com/IgorAleksandroff/gophermart/pkg/logger"
)

const updatePeriodSeconds = 1

type updater struct {
	period   int64
	statuses usecase.UpdaterStatuses
	ctx      context.Context
	l        *logger.Logger
}

func NewUpdater(ctx context.Context, cfg *config.Config) *updater {
	l := logger.New(cfg.App.LogLevel)

	var ordersRepo usecase.OrdersRepository
	var statusesRepo usecase.StatusesRepository
	if cfg.App.DataBaseURI != "" {
		pgRepo := repository.NewPGRepository(ctx, l, cfg.App.DataBaseURI)
		ordersRepo, statusesRepo = pgRepo, pgRepo
	} else {
		inMemoRepo := repository.NewMemoRepository(ctx, l)
		ordersRepo, statusesRepo = inMemoRepo, inMemoRepo
	}

	apiClient := webapi.NewClient(cfg.App.AccrualSystemAddress)
	ordersUsecase := usecase.NewOrders(ordersRepo, apiClient)
	statusesUsecase := usecase.NewStatuses(ordersUsecase, statusesRepo)

	return &updater{
		period:   updatePeriodSeconds,
		statuses: statusesUsecase,
		ctx:      ctx,
		l:        l,
	}
}

func (u *updater) Run() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C

		err := u.statuses.UpdateStatus(u.ctx)
		if err != nil {
			u.l.Warn("can't save metrics, %s", err.Error())
		}
	}
}
