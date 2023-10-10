package handlers

import (
	"errors"
	"net/http"

	rend "github.com/unrolled/render"
)

var render = rend.New()

func getUsernameFromCookie(r *http.Request) (string, error) {
	userCookie, err := r.Cookie("username")
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		return "", err
	}

	return userCookie.Value, nil

}
