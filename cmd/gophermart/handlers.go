package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	add "github.com/Tanya1515/gophermarket/cmd/additional"
)

func (GM *Gophermarket) RegisterUser() http.HandlerFunc {
	registerUser := func(rw http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		var err error
		var user add.User

		_, err = buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while reading new user credentials: %s", err.Error()), http.StatusBadRequest)
			GM.logger.Errorf("Error while reading user credentials: ", err.Error())
			return
		}

		err = json.Unmarshal(buf.Bytes(), &user)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while unmarshalling request body for processing new user: %s", err.Error()), http.StatusInternalServerError)
			GM.logger.Errorf("Error while unmarshalling request body for processing new order: ", err.Error())
			return
		}

		if (user.Login == "") || (user.Password == "") {
			http.Error(rw, "Error while getting user credentials", http.StatusBadRequest)
			GM.logger.Errorf("Error while getting user credentials")
			return
		}

		err = GM.storage.CheckUserLogin(user.Login)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while checking user login: %s", err), http.StatusConflict)
			GM.logger.Errorf("Error while checking user login: ", err)
			return
		}

		err = GM.storage.RegisterNewUser(user)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while regestering new user: %s", err.Error()), http.StatusInternalServerError)
			GM.logger.Errorf("Error while regestering new user: ", err.Error())
			return
		}

		JWTtoken, err := add.GenerateToken(user, GM.secretKey)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while generating jwt token: %s", err.Error()), http.StatusInternalServerError)
			GM.logger.Errorf("Error while generating jwt token: ", err.Error())
			return
		}

		cookie := &http.Cookie{
			Name:   "token",
			Value:  JWTtoken,
			MaxAge: 3600,
		}

		http.SetCookie(rw, cookie)

		rw.WriteHeader(http.StatusOK)

		rw.Write([]byte(fmt.Sprintf("New user %s sucessfully registered and authenticated!", user.Login)))

	}
	return http.HandlerFunc(registerUser)
}

func (GM *Gophermarket) AuthentificateUser() http.HandlerFunc {
	authentificateUser := func(rw http.ResponseWriter, r *http.Request) {

		var buf bytes.Buffer
		var err error
		var user add.User

		_, err = buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while reading user credentials: %s", err.Error()), http.StatusBadRequest)
			GM.logger.Errorf("Error while reading user credentials: ", err.Error())
			return
		}

		err = json.Unmarshal(buf.Bytes(), &user)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while unmarshalling request body for processing user: %s", err.Error()), http.StatusInternalServerError)
			GM.logger.Errorf("Error while unmarshalling request body for processing new order: ", err.Error())
			return
		}

		if (user.Login == "") || (user.Password == "") {
			http.Error(rw, "Error while getting user credentials", http.StatusBadRequest)
			GM.logger.Errorf("Error while getting user credentials")
			return
		}

		ok, err := GM.storage.CheckUser(user.Login, user.Password)
		if (err != nil) && ok {
			http.Error(rw, fmt.Sprintf("Error while processing user with login %s: %s", user.Login, err.Error()), http.StatusInternalServerError)
			GM.logger.Errorf("Error while processing user with login ", user.Login, ": ", err.Error())
			return
		}

		if !ok {
			http.Error(rw, fmt.Sprintf("User %s login/password is incorrect", user.Login), http.StatusUnauthorized)
			GM.logger.Errorf("User ", user.Login, " login/password is incorrect")
			return
		}

		JWTtoken, err := add.GenerateToken(user, GM.secretKey)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while generating jwt token: %s", err.Error()), http.StatusInternalServerError)
			GM.logger.Errorf("Error while generating jwt token: ", err.Error())
			return
		}

		cookie := &http.Cookie{
			Name:   "token",
			Value:  JWTtoken,
			MaxAge: 3600,
		}

		http.SetCookie(rw, cookie)

		rw.WriteHeader(http.StatusOK)

		rw.Write([]byte(fmt.Sprintf("User %s successfully authentificated!", user.Login)))

	}
	return http.HandlerFunc(authentificateUser)
}

func (GM *Gophermarket) GetOrdersInfobyUser() http.HandlerFunc {
	getOrdersInfobyUser := func(rw http.ResponseWriter, r *http.Request) {
		login := r.Context().Value(add.LogginKey)

		orders := make([]add.Order, 0, 10)

		err := GM.storage.GetAllOrders(&orders, login.(string))
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while getting order info for user %s: %s", login, err.Error()), http.StatusInternalServerError)
			GM.logger.Errorf("Error while getting order info for user ", login, ": ", err.Error())
			return
		}

		if len(orders) == 0 {
			rw.WriteHeader(http.StatusNoContent)
			return
		}
		ordersByte, err := json.Marshal(orders)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while unmarshalling orders to bytes: %s", err.Error()), http.StatusInternalServerError)
			GM.logger.Errorf("Error while getting order info for user: ", err.Error())
			return
		}
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)

		rw.Write(ordersByte)

	}
	return http.HandlerFunc(getOrdersInfobyUser)
}

