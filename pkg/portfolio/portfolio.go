package portfolio

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/0xdamahou/jup-go/internal/juphttp"
)

type Client struct{ http *juphttp.Client }

func NewClient(h *juphttp.Client) *Client { return &Client{http: h} }

type Holding struct {
	Mint     string          `json:"mint,omitempty"`
	Raw      string          `json:"rawAmount,omitempty"`
	Amount   float64         `json:"amount,omitempty"`
	USDValue float64         `json:"usdValue,omitempty"`
	Decimals int             `json:"decimals,omitempty"`
	RawJSON  json.RawMessage `json:"-"`
}

func (c *Client) Holdings(ctx context.Context, wallet string) ([]Holding, error) {
	if wallet == "" {
		return nil, errors.New("wallet is required")
	}
	if err := juphttp.ValidatePublicKey(wallet); err != nil {
		return nil, err
	}
	var out []Holding
	return out, c.http.GetJSON(ctx, c.http.Config().BaseURL, "/portfolio/v1/holdings/"+wallet, nil, &out)
}
func (c *Client) Positions(ctx context.Context, wallet string) (json.RawMessage, error) {
	if wallet == "" {
		return nil, errors.New("wallet is required")
	}
	if err := juphttp.ValidatePublicKey(wallet); err != nil {
		return nil, err
	}
	var out json.RawMessage
	return out, c.http.GetJSON(ctx, c.http.Config().BaseURL, "/portfolio/v1/positions/"+wallet, nil, &out)
}
