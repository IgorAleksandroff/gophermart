package repository

import (
	"github.com/IgorAleksandroff/gophermart.git/internal/entity"
)

type memoRep struct {
	orders  map[int64]entity.Order
	users   map[string]entity.User
	balance map[int64]entity.Balance
}

func NewMemoRepository() *memoRep {
	o := make(map[int64]entity.Order)
	u := make(map[string]entity.User)
	b := make(map[int64]entity.Balance)

	return &memoRep{orders: o, users: u, balance: b}
}

func (m *memoRep) SaveOrder(order entity.Order) (*int64, error) {
	orderSaved, ok := m.orders[order.ID]
	if ok {
		return &orderSaved.UserID, nil
	}

	m.orders[order.ID] = order

	return nil, nil
}

func (m *memoRep) GetOrders() map[int64]entity.Order {
	return m.orders
}
