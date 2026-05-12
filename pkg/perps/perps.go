package perps

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"

	"github.com/0xdamahou/jup-go/internal/juphttp"
)

type Client struct{ http *juphttp.Client }

func NewClient(h *juphttp.Client) *Client { return &Client{http: h} }

type Market struct {
	ID  string          `json:"id,omitempty"`
	Raw json.RawMessage `json:"-"`
}
type Position struct {
	ID    string          `json:"id,omitempty"`
	Owner string          `json:"owner,omitempty"`
	Raw   json.RawMessage `json:"-"`
}

func (c *Client) Markets(ctx context.Context) ([]Market, error) {
	var out []Market
	return out, c.http.GetJSON(ctx, c.http.Config().BaseURL, "/perps/v1/markets", nil, &out)
}
func (c *Client) Positions(ctx context.Context, owner string) ([]Position, error) {
	if owner == "" {
		return nil, errors.New("owner is required")
	}
	if err := juphttp.ValidatePublicKey(owner); err != nil {
		return nil, err
	}
	q := url.Values{"owner": []string{owner}}
	var out []Position
	return out, c.http.GetJSON(ctx, c.http.Config().BaseURL, "/perps/v1/positions", q, &out)
}
func (c *Client) OpenPosition(context.Context, any) error {
	return errors.New("perps position mutation is not implemented: current Jupiter docs describe on-chain integration rather than a stable REST mutation endpoint")
}
func (c *Client) ClosePosition(context.Context, any) error {
	return errors.New("perps position mutation is not implemented: current Jupiter docs describe on-chain integration rather than a stable REST mutation endpoint")
}
