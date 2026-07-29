package usecase

import (
	"context"
	"strings"

	"github.com/barnigator/eshop-api-gateway/internal/domain"
)

type SSOClient interface {
	Register(ctx context.Context, email, password string) (int64, error)
	Login(ctx context.Context, email, password string) (string, error)
}

type UseCase struct {
	ssoClient SSOClient
}

func New(ssoClient SSOClient) *UseCase {
	return &UseCase{
		ssoClient: ssoClient,
	}
}

func (uc *UseCase) Register(ctx context.Context, email, password string) (int64, error) {
	cleanEmail := strings.TrimSpace(email)
	if cleanEmail == "" {
		return 0, domain.ErrEmailRequired
	}

	if password == "" {
		return 0, domain.ErrPasswordRequired
	}

	return uc.ssoClient.Register(ctx, cleanEmail, password)
}

func (uc *UseCase) Login(ctx context.Context, email, password string) (string, error) {
	cleanEmail := strings.TrimSpace(email)
	if cleanEmail == "" {
		return "", domain.ErrEmailRequired
	}

	if password == "" {
		return "", domain.ErrPasswordRequired
	}

	return uc.ssoClient.Login(ctx, cleanEmail, password)
}
