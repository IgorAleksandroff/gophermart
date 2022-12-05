package worker

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/IgorAleksandroff/gophermart/internal/usecase"
	"github.com/IgorAleksandroff/gophermart/pkg/logger"
)

const updatePeriod = 100 * time.Millisecond

type Updater struct {
	period   time.Duration
	statuses usecase.UpdaterStatuses
	ctx      context.Context
	l        *logger.Logger
}

func NewUpdater(ctx context.Context, statusesUsecase usecase.UpdaterStatuses, l *logger.Logger) *Updater {
	return &Updater{
		period:   updatePeriod,
		statuses: statusesUsecase,
		ctx:      ctx,
		l:        l,
	}
}

func (u *Updater) Run() {
	ticker := time.NewTicker(u.period)
	defer ticker.Stop()

	for {
		<-ticker.C

		log.Println("worker: start")
		err := u.statuses.UpdateStatus(u.ctx)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			u.l.Warn("can't update metrics, %s", err.Error())
		}
	}
}
