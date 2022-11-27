package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/IgorAleksandroff/gophermart/internal/entity"
	"github.com/IgorAleksandroff/gophermart/pkg/logger"
)

const completedStatus = "PROCESSED"

const (
	queryCreateTables = `CREATE TABLE IF NOT EXISTS users (
			id serial,
			login VARCHAR(64) PRIMARY KEY,
			password VARCHAR(128),
			current DECIMAL(16, 4) NOT NULL DEFAULT 0,
			withdrawn DECIMAL(16, 4) NOT NULL DEFAULT 0
		);
		CREATE TABLE IF NOT EXISTS orders (
			order_id VARCHAR(64) PRIMARY KEY,
			login VARCHAR(64) REFERENCES users(login),
			status VARCHAR(32) NOT NULL,
			accrual DECIMAL(16, 4) NOT NULL DEFAULT 0,
			uploaded_at VARCHAR(32) NOT NULL
		);
		CREATE TABLE IF NOT EXISTS orders_withdraws (
			order_id VARCHAR(64) PRIMARY KEY,
			login VARCHAR(64) REFERENCES users(login),
			value DECIMAL(16, 4) NOT NULL DEFAULT 0,
			processed_at VARCHAR(32) NOT NULL
		);
	`
	querySaveUser = `INSERT INTO users (login, password) VALUES ($1, $2)
		ON CONFLICT (login) DO NOTHING`
	queryGetUser    = `SELECT login, password, current, withdrawn FROM users WHERE login = $1`
	queryUpdateUser = `UPDATE users 
		SET current = $2,
				withdrawn = $3
		WHERE login = $1`
	querySupplementUser = `UPDATE users 
		SET current = current + $2
		WHERE login = $1`

	querySaveOrder = `INSERT INTO orders (order_id, login, status, accrual, uploaded_at) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (order_id) DO UPDATE
		    SET (status, accrual, uploaded_at) = (EXCLUDED.status, EXCLUDED.accrual, EXCLUDED.uploaded_at)`
	queryGetOrder          = `SELECT order_id, login, status, accrual, uploaded_at FROM orders WHERE order_id = $1`
	queryGetOrders         = `SELECT order_id, status, accrual, uploaded_at FROM orders WHERE login = $1`
	queryGetOrderForUpdate = `SELECT order_id, login, status, accrual, uploaded_at FROM orders WHERE status <> $1 ORDER BY uploaded_at`

	querySaveWithdrawn = `INSERT INTO orders_withdraws (order_id, login, value, processed_at) VALUES ($1, $2, $3, $4)
		ON CONFLICT (order_id) DO NOTHING`
	queryGetWithdrawals = `SELECT order_id, login, value, processed_at FROM orders_withdraws WHERE login = $1`
)

type pgRep struct {
	db *sqlx.DB
	l  *logger.Logger
}

var instance pgRep
var once sync.Once

func NewPGRepository(ctx context.Context, log *logger.Logger, addressDB string) *pgRep {
	once.Do(func() {
		db, err := sqlx.Connect("postgres", addressDB)
		if err != nil {
			log.Fatal(fmt.Errorf("app - New - postgres.New: %w", err))
		}

		instance = pgRep{db: db, l: log}
		if err = instance.init(ctx); err != nil {
			log.Fatal(fmt.Errorf("app - New - postgres.`Init`: %w", err))
		}
	})

	return &instance
}

func (p *pgRep) init(ctx context.Context) error {
	_, err := p.db.ExecContext(ctx, queryCreateTables)
	if err != nil {
		return err
	}
	return nil
}

func (p *pgRep) SaveUser(ctx context.Context, user entity.User) error {
	res, err := p.db.ExecContext(ctx, querySaveUser,
		user.Login,
		user.Password,
	)
	if err != nil {
		return fmt.Errorf("error to save user: %w, %+v", err, user)

	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error get rows affected: %w", err)
	}
	if affected == 0 {
		return ErrUserRegister
	}

	return nil
}

func (p *pgRep) GetUser(ctx context.Context, login string) (entity.User, error) {
	var user entity.User

	err := p.db.QueryRowContext(
		ctx,
		queryGetUser,
		login,
	).Scan(&user.Login, &user.Password, &user.Current, &user.Withdrawn)
	if err != nil {
		return entity.User{}, fmt.Errorf("error to get user: %w, %s", err, login)
	}

	return user, nil
}

