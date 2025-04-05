package intaccrual

import (
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
}

func (ac *AccrualSystem) AccrualMain() {
	semaphoreAccrual := NewSemaphore(ac.Limit)

	ac.SemaphoreAccrual = semaphoreAccrual

	processOrderChan := make(chan add.OrderAcc, ac.Limit)
	// orderIDChan := make(chan string, ac.Limit)
	// resultOrderChan := make(chan add.OrderAcc, ac.Limit)

	go ac.Storage.StartProcessingUserOrder(ac.Logger, processOrderChan)

	go ac.SendOrder(processOrderChan) //, orderIDChan)

	// go ac.GetOrderFromAccrual(orderIDChan, resultOrderChan)

	// for order := range resultOrderChan {
	// 	order := order

	// 	go func() {
	// 		for {
	// 			var orderResult add.Order
	// 			orderResult.Number = order.Order
	// 			orderResult.Status = order.Status
	// 			orderResult.Accrual = order.Accrual
	// 			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	// 			defer cancel()
	// 			err := ac.Storage.ProcessAccOrder(ctx, orderResult)
	// 			if err == nil {
	// 				ac.Logger.Infof("Save recent information about order: %s", order.Order)
	// 				break
	// 			}
	// 			time.Sleep(5 * time.Microsecond)
	// 			ac.Logger.Errorf("Error while updating order %s to database: %s", order.Order, err)
	// 		}
	// 	}()
	// }

}
