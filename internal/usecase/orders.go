package usecase

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/IgorAleksandroff/gophermart.git/internal/entity"
)

//go:generate mockery --name Orders
//go:generate mockery --name Repository

const accrualEndpoint = "/api/orders/"

var ErrExistOrderByThisUser = errors.New("order number already uploaded by this user")
var ErrExistOrderByAnotherUser = errors.New("order number already uploaded by another user")
var ErrLowBalance = errors.New("low balance of current user")

type ordersUsecase struct {
	rep           repository
	accrualClient apiClient
}

type Orders interface {
	GetUser(login string) (entity.User, error)
	SaveOrder(order entity.Order) error
	GetOrders() ([]entity.Orders, error)
	SaveWithdrawn(order entity.OrderWithdraw) error
	GetWithdrawals() ([]entity.OrderWithdraw, error)
}

type repository interface {
	GetUser(login string) (entity.User, error)
	SaveOrder(order entity.Order) (*string, error)
	GetOrders() ([]entity.Orders, error)
	UpdateUser(user entity.User) error
	SupplementBalance(order entity.Order) error
	SaveWithdrawn(order entity.OrderWithdraw) error
	GetWithdrawals() ([]entity.OrderWithdraw, error)
}

type apiClient interface {
	DoGet(url string) ([]byte, error)
}

func NewOrders(r repository, c apiClient) ordersUsecase {
	return ordersUsecase{rep: r, accrualClient: c}
}

func (o *ordersUsecase) GetUser(login string) (entity.User, error) {
	return o.rep.GetUser(login)
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
	userLogin, err := o.rep.SaveOrder(order)
	if err != nil {
		return err
	}
	if userLogin != nil {
		if *userLogin == order.UserLogin {
			return ErrExistOrderByThisUser
		}

		return ErrExistOrderByAnotherUser
	}

	err = o.rep.SupplementBalance(order)
	if err != nil {
		return err
	}

	return nil
}

func (o *ordersUsecase) GetOrders() ([]entity.Orders, error) {
	return o.rep.GetOrders()
}

func (o *ordersUsecase) SaveWithdrawn(withdrawn entity.OrderWithdraw) error {
	user, err := o.rep.GetUser(withdrawn.UserLogin)
	if err != nil {
		return err
	}

	if withdrawn.Value > user.Current {
		return ErrLowBalance
	}
	user.Current = -withdrawn.Value
	user.Withdrawn = +withdrawn.Value

	// todo сохранить баланс в пользователя через транзакцию
	err = o.rep.UpdateUser(user)
	if err != nil {
		return err
	}

	if err = o.rep.SaveWithdrawn(withdrawn); err != nil {
		return err
	}

	return nil
}

func (o *ordersUsecase) GetWithdrawals() ([]entity.OrderWithdraw, error) {
	return o.rep.GetWithdrawals()
}
