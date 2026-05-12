package swap

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/0xdamahou/jup-go/internal/juphttp"
)

const (
	testMintA = "So11111111111111111111111111111111111111112"
	testMintB = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
	testUser  = "11111111111111111111111111111111"
)

func TestGetOrderQueryEncoding(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/swap/v2/order" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("inputMint") != testMintA || q.Get("outputMint") != testMintB || q.Get("amount") != "1000" || q.Get("taker") != testUser {
			t.Fatalf("query = %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"inputMint":"` + testMintA + `","outputMint":"` + testMintB + `","inAmount":"1000","outAmount":"900"}`))
	}))
	defer srv.Close()
	c := NewClient(juphttp.NewClient(juphttp.Config{BaseURL: srv.URL, Timeout: time.Second}))
	out, err := c.GetOrder(context.Background(), GetOrderRequest{InputMint: testMintA, OutputMint: testMintB, Amount: "1000", Taker: testUser})
	if err != nil {
		t.Fatal(err)
	}
	if out.InAmount != "1000" || out.OutAmount != "900" {
		t.Fatalf("unexpected output: %+v", out)
	}
}
