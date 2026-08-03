package server

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type AuthHandler interface {
	Register(http.ResponseWriter, *http.Request)
	Login(http.ResponseWriter, *http.Request)
}

type Server struct {
	httpServer *http.Server
}

func New(handler http.Handler, port int, timeout time.Duration) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:              fmt.Sprintf(":%d", port),
			Handler:           handler,
			ReadHeaderTimeout: timeout},
	}
}

func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
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
