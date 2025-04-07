package intaccrual

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"go.uber.org/zap"

	add "github.com/Tanya1515/gophermarket/cmd/additional"
	storage "github.com/Tanya1515/gophermarket/cmd/storage"
	"github.com/go-resty/resty/v2"
)

type AccrualSystem struct {
	AccrualAddress      string
	Storage             storage.StorageInterface
	Limit               int
	Logger              zap.SugaredLogger
	SemaphoreAccrual    *Semaphore
	WG                  *sync.WaitGroup
	GophermarketAddress string
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
		var orderResult add.Order
		orderResult.Number = order.Order
		orderResult.Status = order.Status
		orderResult.Accrual = order.Accrual

		ordersByte, err := json.Marshal(orderResult)
		if err != nil {
			ac.Logger.Errorf("Error while marshalling order %s: %s", order.Order, err)
		}

		client := resty.New()

		for i := 0; i < 3; i = i + 1 {
			resp, err := client.R().SetHeader("Content-Type", "application/json").
				SetBody(ordersByte).
				Post(ac.GophermarketAddress + "/api/accrual/orders")

			if resp.StatusCode() == 200 {
				ac.Logger.Infof("Order has been sent to gophermarket: %s", order.Order)
				break
			}

			time.Sleep(5 * time.Second)
			if err != nil {
				ac.Logger.Errorf("Error while sending order %s to gophermarket: %s", order.Order, err)
			}
		}
	case <-ctx.Done():
		return
	}
}
