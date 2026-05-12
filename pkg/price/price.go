package price

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"

	"github.com/0xdamahou/jup-go/internal/juphttp"
)

const MaxIDsPerRequest = 100

type Client struct{ http *juphttp.Client }

func NewClient(h *juphttp.Client) *Client { return &Client{http: h} }

// GetRequest requests USD prices for mint addresses.
type GetRequest struct {
	IDs []string
}

// Price contains one token price entry.
type Price struct {
	USDPrice    float64         `json:"usdPrice,omitempty"`
	BlockID     int64           `json:"blockId,omitempty"`
	Decimals    int             `json:"decimals,omitempty"`
	PriceChange json.RawMessage `json:"priceChange,omitempty"`
	Raw         json.RawMessage `json:"-"`
}

type Response map[string]Price

func (c *Client) Get(ctx context.Context, req GetRequest) (Response, error) {
	if len(req.IDs) == 0 {
		return nil, errors.New("at least one id is required")
	}
	if len(req.IDs) > MaxIDsPerRequest {
		return nil, errors.New("too many ids for one price request")
	}
	for _, id := range req.IDs {
		if err := juphttp.ValidatePublicKey(id); err != nil {
			return nil, err
		}
	}
	q := url.Values{"ids": []string{strings.Join(req.IDs, ",")}}
	out := Response{}
	return out, c.http.GetJSON(ctx, c.http.Config().LiteBaseURL, "/price/v3", q, &out)
}
