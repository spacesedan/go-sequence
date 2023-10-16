package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Pallinder/go-randomdata"
	rend "github.com/unrolled/render"
)

var render = rend.New()

func generateUserCookie() (string, *http.Cookie) {
	randomNumber := randomdata.Number(42069)
	randomName := randomdata.SillyName()

	// construct the username
	userName := fmt.Sprintf("%s%d", randomName, randomNumber)

	return userName, &http.Cookie{
		Name:     "username",
		Value:    userName,
		Path:     "/",
		MaxAge:   10 * 365 * 24 * 60 * 60,
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
func createWebsocketConnectionString(lobbyId string) string {
	u := url.URL{
		Scheme: "ws",
		Host:   "localhost:42069",
		Path:   "/lobby/ws",
	}
	q := u.Query()
	q.Set("lobby-id", lobbyId)
	u.RawQuery = q.Encode()

	return u.String()
}
