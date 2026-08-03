package server

import "net/http"

type AuthHandler interface {
	Register(http.ResponseWriter, *http.Request)
	Login(http.ResponseWriter, *http.Request)
}

func NewRouter(authHandler AuthHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc(
		"POST /api/v1/auth/register",
		authHandler.Register,
	)
	mux.HandleFunc(
		"POST /api/v1/auth/login",
		authHandler.Login,
	)

	return mux
}
