package additional

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

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

	ok := (len(arrayDigits)) % 2
	for key, value := range arrayDigits {
		if (ok == 1) && ((key % 2) == ok) && (key != 0) {
			value = value * 2
			if value > 9 {
				value = value - 9
			}
		}

		if (ok == 0) && (((key + 1) % 2) == ok) && (key != 0) {
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
