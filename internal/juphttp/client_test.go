package juphttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetJSONRetries429ThenSucceeds(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("x-api-key") != "secret" {
			t.Fatalf("missing api key")
		}
		if calls == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"code":"RATE_LIMITED","message":"slow down"}`))
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()
	client := NewClient(Config{APIKey: "secret", BaseURL: srv.URL, Timeout: time.Second, MaxRetries: 1, RetryBackoff: time.Millisecond})
	var out struct {
		OK bool `json:"ok"`
	}
	if err := client.GetJSON(context.Background(), srv.URL, "/test", nil, &out); err != nil {
		t.Fatal(err)
	}
	if !out.OK || calls != 2 {
		t.Fatalf("out=%+v calls=%d", out, calls)
	}
}

func TestGetJSONStructuredError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-request-id", "req-1")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":"BAD","message":"bad request"}`))
	}))
	defer srv.Close()
	client := NewClient(Config{Timeout: time.Second})
	var out struct{}
	err := client.GetJSON(context.Background(), srv.URL, "/bad", nil, &out)
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T %v", err, err)
	}
	if apiErr.StatusCode != 400 || apiErr.Code != "BAD" || apiErr.RequestID != "req-1" || apiErr.Retryable {
		t.Fatalf("unexpected api error: %+v", apiErr)
	}
}

func TestGetJSONRetries5xx(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`{"message":"upstream"}`))
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()
	client := NewClient(Config{Timeout: time.Second, MaxRetries: 1, RetryBackoff: time.Millisecond})
	var out struct {
		OK bool `json:"ok"`
	}
	if err := client.GetJSON(context.Background(), srv.URL, "/test", nil, &out); err != nil {
		t.Fatal(err)
	}
	if calls != 2 || !out.OK {
		t.Fatalf("calls=%d out=%+v", calls, out)
	}
}

func TestGetJSONMalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{`))
	}))
	defer srv.Close()
	client := NewClient(Config{Timeout: time.Second})
	var out struct{}
	if err := client.GetJSON(context.Background(), srv.URL, "/bad-json", nil, &out); err == nil {
		t.Fatal("expected decode error")
	}
}
