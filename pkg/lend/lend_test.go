package lend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/0xdamahou/jup-go/internal/juphttp"
)

const (
	testOwner = "11111111111111111111111111111111"
	testMint  = "So11111111111111111111111111111111111111112"
)

func TestBorrowValidatesRequest(t *testing.T) {
	c := NewClient(juphttp.NewClient(juphttp.Config{}))
	if _, err := c.Borrow(context.Background(), BorrowRequest{Owner: testOwner, Mint: testMint, Amount: "1.5"}); err == nil {
		t.Fatal("expected raw amount validation error")
	}
}

func TestEarnPositionsValidatesOwner(t *testing.T) {
	c := NewClient(juphttp.NewClient(juphttp.Config{}))
	if _, err := c.EarnPositions(context.Background(), "bad"); err == nil {
		t.Fatal("expected owner validation error")
	}
}

func TestRepayPostsTypedRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/lend/v1/repay" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"transaction":"tx"}`))
	}))
	defer srv.Close()
	c := NewClient(juphttp.NewClient(juphttp.Config{BaseURL: srv.URL, Timeout: time.Second}))
	out, err := c.Repay(context.Background(), RepayRequest{Owner: testOwner, Mint: testMint, Amount: "100"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Transaction != "tx" {
		t.Fatalf("out = %+v", out)
	}
}
