package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v4"

	add "github.com/Tanya1515/gophermarket/cmd/additional"
)

type contextKey string

const logginKey contextKey = "userLogin"

func (GM *Gophermarket) MiddlewareCheckUser(h http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("token")
		if err != nil {
			http.Error(rw, "", http.StatusBadRequest)
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
		err = GM.storage.CheckUserJWT(claims.UserLogin)
		if err != nil {
			http.Error(rw, "User is not anuthorized", http.StatusUnauthorized)
			GM.logger.Errorf("User is not anuthorized")
			return
		}

		h.ServeHTTP(rw, r.WithContext(context.WithValue(r.Context(), logginKey, claims.UserLogin)))

	}
}
