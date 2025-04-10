package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	add "github.com/Tanya1515/gophermarket/cmd/additional"
)

func (db *PostgreSQL) RegisterNewUser(ctx context.Context, user add.User) error {

	_, err := db.dbConn.ExecContext(ctx, "INSERT INTO users (login, password, sum, with_drawn) VALUES($1,crypt($2, gen_salt('xdes')),$3,$4)", user.Login, user.Password, 0, 0)

	if err != nil {
		return fmt.Errorf("error while inserting user with login %s: %w", user.Login, err)
	}
	return nil
}

func (db *PostgreSQL) AddNewOrder(ctx context.Context, orderNumber string) (err error) {

	var userID string
	ownerID := ctx.Value(add.LogginKey)

	rows, err := db.dbConn.QueryContext(ctx, "SELECT user_id FROM orders WHERE id=$1", orderNumber)
	if (err != nil) && !errors.Is(err, sql.ErrNoRows) {
		return
	}
	defer rows.Close()

	if !errors.Is(err, sql.ErrNoRows) {
		for rows.Next() {
			err = rows.Scan(&userID)
			if err != nil {
				return
			}

			if userID == ownerID {
				return fmt.Errorf("the order with number %s is in the system", orderNumber)
			} else {
				return fmt.Errorf("error: order with number %s already exists and belongs to another user", orderNumber)
			}
		}

		err = rows.Err()
		if err != nil {
			return
		}
	}

	_, err = db.dbConn.ExecContext(ctx, "INSERT INTO orders (id, status, accrual, UploadedAt, user_id) VALUES($1, $2, $3, $4, $5)", orderNumber, "NEW", 0, time.Now().Format(time.RFC3339), ctx.Value(add.LogginKey))

	return
}

func (db *PostgreSQL) ProcessPayPoints(ctx context.Context, order add.OrderSpend) (err error) {

	tx, err := db.dbConn.Begin()

	if err != nil {
		return fmt.Errorf("error while starting transaction: %w", err)
	}

	_, err = tx.ExecContext(ctx, "UPDATE users SET sum=sum-$1 WHERE id=$2;", order.Sum, ctx.Value(add.LogginKey))
	if err != nil {
		tx.Rollback()
		return
	}

	_, err = tx.ExecContext(ctx, "UPDATE users SET with_drawn=with_drawn+$1 WHERE id=$2;", order.Sum, ctx.Value(add.LogginKey))
	if err != nil {
		tx.Rollback()
		return
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO order_spend (id, ProcessedAt, sum, user_id) VALUES($1, $2, $3, $4);", order.Number, time.Now().Format(time.RFC3339), order.Sum, ctx.Value(add.LogginKey))
	if err != nil {
		tx.Rollback()
		return
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return
	}

	return
}

func (db *PostgreSQL) CheckUserLogin(ctx context.Context, login string) error {
	var value string

	row := db.dbConn.QueryRowContext(ctx, "SELECT login FROM users WHERE login = $1", login)

	err := row.Scan(&value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	return fmt.Errorf("user with login %s already exists", login)
}

func (db *PostgreSQL) CheckUserJWT(ctx context.Context, login string) (id string, err error) {

	row := db.dbConn.QueryRowContext(ctx, "SELECT id FROM users WHERE login = $1", login)

	err = row.Scan(&id)

	return
}

func (db *PostgreSQL) CheckUser(ctx context.Context, login, password string) (ok bool, err error) {
	ok = true
	row := db.dbConn.QueryRowContext(ctx, `SELECT (password = crypt($1, password)) 
								AS password_match
								FROM users
								WHERE login = $2;`, password, login)

	err = row.Scan(&ok)
	if errors.Is(err, sql.ErrNoRows) {
		ok = false
		return
	}
	return
}

func (db *PostgreSQL) GetUserBalance(ctx context.Context) (balance add.Balance, err error) {

	row := db.dbConn.QueryRowContext(ctx, `SELECT sum, with_drawn FROM Users WHERE id = $1`, ctx.Value(add.LogginKey))
	err = row.Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		return
	}

	return
}

func (db *PostgreSQL) GetAllOrders(ctx context.Context, orders *[]add.Order) (err error) {
	var order add.Order
	rows, err := db.dbConn.QueryContext(ctx, "SELECT id, status, UploadedAt, accrual FROM orders WHERE user_id=$1 ORDER BY UploadedAt DESC", ctx.Value(add.LogginKey))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("no orders for user have been found %w", err)
		}
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&order.Number, &order.Status, &order.UploadedAt, &order.Accrual)
		if err != nil {
			return
		}

		*orders = append(*orders, order)
	}

	err = rows.Err()
	if err != nil {
		return
	}

	return
}

func (db *PostgreSQL) GetSpendOrders(ctx context.Context, orders *[]add.OrderSpend) (err error) {
	var order add.OrderSpend
	rows, err := db.dbConn.QueryContext(ctx, "SELECT id, ProcessedAt, sum FROM order_spend WHERE user_id=$1 ORDER BY ProcessedAt DESC", ctx.Value(add.LogginKey))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("no orders for user have been found %w", err)
		}
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&order.Number, &order.ProcessedAt, &order.Sum)
		if err != nil {
			return
		}

		*orders = append(*orders, order)
	}

	err = rows.Err()
	if err != nil {
		return
	}
	return
}