func (p *pgRep) SaveOrder(ctx context.Context, order entity.Order) error {
	res, err := p.db.ExecContext(ctx, querySaveOrder,
		order.OrderID,
		order.UserLogin,
		order.Status,
		order.Accrual,
		order.UploadedAt,
	)
	if err != nil {
		return fmt.Errorf("error to save order: %w, %+v", err, order)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error to get rows after save order: %w, %+v", err, order)
	}
	if rows <= 0 {
		return fmt.Errorf("rows affected %v <= 0, after save order: %+v", rows, order)
	}

	return nil
}

func (p *pgRep) GetOrder(ctx context.Context, orderID string) (*entity.Order, error) {
	var order entity.Order

	err := p.db.QueryRowContext(
		ctx,
		queryGetOrder,
		orderID,
	).Scan(&order.OrderID, &order.UserLogin, &order.Status, &order.Accrual, &order.UploadedAt)
	if err != nil {
		return &entity.Order{}, fmt.Errorf("error to get order: %w, %s", err, orderID)
	}

	p.l.Info("debug, order for save = %+v", order)

	return &order, nil
}

func (p *pgRep) GetOrders(ctx context.Context, login string) ([]entity.Orders, error) {
	var result []entity.Orders

	err := p.db.SelectContext(
		ctx,
		&result,
		queryGetOrders,
		login,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (p *pgRep) UpdateUser(ctx context.Context, user entity.User) error {
	res, err := p.db.ExecContext(ctx, queryUpdateUser,
		user.Login,
		user.Current,
		user.Withdrawn,
	)
	if err != nil {
		return fmt.Errorf("error to update user: %w, %+v", err, user)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error to get rows after update user: %w, %+v", err, user)
	}
	if rows <= 0 {
		return fmt.Errorf("rows affected %v <= 0, after update user: %+v", rows, user)
	}

	return nil
}

func (p *pgRep) SupplementBalance(ctx context.Context, order entity.Order) error {
	if order.Accrual == 0 {
		return nil
	}

	res, err := p.db.ExecContext(ctx, querySupplementUser,
		order.UserLogin,
		order.Accrual,
	)
	if err != nil {
		return fmt.Errorf("error to supplement balance: %w, %+v", err, order)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error to get rows after supplement balance: %w, %+v", err, order)
	}
	if rows <= 0 {
		return fmt.Errorf("rows affected %v <= 0, after supplement balance: %+v", rows, order)
	}

	return nil
}

func (p *pgRep) SaveWithdrawn(ctx context.Context, withdrawn entity.OrderWithdraw) error {
	res, err := p.db.ExecContext(ctx, querySaveWithdrawn,
		withdrawn.OrderID,
		withdrawn.UserLogin,
		withdrawn.Value,
		withdrawn.ProcessedAt,
	)
	if err != nil {
		return fmt.Errorf("error to save withdrawn: %w, %+v", err, withdrawn)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error to get rows after save withdrawn: %w, %+v", err, withdrawn)
	}
	if rows <= 0 {
		return fmt.Errorf("rows affected %v <= 0, after save withdrawn: %+v", rows, withdrawn)
	}

	return nil
}

func (p *pgRep) GetWithdrawals(ctx context.Context, login string) ([]entity.OrderWithdraw, error) {
	var result []entity.OrderWithdraw

	err := p.db.SelectContext(
		ctx,
		&result,
		queryGetWithdrawals,
		login,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (p *pgRep) GetOrderForUpdate(ctx context.Context) (*entity.Order, error) {
	var order entity.Order

	err := p.db.QueryRowContext(
		ctx,
		queryGetOrderForUpdate,
		completedStatus,
	).Scan(&order.OrderID, &order.UserLogin, &order.Status, &order.Accrual, &order.UploadedAt)
	if err != nil {
		return &entity.Order{}, fmt.Errorf("error to get order for update: %w", err)
	}
	p.l.Info("заказ для обновления: %+v", order)

	return &order, nil
}

func (p *pgRep) Close() {
	p.db.Close()
}
