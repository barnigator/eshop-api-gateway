package sso

import (
	"context"
	"fmt"

	"github.com/barnigator/eshop-api-gateway/internal/domain"
	ssov1 "github.com/barnigator/protos/gen/go/sso/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Client struct {
	conn   *grpc.ClientConn
	client ssov1.AuthClient
	appID  int32
}

func New(address string, appID int32) (*Client, error) {
	if address == "" {
		return nil, ErrSsoAddressIsEmpty
	}

	if appID <= 0 {
		return nil, ErrAppIDMustBePositive
	}

	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("create sso grpc client: %w", err)
	}

	client := ssov1.NewAuthClient(conn)

	return &Client{
		conn:   conn,
		client: client,
		appID:  appID,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Register(ctx context.Context, email, password string) (int64, error) {
	req := &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	}

	resp, err := c.client.Register(ctx, req)
	if err != nil {
		if codes.AlreadyExists == status.Code(err) {
			return 0, fmt.Errorf("sso register: %w", domain.ErrEmailAlreadyExists)
		}
		return 0, fmt.Errorf("sso register: %w", err)
	}

	return resp.UserId, nil
}

func (c *Client) Login(ctx context.Context, email, password string) (string, error) {
	req := &ssov1.LoginRequest{
		Email:    email,
		Password: password,
		AppId:    c.appID,
	}

	resp, err := c.client.Login(ctx, req)
	if err != nil {
		if status.Code(err) == codes.InvalidArgument {
			return "", fmt.Errorf("sso login: %w", domain.ErrInvalidCredentials)
		}

		return "", fmt.Errorf("sso login: %w", err)
	}

	return resp.Token, nil
}
