package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/barnigator/eshop-api-gateway/internal/domain"
)

type fakeUseCase struct {
	receivedEmail    string
	receivedPassword string
	receivedCtx      context.Context

	id     int64
	err    error
	called bool
}

func (f *fakeUseCase) Register(ctx context.Context, email, password string) (int64, error) {
	f.receivedEmail = email
	f.receivedPassword = password
	f.receivedCtx = ctx
	f.called = true

	return f.id, f.err
}

func (f *fakeUseCase) Login(ctx context.Context, email, password string) (string, error) {
	f.called = true
	return "", nil
}

type errorResponse struct {
	Error string `json:"error"`
}

func TestHandler_Register(t *testing.T) {
	tests := []struct {
		name               string
		body               string
		expectedEmail      string
		expectedPassword   string
		expectedID         int64
		expectedStatusCode int
		usecaseErr         error
		expectedErr        string
		expectedCall       bool
	}{
		{
			name:               "success",
			body:               `{"email":"user@example.com", "password":"secret"}`,
			expectedEmail:      "user@example.com",
			expectedPassword:   "secret",
			expectedID:         46,
			expectedStatusCode: http.StatusCreated,
			expectedCall:       true,
		},
		{
			name:               "invalid json",
			body:               `{"email":"user@example.com","password":`,
			expectedStatusCode: http.StatusBadRequest,
			expectedErr:        "invalid input data",
		},
		{
			name:               "internal error",
			body:               `{"email":"user@example.com", "password":"secret"}`,
			usecaseErr:         errors.New("internal error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedErr:        "internal error",
			expectedCall:       true,
		},
		{
			name:               "unknown field",
			body:               `{"login":"user@example.com", "password":"secret"}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedErr:        "invalid input data",
		},
		{
			name:               "double json",
			body:               `{"email":"a@example.com","password":"one"} {"email":"b@example.com","password":"two"}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedErr:        "invalid input data",
		},
		{
			name:               "email required",
			body:               `{"email":"", "password":"secret"}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedErr:        domain.ErrEmailRequired.Error(),
			usecaseErr:         domain.ErrEmailRequired,
			expectedCall:       true,
		},
		{
			name:               "password required",
			body:               `{"email":"user@example.com", "password":""}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedErr:        domain.ErrPasswordRequired.Error(),
			usecaseErr:         domain.ErrPasswordRequired,
			expectedCall:       true,
		},
		{
			name:               "generic error",
			body:               `{"email":"user@example.com", "password":"secret"}`,
			usecaseErr:         errors.New("sso database connection refused"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedErr:        "internal error",
			expectedCall:       true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			uc := &fakeUseCase{
				id:  test.expectedID,
				err: test.usecaseErr,
			}

			handler := New(uc)

			recorder := httptest.NewRecorder()

			request := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/auth/register",
				strings.NewReader(test.body),
			)

			type contextKey struct{}
			key := contextKey{}
			ctx := context.WithValue(request.Context(), key, "register-test")
			request = request.WithContext(ctx)

			handler.Register(recorder, request)

			response := recorder.Result()
			defer response.Body.Close()

			if response.StatusCode != test.expectedStatusCode {
				t.Fatalf("unexpected status code: got %v, want %v", response.StatusCode, test.expectedStatusCode)
			}

			if uc.called != test.expectedCall {
				t.Fatalf("unexpected register call state: got %v, want %v", uc.called, test.expectedCall)
			}

			if test.expectedCall {
				if got := uc.receivedCtx.Value(key); got != "register-test" {
					t.Fatalf("unexpected context value: got %v", got)
				}
			}

			if contentType := response.Header.Get("Content-Type"); contentType != "application/json" {
				t.Fatalf("unexpected content/type: got %s", contentType)
			}

			if test.expectedErr != "" {
				var responseBodyErr errorResponse
				if err := json.NewDecoder(response.Body).Decode(&responseBodyErr); err != nil {
					t.Fatalf("decode error response: %v", err)
				}

				if responseBodyErr.Error != test.expectedErr {
					t.Fatalf("unexpected error: %s, want %s", responseBodyErr.Error, test.expectedErr)
				}

				return
			}
			var responseBody registerResponse
			err := json.NewDecoder(response.Body).Decode(&responseBody)
			if err != nil {
				t.Fatalf("decode response: %v", err)
			}

			if responseBody.UserID != test.expectedID {
				t.Fatalf("user_id = %d, want %d", responseBody.UserID, test.expectedID)
			}

			if uc.receivedEmail != test.expectedEmail {
				t.Fatalf("unexpected email: got %v, want %v", uc.receivedEmail, test.expectedEmail)
			}

			if uc.receivedPassword != test.expectedPassword {
				t.Fatalf("unexpected password: got %v, want %v", uc.receivedPassword, test.expectedPassword)
			}

		})
	}
}
