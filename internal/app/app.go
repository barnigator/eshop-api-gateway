package app

import (
	"fmt"
	"log/slog"

	"github.com/barnigator/eshop-api-gateway/internal/client/grpc/sso"
	"github.com/barnigator/eshop-api-gateway/internal/config"
	"github.com/barnigator/eshop-api-gateway/internal/http/handler"
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
		a.cfg.Clients.SSO.AppId)
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

	authHandler := handler.New(uc)

	return nil
}