func (GM *Gophermarket) AddOrdersInfobyUser() http.HandlerFunc {
	addOrdersInfobyUser := func(rw http.ResponseWriter, r *http.Request) {
		var err error

		login := r.Context().Value(add.LogginKey)

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(rw, "Unable to read request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		orderID := string(body)

		GM.logger.Infoln("OrderID: ", orderID)

		if !add.CheckOrderNumber(orderID) {
			http.Error(rw, "Order number is invalid", http.StatusUnprocessableEntity)
			GM.logger.Errorln("Order number is invalid")
			return
		}
		err = GM.storage.AddNewOrder(login.(string), orderID)
		if err != nil {
			if strings.Contains(err.Error(), "error: order with number") {
				http.Error(rw, err.Error(), http.StatusConflict)
				return
			} else if strings.Contains(err.Error(), "the order with number") {
				rw.WriteHeader(http.StatusOK)
				return
			} else {
				http.Error(rw, fmt.Sprintf("Error while adding new order to database: %s", err.Error()), http.StatusInternalServerError)
				return
			}
		}

		rw.WriteHeader(http.StatusAccepted)

		rw.Write([]byte(fmt.Sprintf("New order %s is processing", orderID)))
	}
	return http.HandlerFunc(addOrdersInfobyUser)
}

func (GM *Gophermarket) GetUserBalance() http.HandlerFunc {
	getUserBalance := func(rw http.ResponseWriter, r *http.Request) {

		login := r.Context().Value(add.LogginKey)

		balance, err := GM.storage.GetUserBalance(login.(string))
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while getting user %s balance: %s", login, err.Error()), http.StatusInternalServerError)
			GM.logger.Errorf("Error while unmarshalling request body for processing new order: ", err.Error())
			return
		}

		balanceByte, err := json.Marshal(balance)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while unmarshalling request body for processing new order: %s", err.Error()), http.StatusInternalServerError)
			GM.logger.Errorf("Error while unmarshalling request body for processing new order: ", err.Error())
			return
		}
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)

		rw.Write(balanceByte)

	}
	return http.HandlerFunc(getUserBalance)
}

func (GM *Gophermarket) GetUserWastes() http.HandlerFunc {
	getUserWastes := func(rw http.ResponseWriter, r *http.Request) {
		login := r.Context().Value(add.LogginKey)
		orders := make([]add.OrderSpend, 0, 10)

		err := GM.storage.GetSpendOrders(&orders, login.(string))
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while getting order with points to spend info for user %s: %s", login, err.Error()), http.StatusInternalServerError)
			GM.logger.Errorf("Error while getting order with points to spend info for user ", login, ": ", err.Error())
			return
		}

		if len(orders) == 0 {
			rw.WriteHeader(http.StatusNoContent)
			return
		}
		ordersByte, err := json.Marshal(orders)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while unmarshalling orders with points to spend to bytes: %s", err.Error()), http.StatusInternalServerError)
			GM.logger.Errorf("Error while getting order with points to spend info for user: ", err.Error())
			return
		}
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)

		rw.Write(ordersByte)

	}
	return http.HandlerFunc(getUserWastes)
}

func (GM *Gophermarket) PayByPoints() http.HandlerFunc {
	payByPoints := func(rw http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		var err error
		var order add.OrderSpend
		login := r.Context().Value(add.LogginKey)

		_, err = buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while reading order number for processing it: %s", err.Error()), http.StatusBadRequest)
			GM.logger.Errorf("Error while reading order number for processing it: ", err.Error())
			return
		}

		err = json.Unmarshal(buf.Bytes(), &order)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while unmarshalling request body for processing new order: %s", err.Error()), http.StatusInternalServerError)
			GM.logger.Errorf("Error while unmarshalling request body for processing new order: ", err.Error())
			return
		}

		if !add.CheckOrderNumber(order.Number) {
			http.Error(rw, "Order number is invalid", http.StatusUnprocessableEntity)
			GM.logger.Errorln("Order number is invalid")
			return
		}

		err = GM.storage.ProcessPayPoints(order, login.(string))
		if err != nil {

			if strings.Contains(err.Error(), "violates check constraint \"valid_sum\"") {
				http.Error(rw, fmt.Sprintf("Users's balance is not enough: %s", err.Error()), http.StatusPaymentRequired)
				GM.logger.Errorln("Users's balance is not enough: ", err.Error())
				return
			}

			http.Error(rw, fmt.Sprintf("Error while processing order for spending points: %s, the result is: %v", err.Error(), strings.Contains(err.Error(), "violates check constraint \"valid_sum\"")), http.StatusInternalServerError)
			GM.logger.Errorln("Error while processing order for spending points: ", err.Error())
			return
		}

		rw.WriteHeader(http.StatusOK)

	}
	return http.HandlerFunc(payByPoints)
}
