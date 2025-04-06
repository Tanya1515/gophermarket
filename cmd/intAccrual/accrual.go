package intaccrual

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	add "github.com/Tanya1515/gophermarket/cmd/additional"
	storage "github.com/Tanya1515/gophermarket/cmd/storage"
)

type AccrualSystem struct {
	AccrualAddress   string
	Storage          storage.StorageInterface
	Limit            int
	Logger           zap.SugaredLogger
	SemaphoreAccrual *Semaphore
	WG       *sync.WaitGroup
}

func (ac *AccrualSystem) AccrualMain(ctx context.Context) {
	semaphoreAccrual := NewSemaphore(ac.Limit)

	ac.SemaphoreAccrual = semaphoreAccrual

	processOrderChan := make(chan add.OrderAcc, ac.Limit)
	orderIDChan := make(chan string, ac.Limit)
	resultOrderChan := make(chan add.OrderAcc, ac.Limit)

	go ac.Storage.StartProcessingUserOrder(ctx, ac.Logger, processOrderChan)

	go ac.SendOrder(ctx, processOrderChan, orderIDChan)

	go ac.GetOrderFromAccrual(ctx, orderIDChan, resultOrderChan)

	select {
	case order := <-resultOrderChan:
		for i := 0; i < 3; i = i + 1 {
			var orderResult add.Order
			orderResult.Number = order.Order
			orderResult.Status = order.Status
			orderResult.Accrual = order.Accrual

			queryCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			err := ac.Storage.ProcessAccOrder(queryCtx, orderResult)
			if err == nil {
				ac.Logger.Infof("Save recent information about order: %s", order.Order)
				break
			}
			time.Sleep(5 * time.Second)
			ac.Logger.Errorf("Error while updating order %s to database: %s", order.Order, err)
		}
	case <-ctx.Done():
		return
	}
}
