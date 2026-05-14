package recurring

import (
	"context"
	"testing"

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
