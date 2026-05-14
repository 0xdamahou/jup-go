package recurring

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
	testUser  = "11111111111111111111111111111111"
	testMintA = "So11111111111111111111111111111111111111112"
	testMintB = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
)

func TestCreateOrderValidatesSchedule(t *testing.T) {
	c := NewClient(juphttp.NewClient(juphttp.Config{}))
	_, err := c.CreateOrder(context.Background(), CreateOrderRequest{
		User:           testUser,
		InputMint:      testMintA,
		OutputMint:     testMintB,
		Amount:         "100",
		NumberOfOrders: 2,
	})
	if err == nil {
		t.Fatal("expected interval validation error")
	}
}

func TestCreateOrderValidatesPriceRange(t *testing.T) {
	c := NewClient(juphttp.NewClient(juphttp.Config{}))
	_, err := c.CreateOrder(context.Background(), CreateOrderRequest{
		User:            testUser,
		InputMint:       testMintA,
		OutputMint:      testMintB,
		Amount:          "100",
		NumberOfOrders:  2,
		IntervalSeconds: 60,
		MinPrice:        "10",
		MaxPrice:        "1",
	})
	if err == nil {
		t.Fatal("expected price range validation error")
	}
}

func TestCreateOrderUsesOfficialTimeParamsShape(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/recurring/v1/createOrder" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		var body struct {
			User       string `json:"user"`
			InputMint  string `json:"inputMint"`
			OutputMint string `json:"outputMint"`
			Params     struct {
				Time *TimeParams `json:"time"`
			} `json:"params"`
			Amount string `json:"amount"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Amount != "" || body.Params.Time == nil || body.Params.Time.InAmount != "100" || body.Params.Time.Interval != 60 {
			t.Fatalf("unexpected body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"transaction":"tx","requestId":"req"}`))
	}))
	defer srv.Close()
	startAt := time.Now().Add(time.Hour).Unix()
	c := NewClient(juphttp.NewClient(juphttp.Config{BaseURL: srv.URL, Timeout: time.Second}))
	out, err := c.CreateOrder(context.Background(), CreateOrderRequest{
		User:       testUser,
		InputMint:  testMintA,
		OutputMint: testMintB,
		Params: OrderParams{Time: &TimeParams{
			InAmount:       "100",
			NumberOfOrders: 2,
			Interval:       60,
			MinPrice:       "1",
			MaxPrice:       "2",
			StartAt:        &startAt,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Transaction != "tx" {
		t.Fatalf("out = %+v", out)
	}
}

func TestHistoryUsesGetRecurringOrdersFilters(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/recurring/v1/getRecurringOrders" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("user") != testUser || q.Get("page") != "2" || q.Get("orderStatus") != "history" || q.Get("recurringType") != "time" {
			t.Fatalf("query = %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()
	c := NewClient(juphttp.NewClient(juphttp.Config{BaseURL: srv.URL, Timeout: time.Second}))
	if _, err := c.History(context.Background(), testUser, 2); err != nil {
		t.Fatal(err)
	}
}
