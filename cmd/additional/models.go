package additional

import (
	"github.com/golang-jwt/jwt/v4"
	"time"
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