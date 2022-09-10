package usecase

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/IgorAleksandroff/gophermart/internal/entity"
)

//go:generate mockery --name Orders
//go:generate mockery --name OrdersRepository

const accrualEndpoint = "/api/orders/"

var ErrExistOrderByThisUser = errors.New("order number already uploaded by this user")
var ErrExistOrderByAnotherUser = errors.New("order number already uploaded by another user")
var ErrLowBalance = errors.New("low balance of current user")

type ordersUsecase struct {
	repo          ordersRepository
	accrualClient apiClient
}

type Orders interface {
	GetUser(login string) (entity.User, error)
	SaveOrder(order entity.Order) error
	GetOrders(login string) ([]entity.Orders, error)
	SaveWithdrawn(order entity.OrderWithdraw) error
	GetWithdrawals(login string) ([]entity.OrderWithdraw, error)
}

type ordersRepository interface {
	GetUser(login string) (entity.User, error)
	SaveOrder(order entity.Order) (*string, error)
	GetOrders(login string) ([]entity.Orders, error)
	UpdateUser(user entity.User) error
	SupplementBalance(order entity.Order) error
	SaveWithdrawn(order entity.OrderWithdraw) error
	GetWithdrawals(login string) ([]entity.OrderWithdraw, error)
}

type apiClient interface {
	DoGet(url string) ([]byte, error)
}

func NewOrders(r ordersRepository, c apiClient) *ordersUsecase {
	return &ordersUsecase{repo: r, accrualClient: c}
}

func (o *ordersUsecase) GetUser(login string) (entity.User, error) {
	return o.repo.GetUser(login)
}

func (o *ordersUsecase) SaveOrder(order entity.Order) error {
	var accrual entity.Accrual
	out, err := o.accrualClient.DoGet(accrualEndpoint + order.OrderID)
	if err != nil {
		return fmt.Errorf("error from service accurual: %w", err)
	}

	err = json.Unmarshal(out, &accrual)
	if err != nil {
		return fmt.Errorf("error parse answer from service accurual: %w", err)
	}
	order.Status = accrual.Status
	order.Accrual = accrual.Accrual

	// todo сохранить баланс в пользователя через транзакцию
	userLogin, err := o.repo.SaveOrder(order)
	if err != nil {
		return err
	}
	if userLogin != nil {
		if *userLogin == order.UserLogin {
			return ErrExistOrderByThisUser
		}

		return ErrExistOrderByAnotherUser
	}

	err = o.repo.SupplementBalance(order)
	if err != nil {
		return err
	}

	return nil
}

func (o *ordersUsecase) GetOrders(login string) ([]entity.Orders, error) {
	return o.repo.GetOrders(login)
}

func (o *ordersUsecase) SaveWithdrawn(withdrawn entity.OrderWithdraw) error {
	user, err := o.repo.GetUser(withdrawn.UserLogin)
	if err != nil {
		return err
	}

	if withdrawn.Value > user.Current {
		return ErrLowBalance
	}
	user.Current = -withdrawn.Value
	user.Withdrawn = +withdrawn.Value

	// todo сохранить баланс в пользователя через транзакцию
	err = o.repo.UpdateUser(user)
	if err != nil {
		return err
	}

	if err = o.repo.SaveWithdrawn(withdrawn); err != nil {
		return err
	}

	return nil
}

func (o *ordersUsecase) GetWithdrawals(login string) ([]entity.OrderWithdraw, error) {
	return o.repo.GetWithdrawals(login)
}
