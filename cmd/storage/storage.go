package storage

import (
	"context"

	add "github.com/Tanya1515/gophermarket/cmd/additional"
	"go.uber.org/zap"
)

type StorageInterface interface {
	Init() error

	RegisterNewUser(ctx context.Context, user add.User) error

	AddNewOrder(ctx context.Context, orderNumber string) error

	CheckUserLogin(ctx context.Context, login string) error

	CheckUser(ctx context.Context, login, password string) (bool, error)

	CheckUserJWT(ctx context.Context, login string) (string, error)

	GetUserBalance(ctx context.Context) (add.Balance, error)

	GetAllOrders(ctx context.Context, orders *[]add.Order) error

	GetSpendOrders(ctx context.Context, orders *[]add.OrderSpend) error

	ProcessPayPoints(ctx context.Context, order add.OrderSpend) error

	StartProcessingUserOrder(logger zap.SugaredLogger, result chan add.OrderAcc)

	ProcessAccOrder(ctx context.Context, order add.Order) error
}
