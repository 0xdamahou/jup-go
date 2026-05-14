package trigger

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/0xdamahou/jup-go/internal/juphttp"
)

const (
	testWallet = "11111111111111111111111111111111"
	testMintA  = "So11111111111111111111111111111111111111112"
	testMintB  = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
)

func TestCreateOCOPriceOrderValidatesLegs(t *testing.T) {
	c := NewClient(juphttp.NewClient(juphttp.Config{}))
	_, err := c.CreateOCOPriceOrder(context.Background(), OCOOrderRequest{
		TakeProfit: leg(Above),
		StopLoss:   leg(Above),
	})
	if err == nil {
		t.Fatal("expected duplicate condition error")
	}
}

func TestCreateOCOPriceOrderPostsTypedBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/trigger/v2/orders/price" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		var body map[string]CreateOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["takeProfit"].OrderType != OCO || body["stopLoss"].OrderType != OCO {
			t.Fatalf("body = %+v", body)
		}
		_, _ = w.Write([]byte(`{"id":"order"}`))
	}))
	defer srv.Close()
	c := NewClient(juphttp.NewClient(juphttp.Config{BaseURL: srv.URL, Timeout: time.Second}))
	out, err := c.CreateOCOPriceOrder(context.Background(), OCOOrderRequest{TakeProfit: leg(Above), StopLoss: leg(Below)})
	if err != nil {
		t.Fatal(err)
	}
	if out.ID != "order" {
		t.Fatalf("out = %+v", out)
	}
}

func leg(condition TriggerCondition) CreateOrderRequest {
	return CreateOrderRequest{
		WalletPubkey:     testWallet,
		InputMint:        testMintA,
		OutputMint:       testMintB,
		MakingAmount:     "100",
		TriggerPrice:     "1",
		TriggerCondition: condition,
	}
}
