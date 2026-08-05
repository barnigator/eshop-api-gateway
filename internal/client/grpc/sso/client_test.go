package sso

import (
	"context"
	"errors"
	"testing"

	"github.com/barnigator/eshop-api-gateway/internal/domain"
	ssov1 "github.com/barnigator/protos/gen/go/sso/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type authClientStub struct {
	ssov1.AuthClient

	registerFunc func(
		ctx context.Context,
		request *ssov1.RegisterRequest,
		options ...grpc.CallOption,
	) (*ssov1.RegisterResponse, error)

	loginFunc func(
		ctx context.Context,
		request *ssov1.LoginRequest,
		options ...grpc.CallOption,
	) (*ssov1.LoginResponse, error)
}

func (s authClientStub) Register(
	ctx context.Context,
	req *ssov1.RegisterRequest,
	opts ...grpc.CallOption,
) (*ssov1.RegisterResponse, error) {
	return s.registerFunc(ctx, req, opts...)
}

func (s authClientStub) Login(
	ctx context.Context,
	req *ssov1.LoginRequest,
	opts ...grpc.CallOption,
) (*ssov1.LoginResponse, error) {
	return s.loginFunc(ctx, req, opts...)
}

func TestClient_New(t *testing.T) {
	tests := []struct {
		name        string
		address     string
		appID       int32
		expectedErr error
	}{
		{
			name:    "success",
			address: "address",
			appID:   1,
		},
		{
			name:        "empty address",
			address:     "",
			expectedErr: ErrSsoAddressIsEmpty,
		},
		{
			name:        "appID = 0",
			address:     "address",
			appID:       0,
			expectedErr: ErrAppIDMustBePositive,
		},
		{
			name:        "appID = -1",
			address:     "address",
			appID:       -1,
			expectedErr: ErrAppIDMustBePositive,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cc, err := New(test.address, test.appID)
			if test.expectedErr != nil {
				if !errors.Is(err, test.expectedErr) {
					t.Fatalf("unexpected error: got %v, want %v", err, test.expectedErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: got %v", err)
			}

			t.Cleanup(func() {
				if err = cc.Close(); err != nil {
					t.Errorf("close client: %v", err)
				}
			})

			if cc.appID != test.appID {
				t.Fatalf("unexpected app id: got %v, want %v", cc.appID, test.appID)
			}
		})
	}
}

func TestClient_Register(t *testing.T) {
	tests := []struct {
		name                   string
		errCode                codes.Code
		wantEmailAlreadyExists bool
	}{
		{
			name:                   "email already exists",
			errCode:                codes.AlreadyExists,
			wantEmailAlreadyExists: true,
		},
		{
			name:                   "internal error",
			errCode:                codes.Internal,
			wantEmailAlreadyExists: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &Client{
				client: authClientStub{
					registerFunc: func(
						ctx context.Context,
						request *ssov1.RegisterRequest,
						options ...grpc.CallOption,
					) (*ssov1.RegisterResponse, error) {
						return nil, status.Error(test.errCode, "sso error")
					},
				},
			}

			_, err := client.Register(context.Background(), "user@example.com", "secret")

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			gotEmailAlreadyExists := errors.Is(err, domain.ErrEmailAlreadyExists)
			if gotEmailAlreadyExists != test.wantEmailAlreadyExists {
				t.Fatalf(
					"ErrEmailAlreadyExists: got %t, want %t; error: %v",
					gotEmailAlreadyExists,
					test.wantEmailAlreadyExists,
					err,
				)
			}
		})
	}
}

func TestClient_Login(t *testing.T) {
	tests := []struct {
		name                   string
		errCode                codes.Code
		wantInvalidCredentials bool
	}{
		{
			name:                   "invalid credentials",
			errCode:                codes.InvalidArgument,
			wantInvalidCredentials: true,
		},
		{
			name:                   "internal error",
			errCode:                codes.Internal,
			wantInvalidCredentials: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &Client{
				client: authClientStub{
					loginFunc: func(
						ctx context.Context,
						request *ssov1.LoginRequest,
						options ...grpc.CallOption,
					) (*ssov1.LoginResponse, error) {
						return nil, status.Error(test.errCode, "sso error")
					},
				},
			}

			_, err := client.Login(context.Background(), "user@example.com", "secret")

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			gotInvalidCredentials := errors.Is(err, domain.ErrInvalidCredentials)
			if gotInvalidCredentials != test.wantInvalidCredentials {
				t.Fatalf(
					"ErrInvalidCredentials: got %t, want %t; error: %v",
					gotInvalidCredentials,
					test.wantInvalidCredentials,
					err,
				)
			}
		})
	}
}
