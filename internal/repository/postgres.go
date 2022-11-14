package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/IgorAleksandroff/gophermart/internal/entity"
	"github.com/IgorAleksandroff/gophermart/pkg/logger"
)

const (
	queryCreateTables = `CREATE TABLE IF NOT EXISTS users (
			id serial,
			login VARCHAR(64) primary key,
			password VARCHAR(128) DEFAULT NULL,
			current DECIMAL(16, 4) NOT NULL DEFAULT 0,
			withdrawn DECIMAL(16, 4) NOT NULL DEFAULT 0
		);
		CREATE TABLE IF NOT EXISTS orders (
			id serial,
			order_id VARCHAR(64) primary key,
			user_login VARCHAR(32) NOT NULL,
			status VARCHAR(32) NOT NULL,
			accrual DECIMAL(16, 4) NOT NULL DEFAULT 0,
			uploaded_at VARCHAR(32) NOT NULL
		);
		CREATE TABLE IF NOT EXISTS orders_withdraws (
			id serial,
			order_id VARCHAR(64) primary key,
			user_login VARCHAR(32) NOT NULL,
			value DECIMAL(16, 4) NOT NULL DEFAULT 0,
			processed_at VARCHAR(32) NOT NULL
		);
	`
	querySaveUser = `INSERT INTO users (login, password) VALUES ($1, $2)
		ON CONFLICT (login) DO NOTHING`
	queryGetUser    = `SELECT login, password, current, withdrawn FROM users WHERE login = $1`
	queryUpdateUser = `UPDATE users 
		SET password = $2,
				current = $3,
				withdrawn = $4
		WHERE login = $1`
	querySupplementUser = `UPDATE users 
		SET current = current + $2
		WHERE login = $1`

	querySaveOrder = `INSERT INTO orders (order_id, user_login, status, accrual, uploaded_at) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (order_id) DO UPDATE
		    SET (user_login, status, accrual, uploaded_at) = (EXCLUDED.user_login, EXCLUDED.status, EXCLUDED.accrual, EXCLUDED.uploaded_at)`
	queryGetOrder  = `SELECT order_id, user_login, status, accrual, uploaded_at FROM orders WHERE order_id = $1`
	queryGetOrders = `SELECT order_id, status, accrual, uploaded_at FROM orders WHERE user_login = $1`

	querySaveWithdrawn = `INSERT INTO orders_withdraws (order_id, user_login, value, processed_at) VALUES ($1, $2)
		ON CONFLICT (order_id) DO NOTHING`
	queryGetWithdrawals = `SELECT order_id, user_login, value, processed_at FROM orders_withdraws WHERE user_login = $1`
)

type pgRep struct {
	ctx context.Context
	db  *sqlx.DB
	l   *logger.Logger
}

func NewPGRepository(log *logger.Logger, addressDB string) *pgRep {
	db, err := sqlx.Connect("postgres", addressDB)
	if err != nil {
		log.Fatal(fmt.Errorf("app - New - postgres.New: %w", err))
	}

	repositoryPG := pgRep{ctx: context.Background(), db: db, l: log}
	if err = repositoryPG.init(); err != nil {
		log.Fatal(fmt.Errorf("app - New - postgres.`Init`: %w", err))
	}

	return &repositoryPG
}

func (p *pgRep) init() error {
	_, err := p.db.ExecContext(p.ctx, queryCreateTables)
	if err != nil {
		return err
	}
	return nil
}

func (p *pgRep) SaveUser(user entity.User) error {
	res, err := p.db.ExecContext(p.ctx, querySaveUser,
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

func (p *pgRep) GetUser(login string) (entity.User, error) {
	var users []entity.User

	err := p.db.SelectContext(
		p.ctx,
		&users,
		queryGetUser,
		login,
	)
	if err != nil {
		return entity.User{}, fmt.Errorf("error to get users: %w, %s", err, login)
	}

	if len(users) == 0 {
		return entity.User{}, fmt.Errorf("unknown user: %s", login)
	}

	return users[0], nil
}

func (p *pgRep) SaveOrder(order entity.Order) error {
	_, err := p.db.ExecContext(p.ctx, querySaveOrder,
		order.OrderID,
		order.UserLogin,
		order.Status,
		order.Accrual,
		order.UploadedAt,
	)

	return fmt.Errorf("error to save order: %w, %+v", err, order)
}

func (p *pgRep) GetOrder(orderID string) (*entity.Order, error) {
	var order entity.Order

	err := p.db.SelectContext(
		p.ctx,
		&order,
		queryGetOrder,
		orderID,
	)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (p *pgRep) GetOrders(login string) ([]entity.Orders, error) {
	var result []entity.Orders

	err := p.db.SelectContext(
		p.ctx,
		&result,
		queryGetOrders,
		login,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (p *pgRep) UpdateUser(user entity.User) error {
	_, err := p.db.ExecContext(p.ctx, queryUpdateUser,
		user.Login,
		user.Password,
		user.Current,
		user.Withdrawn,
	)

	return fmt.Errorf("error to update user: %w, %+v", err, user)
}

func (p *pgRep) SupplementBalance(order entity.Order) error {
	if order.Accrual == 0 {
		return nil
	}

	_, err := p.db.ExecContext(p.ctx, querySupplementUser,
		order.UserLogin,
		order.Accrual,
	)

	return fmt.Errorf("error to supplement balance: %w, %+v", err, order)
}

func (p *pgRep) SaveWithdrawn(withdrawn entity.OrderWithdraw) error {
	_, err := p.db.ExecContext(p.ctx, querySaveWithdrawn,
		withdrawn.OrderID,
		withdrawn.UserLogin,
		withdrawn.Value,
		withdrawn.ProcessedAt,
	)

	return fmt.Errorf("error to save withdrawn: %w, %+v", err, withdrawn)
}

func (p *pgRep) GetWithdrawals(login string) ([]entity.OrderWithdraw, error) {
	var result []entity.OrderWithdraw

	err := p.db.SelectContext(
		p.ctx,
		&result,
		queryGetWithdrawals,
		login,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (p *pgRep) Close() {
	p.db.Close()
}
