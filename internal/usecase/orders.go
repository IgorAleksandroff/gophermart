package usecase

import (
	"errors"

	"github.com/IgorAleksandroff/gophermart.git/internal/entity"
)

//go:generate mockery --name Orders
//go:generate mockery --name Repository

var ErrExistOrderByThisUser = errors.New("order number already uploaded by this user")
var ErrExistOrderByAnotherUser = errors.New("order number already uploaded by another user")

type ordersUsecase struct {
	rep repository
}

type Orders interface {
	SaveOrder(order entity.Order) error
	GetOrders() map[int64]entity.Order
}

type repository interface {
	SaveOrder(order entity.Order) (*int64, error)
	GetOrders() map[int64]entity.Order
}

func NewOrders(r repository) ordersUsecase {
	return ordersUsecase{rep: r}
}

func (o *ordersUsecase) SaveOrder(order entity.Order) error {
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

func (o *ordersUsecase) GetOrders() map[int64]entity.Order {
	return o.rep.GetOrders()
}
