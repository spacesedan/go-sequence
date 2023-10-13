package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Pallinder/go-randomdata"
	rend "github.com/unrolled/render"
)

var render = rend.New()

func generateUserCookie() (string, *http.Cookie) {
	randomNumber := randomdata.Number(42069)
	randomName := randomdata.SillyName()

	// construct the username
	userName := fmt.Sprintf("%d%s", randomNumber, randomName)

	return userName,  &http.Cookie{
		Name:     "username",
		Value:    userName,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}
}

func getUsernameFromCookie(r *http.Request) (string, error) {
	userCookie, err := r.Cookie("username")
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		return "", err
	}

	return userCookie.Value, nil

}
