package portfolio

import (
	"context"
	"testing"

	"github.com/0xdamahou/jup-go/internal/juphttp"
)

func TestPositionsValidatesWallet(t *testing.T) {
	c := NewClient(juphttp.NewClient(juphttp.Config{}))
	if _, err := c.Positions(context.Background(), "bad"); err == nil {
		t.Fatal("expected wallet validation error")
	}
}
