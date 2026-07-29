package sso

import (
	"context"
	"errors"
	"fmt"

	ssov1 "github.com/barnigator/protos/gen/go/sso/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client ssov1.AuthClient
}

func New(address string) (*Client, error) {
	if address == "" {
		return nil, errors.New("sso address is empty")
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
		return 0, fmt.Errorf("sso register: %w", err)
	}

	return resp.UserId, nil
}

func (c *Client) Login(ctx context.Context, email, password string) (string, error) {
	req := &ssov1.LoginRequest{
		Email:    email,
		Password: password,
	}

	resp, err := c.client.Login(ctx, req)
	if err != nil {
		return "", fmt.Errorf("sso login: %w", err)
	}

	return resp.Token, nil
}
