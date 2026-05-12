package prediction

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/0xdamahou/jup-go/internal/juphttp"
)

type Client struct{ http *juphttp.Client }

func NewClient(h *juphttp.Client) *Client { return &Client{http: h} }

type Market struct {
	ID       string          `json:"id,omitempty"`
	Title    string          `json:"title,omitempty"`
	Outcomes json.RawMessage `json:"outcomes,omitempty"`
	Raw      json.RawMessage `json:"-"`
}
type Position struct {
	ID  string          `json:"id,omitempty"`
	Raw json.RawMessage `json:"-"`
}

func (c *Client) Markets(ctx context.Context) ([]Market, error) {
	var out []Market
	return out, c.http.GetJSON(ctx, c.http.Config().BaseURL, "/prediction/v1/events", nil, &out)
}
func (c *Client) Outcomes(ctx context.Context, marketID string) (json.RawMessage, error) {
	if marketID == "" {
		return nil, errors.New("market id is required")
	}
	var out json.RawMessage
	return out, c.http.GetJSON(ctx, c.http.Config().BaseURL, "/prediction/v1/events/"+marketID, nil, &out)
}
func (c *Client) CreatePosition(context.Context, any) error {
	return errors.New("prediction position mutation is not implemented without a confirmed public API contract")
}
