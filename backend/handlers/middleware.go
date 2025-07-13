package handlers

import (
	"context"
	"net/http"
	"slices"
	"strings"
)

type contextKey string

const contextKeyUser = contextKey("Context_key_User")

var publicRoutes = []string{
	"/",
	"/signup",
	"/signin",
	"/public/",
}

func isPublicPath(path string) bool {
	if slices.Contains(publicRoutes, path) || strings.HasPrefix(path, "/public/") {
		return true
	}
	return false
}

func (App *WebApp) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isPublicPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		cookie, err := r.Cookie("session_id")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		session, errCode, err := App.Sessions.GetUserBySession(cookie.Value)
		if err != nil {
			w.WriteHeader(errCode)
			return
		}
		user, err := App.Users.GetUserByID(session.UserID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// log.Println("auth", contextKeyUser, user)

		ctx := context.WithValue(r.Context(), contextKeyUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
