package repository

import (
	"github.com/IgorAleksandroff/gophermart.git/internal/entity"
)

// balance    DECIMAL(16, 4) NOT NULL DEFAULT 0
// На практике - флоат нужен только для баланса
type memoRep struct {
	orders   map[string]entity.Order
	users    map[string]entity.User
	withdraw map[string]entity.OrderWithdraw
}

func NewMemoRepository() *memoRep {
	o := make(map[string]entity.Order)
	u := make(map[string]entity.User)
	w := make(map[string]entity.OrderWithdraw)

	return &memoRep{orders: o, users: u, withdraw: w}
}

func (m *memoRep) SaveOrder(order entity.Order) (*int64, error) {
	orderSaved, ok := m.orders[order.OrderID]
	if ok {
		return &orderSaved.UserID, nil
	}

	m.orders[order.OrderID] = order

	return nil, nil
}

func (m *memoRep) GetOrders() ([]entity.Orders, error) {
	result := make([]entity.Orders, 0, len(m.orders))
	for _, order := range m.orders {
		result = append(result, entity.Orders{
			OrderID:    order.OrderID,
			Status:     order.Status,
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt,
		})
	}

	return result, nil
}
