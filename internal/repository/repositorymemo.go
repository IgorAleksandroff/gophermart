package repository

import (
	"github.com/IgorAleksandroff/gophermart.git/internal/entity"
)

type memoRep struct {
	orders map[string]entity.Order
}

func NewMemoRepository() *memoRep {
	o := make(map[string]entity.Order)
	o["test"] = entity.Order{ID: 1}

	return &memoRep{orders: o}
}

func (m *memoRep) GetOrders() map[string]entity.Order {
	return m.orders
}
