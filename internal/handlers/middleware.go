package handlers

import "net/http"

func CheckUsernameCookie(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("username")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			next.ServeHTTP(w, r)
		}
	}

	return http.HandlerFunc(fn)
}
