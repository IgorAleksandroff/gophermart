package usecase

import "github.com/IgorAleksandroff/gophermart.git/internal/entity"

//go:generate mockery --name Orders
//go:generate mockery --name Repository

type ordersUsecase struct {
	rep repository
}

type Orders interface {
	GetOrders() map[string]entity.Order
}

type repository interface {
	GetOrders() map[string]entity.Order
}

func NewOrders(r repository) ordersUsecase {
	return ordersUsecase{rep: r}
}

func (o *ordersUsecase) GetOrders() map[string]entity.Order {
	return o.rep.GetOrders()
}
