package usecase

import (
	"context"
	"fmt"

	"github.com/IgorAleksandroff/gophermart/internal/entity"
)

//go:generate mockery --name UpdaterStatuses

type statusesUsecase struct {
	orders orders
	repo   StatusesRepository
}

type UpdaterStatuses interface {
	UpdateStatus(ctx context.Context) error
}

type orders interface {
	SaveOrder(ctx context.Context, order entity.Order) error
}

type StatusesRepository interface {
	GetOrderForUpdate(ctx context.Context) (*entity.Order, error)
}

func NewStatuses(o orders, r StatusesRepository) *statusesUsecase {
	return &statusesUsecase{orders: o, repo: r}
}

func (s *statusesUsecase) UpdateStatus(ctx context.Context) error {
	order, err := s.repo.GetOrderForUpdate(ctx)
	if err != nil {
		return fmt.Errorf("error to get order for update: %w", err)
	}

	return s.orders.SaveOrder(ctx, *order)
}
