package repository

import (
	"errors"
	"sync"

	"github.com/IgorAleksandroff/gophermart/internal/entity"
	"github.com/IgorAleksandroff/gophermart/pkg/logger"
)

var ErrUserRegister = errors.New("user already exist")
var ErrUserLogin = errors.New("unknown user")

// balance    DECIMAL(16, 4) NOT NULL DEFAULT 0
// На практике - флоат нужен только для баланса
type memoRep struct {
	orders   map[string]entity.Order
	users    map[string]entity.User
	withdraw map[string]entity.OrderWithdraw
	mu       *sync.Mutex
	l        *logger.Logger
}

func NewMemoRepository() *memoRep {
	o := make(map[string]entity.Order)
	u := make(map[string]entity.User)
	w := make(map[string]entity.OrderWithdraw)

	return &memoRep{orders: o, users: u, withdraw: w, mu: &sync.Mutex{}}
}

func (m *memoRep) SaveUser(user entity.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.users[user.Login]
	if ok {
		return ErrUserRegister
	}

	m.users[user.Login] = user

	return nil
}

func (m *memoRep) GetUser(login string) (entity.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	userSaved, ok := m.users[login]
	if !ok {
		return entity.User{}, ErrUserLogin
	}

	return userSaved, nil
}

func (m *memoRep) SaveOrder(order entity.Order) (*string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	orderSaved, ok := m.orders[order.OrderID]
	if ok {
		return &orderSaved.UserLogin, nil
	}

	m.orders[order.OrderID] = order

	return nil, nil
}

func (m *memoRep) GetOrders(login string) ([]entity.Orders, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]entity.Orders, 0, len(m.orders))
	for _, order := range m.orders {
		if order.UserLogin == login {
			result = append(result, entity.Orders{
				OrderID:    order.OrderID,
				Status:     order.Status,
				Accrual:    order.Accrual,
				UploadedAt: order.UploadedAt,
			})
		}
	}

	return result, nil
}

func (m *memoRep) UpdateUser(user entity.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.users[user.Login]
	if !ok {
		return errors.New("unknown user")
	}

	m.users[user.Login] = user

	return nil
}

func (m *memoRep) SupplementBalance(order entity.Order) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if order.Accrual == 0 {
		return nil
	}

	userSaved, ok := m.users[order.UserLogin]
	if !ok {
		return errors.New("unknown user")
	}

	userSaved.Current = userSaved.Current + order.Accrual
	m.users[order.UserLogin] = userSaved

	return nil
}

func (m *memoRep) SaveWithdrawn(withdrawn entity.OrderWithdraw) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.withdraw[withdrawn.OrderID]
	if ok {
		return errors.New("withdrawn already exist")
	}

	m.withdraw[withdrawn.OrderID] = withdrawn

	return nil
}

func (m *memoRep) GetWithdrawals(login string) ([]entity.OrderWithdraw, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]entity.OrderWithdraw, 0, len(m.orders))
	for _, order := range m.withdraw {
		if order.UserLogin == login {
			result = append(result, order)
		}
	}

	return result, nil
}
