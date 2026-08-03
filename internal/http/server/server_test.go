package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

func TestNew(t *testing.T) {
	handler := http.NewServeMux()
	timeout := 5 * time.Second

	server := New(handler, 8080, timeout)

	if server.httpServer == nil {
		t.Fatal("http server is nil")
	}

	if server.httpServer.Addr != ":8080" {
		t.Fatalf("unexpected port: got %s, want \":8080\"", server.httpServer.Addr)
	}

	if server.httpServer.Handler != handler {
		t.Fatal("handler was not preserved")
	}

	if server.httpServer.ReadHeaderTimeout != timeout {
		t.Fatalf("wrong ReadHeaderTimeout: got %s, want %s", server.httpServer.ReadHeaderTimeout, timeout)
	}
}

func TestServer_Shutdown(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	s := New(handler, 0, time.Second)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- s.httpServer.Serve(listener)
	}()

	client := &http.Client{Timeout: time.Second}
	resp, err := client.Get("http://" + listener.Addr().String())
	if err != nil {
		t.Fatalf("request to running server: %v", err)
	}
	resp.Body.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err = s.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	if err = <-serveErr; !errors.Is(err, http.ErrServerClosed) {
		t.Fatalf("Serve error: got %v, want %v", err, http.ErrServerClosed)
	}
}
