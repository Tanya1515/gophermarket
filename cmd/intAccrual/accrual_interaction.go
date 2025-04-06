package intaccrual

import (
	"encoding/json"
	"time"

	"github.com/go-resty/resty/v2"

	add "github.com/Tanya1515/gophermarket/cmd/additional"
)

func (ac *AccrualSystem) SendOrder(inputChan chan add.OrderAcc, resultChan chan string) {

	// select + case
	for order := range inputChan {
		order := order

		// wg.Add(1)
		go func() {
			// defer wg.Done()
			client := resty.New()

			ordersByte, err := json.Marshal(order)
			if err != nil {
				ac.Logger.Errorf("Error while marshalling order %s: %s", order.Order, err)
			}
			ac.SemaphoreAccrual.Lock()
			defer ac.SemaphoreAccrual.Acquire()
			var timer time.Duration = 5
			for i := 0; i < 3; i = i + 1 {
				resp, err := client.R().SetHeader("Content-Type", "application/json").
					SetBody(ordersByte).
					Post(ac.AccrualAddress + "/api/orders")

				if resp.StatusCode() == 202 || resp.StatusCode() == 409 || resp.StatusCode() == 200 {
					ac.Logger.Infof("Order has been sent: %s", order.Order)
					break
				}

				if resp.StatusCode() == 429 {
					timer += time.Duration(i)
					ac.Logger.Infof("Can not send order %s to accrual system, the accrual system is overloaded", order.Order)
				}

				time.Sleep(timer * time.Second)
				if err != nil {
					ac.Logger.Errorf("Error while sending order %s to accrual system: %s", order.Order, err)
				}
			}

			resultChan <- order.Order
		}()

	}
}

func (ac *AccrualSystem) GetOrderFromAccrual(inputChan chan string, resultChan chan add.OrderAcc) {

	var order add.OrderAcc

	// select + case
	for orderID := range inputChan {
		orderID := orderID
		// wg.Add(1)
		go func() {
			// defer wg.Done()
			client := resty.New()
			ac.SemaphoreAccrual.Lock()
			defer ac.SemaphoreAccrual.Acquire()
			// ограничить количество попыток отправки заказа в accrual
			for i := 0; i < 3; i = i + 1 {
				resp, err := client.R().Get(ac.AccrualAddress + "/api/orders/" + orderID)

				if err != nil {
					time.Sleep(5 * time.Second)
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
