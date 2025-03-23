package main

import (
	"flag"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	_ "github.com/Tanya1515/gophermarket/cmd/additional"
	acc "github.com/Tanya1515/gophermarket/cmd/intAccrual"
	storage "github.com/Tanya1515/gophermarket/cmd/storage"
	psql "github.com/Tanya1515/gophermarket/cmd/storage/postgresql"
)

type Gophermarket struct {
	storage   storage.StorageInterface
	logger    zap.SugaredLogger
	secretKey string
}

func init() {
	marketAddressFlag = flag.String("a", "localhost:8081", "gophermarket address")
	storageURLFlag = flag.String("d", "localhost:5432", "database url")
	accrualSystemAddressFlag = flag.String("r", "localhost:8080", "acccrual system address")
	accrualLimitFlag = flag.Int("l", 100, "request limits for accrual system")
}

var (
	marketAddressFlag        *string
	storageURLFlag           *string
	accrualSystemAddressFlag *string
	accrualLimitFlag         *int
)

func main() {
	var GM Gophermarket

	var Accrual acc.AccrualSystem

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	MarketLogger := *logger.Sugar()

	loggerApp := MarketLogger

	defer loggerApp.Sync()

	GM.logger = loggerApp
	flag.Parse()

	marketAddress, ok := os.LookupEnv("RUN_ADDRESS")
	if !ok {
		marketAddress = *marketAddressFlag
	}

	storageAddress, ok := os.LookupEnv("DATABASE_URI")
	if !ok {
		storageAddress = *storageURLFlag
	}

	accrualSystemAddress, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS")
	if !ok {
		accrualSystemAddress = *accrualSystemAddressFlag
	}

	accrualLimits := *accrualLimitFlag
	limits, ok := os.LookupEnv("ACCRUAL_LIMIT_REQUESTS")
	if ok {
		accrualLimits, err = strconv.Atoi(limits)
		if err != nil {
			GM.logger.Error("Error while converting accrualLimits from string to integer")
		}
	}

	storageArgs := strings.Split(strings.Split(storageAddress, "//")[1], ":")
	storagePass := strings.Split(storageArgs[1], "@")
	storageAddr := strings.Split(storagePass[1], "/")
	storageDB := strings.Split(storageAddr[1], "?")

	Storage := &psql.PostgreSQL{Address: storageAddr[0], UserName: storageArgs[0], Password: storagePass[0], DBName: storageDB[0]}

	// Storage := &psql.PostgreSQL{Address: "localhost", UserName: "collector", Password: "postgres", DBName: "gophermarket"}
	GM.storage = Storage

	Accrual.Logger = loggerApp
	Accrual.Storage = Storage
	Accrual.AccrualAddress = accrualSystemAddress
	Accrual.Limit = accrualLimits

	err = GM.storage.Init()
	if err != nil {
		panic(err)
	}

	GM.logger.Infow(
		"Gophermarket starts working",
		"address: ", marketAddress,
	)

	GM.secretKey = "secretKey"

	go Accrual.AccrualMain()

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Post("/api/user/register", GM.RegisterUser())
		r.Post("/api/user/login", GM.AuthentificateUser())
		r.Post("/api/user/orders", GM.MiddlewareCheckUser(GM.AddOrdersInfobyUser()))
		r.Get("/api/user/orders", GM.MiddlewareCheckUser(GM.GetOrdersInfobyUser()))
		r.Get("/api/user/balance", GM.MiddlewareCheckUser(GM.GetUserBalance()))
		r.Get("/api/user/withdrawals", GM.MiddlewareCheckUser(GM.GetUserWastes()))
		r.Post("/api/user/balance/withdraw", GM.MiddlewareCheckUser(GM.PayByPoints()))

	})

	err = http.ListenAndServe(marketAddress, r)
	if err != nil {
		GM.logger.Fatalw(err.Error(), "event", "start server")
	}
}
