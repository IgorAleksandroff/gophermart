package repository

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/IgorAleksandroff/gophermart/internal/entity"
	"github.com/IgorAleksandroff/gophermart/pkg/logger"
)

type memoRep struct {
	orders   map[string]entity.Order
	users    map[string]entity.User
	withdraw map[string]entity.OrderWithdraw
	mu       *sync.Mutex
	l        *logger.Logger
}

func NewMemoRepository(ctx context.Context, log *logger.Logger) *memoRep {
	o := make(map[string]entity.Order)
	u := make(map[string]entity.User)
	w := make(map[string]entity.OrderWithdraw)

	return &memoRep{orders: o, users: u, withdraw: w, mu: &sync.Mutex{}, l: log}
}

func (m *memoRep) SaveUser(ctx context.Context, user entity.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.users[user.Login]
	if ok {
		return ErrUserRegister
	}

	m.users[user.Login] = user

	return nil
}

func (m *memoRep) GetUser(ctx context.Context, login string) (entity.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	userSaved, ok := m.users[login]
	if !ok {
		return entity.User{}, ErrUserLogin
	}

	return userSaved, nil
}

func (m *memoRep) GetOrder(ctx context.Context, orderID string) (*entity.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	existedOrder, ok := m.orders[orderID]
	if !ok {
		return nil, errors.New("unknown order")
	}

	return &existedOrder, nil
}

func (m *memoRep) SaveOrder(ctx context.Context, order entity.Order) error {
	order.UploadedAt = time.Now().Format(time.RFC3339)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.orders[order.OrderID] = order

	return nil
}

func (m *memoRep) GetOrders(ctx context.Context, login string) ([]entity.Orders, error) {
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

func (m *memoRep) UpdateUser(ctx context.Context, user entity.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.users[user.Login]
	if !ok {
		return errors.New("unknown user")
	}

	m.users[user.Login] = user

	return nil
}

func (m *memoRep) SupplementBalance(ctx context.Context, order entity.Order) error {
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

func (m *memoRep) SaveWithdrawn(ctx context.Context, withdrawn entity.OrderWithdraw) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.withdraw[withdrawn.OrderID]
	if ok {
		return errors.New("withdrawn already exist")
	}

	m.withdraw[withdrawn.OrderID] = withdrawn

	return nil
}

func (m *memoRep) GetWithdrawals(ctx context.Context, login string) ([]entity.OrderWithdraw, error) {
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

func (m *memoRep) GetOrderForUpdate(ctx context.Context) (*entity.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var oldestOrder *entity.Order
	for _, o := range m.orders {
		if o.Status != completedStatus {
			continue
		}

		if oldestOrder == nil {
			oldestOrder = &o
			continue
		}

		if oldestOrder.UploadedAt < o.UploadedAt {
			oldestOrder = &o
		}
	}

	return oldestOrder, nil
}

func (m *memoRep) Close() {}
