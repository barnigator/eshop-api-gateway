package sso

import (
	"errors"
	"testing"
)

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
