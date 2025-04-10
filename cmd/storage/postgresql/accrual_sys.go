package storage

import (
	"context"
	"time"

	add "github.com/Tanya1515/gophermarket/cmd/additional"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func (db *PostgreSQL) ProcessAccOrder(ctx context.Context, order add.Order) (err error) {
	var userID int
	var currentOrderStatus string
	tx, err := db.dbConn.Begin()
	if err != nil {
		return
	}

	err = tx.QueryRowContext(ctx, "SELECT status FROM orders WHERE id=$1", order.Number).Scan(&currentOrderStatus)
	if err != nil {
		tx.Rollback()
		return
	}

	if currentOrderStatus != "PROCESSED" {
		if (order.Status == "PROCESSED") && (order.Accrual > 0) {

			err = tx.QueryRowContext(ctx, "UPDATE orders SET accrual=$1, status=$2 WHERE id=$3 RETURNING user_id", order.Accrual, order.Status, order.Number).Scan(&userID)
			if err != nil {
				tx.Rollback()
				return
			}

			_, err = tx.ExecContext(ctx, "UPDATE users SET sum=sum+$1 WHERE id=$2", order.Accrual, userID)
			if err != nil {
				tx.Rollback()
				return
			}

		} else {
			_, err = tx.ExecContext(ctx, "UPDATE orders SET status=$1 WHERE id=$2", order.Status, order.Number)
			if err != nil {
				tx.Rollback()
				return
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return
	}

	return
}

func (db *PostgreSQL) StartProcessingUserOrder(ctx context.Context, logger zap.SugaredLogger, result chan add.OrderAcc) {

	var order add.OrderAcc
	g := new(errgroup.Group)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(40 * time.Second):
			rows, err := db.dbConn.Query("SELECT id, accrual FROM orders WHERE status=$1", "")
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
		default:
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
	
}
