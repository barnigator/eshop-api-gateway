package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeHandler struct {
	registerCall bool
	loginCall    bool
}

func (f *fakeHandler) Register(http.ResponseWriter, *http.Request) {
	f.registerCall = true
}

func (f *fakeHandler) Login(http.ResponseWriter, *http.Request) {
	f.loginCall = true
}

func TestNewRouter(t *testing.T) {
	tests := []struct {
		name               string
		target             string
		method             string
		expectedStatusCode int
		expectedRegCall    bool
		expectedLogCall    bool
	}{
		{
			name:               "success login",
			target:             "/api/v1/auth/login",
			method:             http.MethodPost,
			expectedLogCall:    true,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "success register",
			target:             "/api/v1/auth/register",
			method:             http.MethodPost,
			expectedRegCall:    true,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "fail get login",
			target:             "/api/v1/auth/login",
			method:             http.MethodGet,
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:               "fail get register",
			target:             "/api/v1/auth/register",
			method:             http.MethodGet,
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:               "unknown target",
			target:             "/unknown",
			method:             http.MethodPost,
			expectedStatusCode: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := &fakeHandler{}
			handler := NewRouter(h)

			recorder := httptest.NewRecorder()

			request := httptest.NewRequest(
				test.method,
				test.target,
				strings.NewReader(""),
			)

			handler.ServeHTTP(recorder, request)

			response := recorder.Result()
			defer response.Body.Close()

			if response.StatusCode != test.expectedStatusCode {
				t.Fatalf("unexpected status code: got %v, want %v", response.StatusCode, test.expectedStatusCode)
			}

			if h.registerCall != test.expectedRegCall {
				t.Fatalf("unexpected register call: got %t, want %t", h.registerCall, test.expectedRegCall)
			}

			if h.loginCall != test.expectedLogCall {
				t.Fatalf("unexpected login call: got %t, want %t", h.loginCall, test.expectedLogCall)
			}

		})
	}
}
