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

type ordersUsecase struct {
	rep           repository
	accrualClient apiClient
}

type Orders interface {
	SaveOrder(order entity.Order) error
	GetOrders() ([]entity.Orders, error)
}

type repository interface {
	SaveOrder(order entity.Order) (*int64, error)
	GetOrders() ([]entity.Orders, error)
}

type apiClient interface {
	DoGet(url string) ([]byte, error)
}

func NewOrders(r repository, c apiClient) ordersUsecase {
	return ordersUsecase{rep: r, accrualClient: c}
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

	// todo сохранить баланс в пользователя
	userID, err := o.rep.SaveOrder(order)
	if err != nil {
		return err
	}
	if userID != nil {
		if *userID == order.UserID {
			return ErrExistOrderByThisUser
		}

		return ErrExistOrderByAnotherUser
	}

	return nil
}

func (o *ordersUsecase) GetOrders() ([]entity.Orders, error) {
	return o.rep.GetOrders()
}
