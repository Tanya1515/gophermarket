package additional

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Status string

const (
	New        Status = "NEW"
	Processing Status = "PROCESSING"
	Invalid    Status = "INVALID"
	Processed  Status = "PROCESSED"
)

type OrderSpend struct {
	Number      string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"ProcessedAt"`
}

type Order struct {
	Number     string    `json:"number"`
	Status     Status    `json:"status"`
	Accrual    float64   `json:"accrual"`
	UploadedAt time.Time `json:"UploadedAt"`
}

type OrderAcc struct {
	Order   string  `json:"order"`
	Accrual float64 `json:"accrual"`
	Status  Status  `json:"status"`
}

const TokenExp = time.Hour

type Claims struct {
	jwt.RegisteredClaims
	UserLogin    string
	UserPassword string
}

func GenerateToken(user User, secretKey string) (JWTtoken string, err error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserLogin:    user.Login,
		UserPassword: user.Password,
	})

	JWTtoken, err = token.SignedString([]byte(secretKey))
	if err != nil {
		return
	}

	return
}

func CheckOrderNumber(orderNumber string) bool {
	var num, sum int64
	arrayDigits := make([]int64, 0, 10)

	res64, err := strconv.ParseInt(orderNumber, 10, 64)
	if err != nil {
		panic(err)
	}
	for res64 > 0 {
		num = res64 % 10
		res64 = res64 / 10
		arrayDigits = append(arrayDigits, num)
	}

	for key, value := range arrayDigits {
		if ((key + 1) % 2) == ((len(arrayDigits)) % 2) {
			value = value * 2
			if value > 9 {
				value = value - 9
			}
		}
		sum += value
	}

	if sum%10 == 0 {
		return true
	} else {
		return false
	}
}
