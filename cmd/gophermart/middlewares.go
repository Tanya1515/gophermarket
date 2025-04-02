package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"

	add "github.com/Tanya1515/gophermarket/cmd/additional"
)

func (GM *Gophermarket) MiddlewareCheckUser(h http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("token")
		if err != nil {
			http.Error(rw, "", http.StatusUnauthorized)
			GM.logger.Errorf("Error while processing cookie")
			return
		}

		claims := &add.Claims{}
		jwt.ParseWithClaims(token.Value, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(GM.secretKey), nil
		})

		err = claims.Valid()

		if err != nil {
			http.Error(rw, fmt.Sprintf("User is not anuthorized: %s", err), http.StatusUnauthorized)
			GM.logger.Errorf("User is not anuthorized: ", err.Error())
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		id, err := GM.storage.CheckUserJWT(ctx, claims.UserLogin)
		if err != nil {
			http.Error(rw, "User is not anuthorized", http.StatusUnauthorized)
			GM.logger.Errorf("User is not anuthorized")
			return
		}

		h.ServeHTTP(rw, r.WithContext(context.WithValue(r.Context(), add.LogginKey, id)))

	}
}
