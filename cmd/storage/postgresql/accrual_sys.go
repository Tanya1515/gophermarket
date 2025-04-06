package storage

import (
	"context"

	add "github.com/Tanya1515/gophermarket/cmd/additional"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func (db *PostgreSQL) ProcessAccOrder(ctx context.Context, order add.Order) (err error) {
	var userID int
	tx, err := db.dbConn.Begin()
	if err != nil {
		return
	}

	if (order.Status == "PROCESSED") && (order.Accrual > 0) {

		err = tx.QueryRow("UPDATE orders SET accrual=$1, status=$2 WHERE id=$3 RETURNING user_id", order.Accrual, order.Status, order.Number).Scan(&userID)
		if err != nil {
			tx.Rollback()
			return
		}

		_, err = tx.Exec("UPDATE users SET sum=sum+$1 WHERE id=$2", order.Accrual, userID)
		if err != nil {
			tx.Rollback()
			return
		}

	} else {
		_, err = tx.Exec("UPDATE orders SET status=$1 WHERE id=$2", order.Status, order.Number)
		if err != nil {
			tx.Rollback()
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return
	}

	return
}

func (db *PostgreSQL) StartProcessingUserOrder(logger zap.SugaredLogger, result chan add.OrderAcc) {

	var order add.OrderAcc
	g := new(errgroup.Group)

	// select - ожидаем завершения контекста или берем запрос
	for {
		rows, err := db.dbConn.Query("SELECT id, accrual FROM orders WHERE status=$1", "NEW")
		if err != nil {
			logger.Errorf("Error while getting new orders: ", err)
			continue
		}

		for rows.Next() {

			err := rows.Scan(&order.Order, &order.Accrual)
			if err != nil {
				logger.Errorf("Error while scanning new order: ", err)
			}

			g.Go(func() error {

				order.Status = "PROCESSING"
				_, err = db.dbConn.Exec("UPDATE orders SET status=$1 WHERE id=$2 AND status=$3", order.Status, order.Order, "NEW")
				if err != nil {
					return err
				}
				result <- order
				return nil
			})
		}

		err = rows.Err()
		if err != nil {
			logger.Errorf("Error while reading rows of new orders: ", err)
		}

		rows.Close()
		if err := g.Wait(); err != nil {
			logger.Errorf("Error from goroutine: ", err)
		}
	}
}
