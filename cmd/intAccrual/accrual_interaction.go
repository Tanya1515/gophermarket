package intaccrual

import (
	"encoding/json"
	"time"

	"github.com/go-resty/resty/v2"

	add "github.com/Tanya1515/gophermarket/cmd/additional"
)

func (ac *AccrualSystem) SendOrder(inputChan chan add.OrderAcc){ //, resultChan chan string) {

	for order := range inputChan {
		order := order

		go func() {
			client := resty.New()

			ordersByte, err := json.Marshal(order)
			if err != nil {
				ac.Logger.Errorf("Error while marshalling order %s: %s", order.Order, err)
			}
			ac.SemaphoreAccrual.Lock()
			defer ac.SemaphoreAccrual.Acquire()
			for {
				_, err = client.R().SetHeader("Content-Type", "application/json").
					SetBody(ordersByte).
					Post(ac.AccrualAddress + "/api/orders")

				if err == nil {
					ac.Logger.Infof("Send order: %s", order.Order)
					break
				}

				time.Sleep(5 * time.Microsecond)
				ac.Logger.Errorf("Error while sending order %s to accrual system: %s", order.Order, err)

			}

			// resultChan <- order.Order
		}()

	}
}

func (ac *AccrualSystem) GetOrderFromAccrual(inputChan chan string, resultChan chan add.OrderAcc) {

	var order add.OrderAcc

	for orderID := range inputChan {
		orderID := orderID

		go func() {
			client := resty.New()
			ac.SemaphoreAccrual.Lock()
			defer ac.SemaphoreAccrual.Acquire()
			for {
				resp, err := client.R().Get(ac.AccrualAddress + "/api/orders/" + orderID)

				if err != nil {
					time.Sleep(5 * time.Microsecond)
					ac.Logger.Errorf("Error while getting order %s from accrual system: %s", orderID, err)
					continue
				}

				err = json.Unmarshal(resp.Body(), &order)
				if err != nil {
					ac.Logger.Errorf("Error while unmarshalling order %s: %s", order.Order, err)
				}

				if (order.Status == "PROCESSED") || (order.Status == "INVALID") {
					resultChan <- order
					break
				}
			}

		}()

	}
}
