package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/IgorAleksandroff/gophermart/internal/entity"
)

//go:generate mockery --name Orders
//go:generate mockery --name OrdersRepository

const accrualEndpoint = "/api/orders/"
const accrualStatusNoContent = "PROCESSING"

var ErrExistOrderByThisUser = errors.New("order number already uploaded by this user")
var ErrExistOrderByAnotherUser = errors.New("order number already uploaded by another user")
var ErrLowBalance = errors.New("low balance of current user")

type ordersUsecase struct {
	repo          OrdersRepository
	accrualClient apiClient
}

type Orders interface {
	GetUser(ctx context.Context, login string) (entity.User, error)
	SaveOrder(ctx context.Context, order entity.Order) error
	GetOrders(ctx context.Context, login string) ([]entity.Orders, error)
	SaveWithdrawn(ctx context.Context, order entity.OrderWithdraw) error
	GetWithdrawals(ctx context.Context, login string) ([]entity.OrderWithdraw, error)
}

type OrdersRepository interface {
	GetUser(ctx context.Context, login string) (entity.User, error)
	SaveOrder(ctx context.Context, order entity.Order) error
	GetOrder(ctx context.Context, orderID string) (*entity.Order, error)
	GetOrders(ctx context.Context, login string) ([]entity.Orders, error)
	UpdateUser(ctx context.Context, user entity.User) error
	SupplementBalance(ctx context.Context, order entity.Order) error
	SaveWithdrawn(ctx context.Context, order entity.OrderWithdraw) error
	GetWithdrawals(ctx context.Context, login string) ([]entity.OrderWithdraw, error)
	Close()
}

type apiClient interface {
	DoGet(url string) ([]byte, error)
}

func NewOrders(r OrdersRepository, c apiClient) *ordersUsecase {
	return &ordersUsecase{repo: r, accrualClient: c}
}

func (o *ordersUsecase) GetUser(ctx context.Context, login string) (entity.User, error) {
	return o.repo.GetUser(ctx, login)
}

func (o *ordersUsecase) SaveOrder(ctx context.Context, order entity.Order) error {
	var accrual entity.Accrual
	out, err := o.accrualClient.DoGet(accrualEndpoint + order.OrderID)
	if err != nil {
		return fmt.Errorf("error with order %v from service accurual: %w", order.OrderID, err)
	}

	err = json.Unmarshal(out, &accrual)
	if err != nil && len(out) != 0 {
		return fmt.Errorf("error with order %v parse answer from service accurual: %w", order.OrderID, err)
	}

	if len(out) == 0 {
		accrual.Status = accrualStatusNoContent
	}

	order.Status = accrual.Status
	if accrual.Accrual != nil {
		order.Accrual = *accrual.Accrual
	}

	var existError error
	existedOrder, _ := o.repo.GetOrder(ctx, order.OrderID)
	if existedOrder != nil {
		if existedOrder.UserLogin != order.UserLogin {
			log.Println(fmt.Printf("debug, existedOrder.UserLogin = %v, order.UserLogin = %v", existedOrder.UserLogin, order.UserLogin))
			return ErrExistOrderByAnotherUser
		}

		existError = ErrExistOrderByThisUser
	}

	err = o.repo.SaveOrder(ctx, order)
	if err != nil {
		return err
	}

	err = o.repo.SupplementBalance(ctx, order)
	if err != nil {
		return err
	}

	return existError
}

func (o *ordersUsecase) GetOrders(ctx context.Context, login string) ([]entity.Orders, error) {
	return o.repo.GetOrders(ctx, login)
}

func (o *ordersUsecase) SaveWithdrawn(ctx context.Context, withdrawn entity.OrderWithdraw) error {
	user, err := o.repo.GetUser(ctx, withdrawn.UserLogin)
	if err != nil {
		return err
	}

	if withdrawn.Value > user.Current {
		return ErrLowBalance
	}
	user.Current = user.Current - withdrawn.Value
	user.Withdrawn = user.Withdrawn + withdrawn.Value

	err = o.repo.UpdateUser(ctx, user)
	if err != nil {
		return err
	}

	if err = o.repo.SaveWithdrawn(ctx, withdrawn); err != nil {
		return err
	}

	return nil
}

func (o *ordersUsecase) GetWithdrawals(ctx context.Context, login string) ([]entity.OrderWithdraw, error) {
	return o.repo.GetWithdrawals(ctx, login)
}
