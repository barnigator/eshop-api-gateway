package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/barnigator/eshop-api-gateway/internal/client/grpc/sso"
	"github.com/barnigator/eshop-api-gateway/internal/config"
	"github.com/barnigator/eshop-api-gateway/internal/http/handler"
	"github.com/barnigator/eshop-api-gateway/internal/http/server"
	"github.com/barnigator/eshop-api-gateway/internal/usecase"
)

type App struct {
	cfg *config.Config
	log *slog.Logger
}

func New(cfg *config.Config, log *slog.Logger) *App {
	return &App{
		cfg: cfg,
		log: log,
	}
}

func (a *App) Run() error {
	a.log.Info(
		"starting application",
		slog.Int("http_port", a.cfg.HTTP.Port),
		slog.String("sso_address", a.cfg.Clients.SSO.Address),
	)

	ssoClient, err := sso.New(
		a.cfg.Clients.SSO.Address,
		a.cfg.Clients.SSO.AppId,
	)
	if err != nil {
		return fmt.Errorf("initialize sso client: %w", err)
	}
	defer func() {
		a.log.Info("closing sso grpc client")
		closeErr := ssoClient.Close()
		if closeErr != nil {
			a.log.Error(
				"failed to close sso grpc client",
				slog.String("error", closeErr.Error()))
		}
		a.log.Info("sso grpc client closed")
	}()

	uc := usecase.New(ssoClient)

	h := handler.New(uc)

	router := server.NewRouter(h)

	s := server.New(router, a.cfg.HTTP.Port, a.cfg.HTTP.Timeout)

	shutdownSignalCtx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- s.Run()
	}()

	select {
	case serverErr := <-serverErrors:
		if serverErr != nil && !errors.Is(serverErr, http.ErrServerClosed) {
			return fmt.Errorf("http server stopped unexpectedly: %w", serverErr)
		}

		return nil
	case <-shutdownSignalCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			a.cfg.HTTP.ShutdownTimeout,
		)
		defer cancel()

		err = s.Shutdown(shutdownCtx)
		if err != nil {
			return fmt.Errorf("server shutdown: %w", err)
		}

		serverErr := <-serverErrors
		if serverErr != nil && !errors.Is(serverErr, http.ErrServerClosed) {
			return fmt.Errorf("http server stopped unexpectedly: %w", serverErr)
		}
	}

	return nil
}
