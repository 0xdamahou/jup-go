package lend

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
	ID     string          `json:"id,omitempty"`
	Wallet string          `json:"wallet,omitempty"`
	Raw    json.RawMessage `json:"-"`
}
type TransactionResponse struct {
	Transaction string          `json:"transaction,omitempty"`
	RequestID   string          `json:"requestId,omitempty"`
	Raw         json.RawMessage `json:"-"`
}
type EarnActionRequest struct {
	Owner  string `json:"owner"`
	Mint   string `json:"mint"`
	Amount string `json:"amount"`
}

func (c *Client) EarnTokens(ctx context.Context) ([]Market, error) {
	var out []Market
	return out, c.http.GetJSON(ctx, c.http.Config().BaseURL, "/lend/v1/earn/tokens", nil, &out)
}
func (c *Client) EarnPositions(ctx context.Context, owner string) ([]Position, error) {
	if owner == "" {
		return nil, errors.New("owner is required")
	}
	q := url.Values{"owner": []string{owner}}
	var out []Position
	return out, c.http.GetJSON(ctx, c.http.Config().BaseURL, "/lend/v1/earn/positions", q, &out)
}
func (c *Client) Deposit(ctx context.Context, req EarnActionRequest) (*TransactionResponse, error) {
	if err := validate(req); err != nil {
		return nil, err
	}
	var out TransactionResponse
	return &out, c.http.PostJSON(ctx, c.http.Config().BaseURL, "/lend/v1/earn/deposit", req, &out, false)
}
func (c *Client) Withdraw(ctx context.Context, req EarnActionRequest) (*TransactionResponse, error) {
	if err := validate(req); err != nil {
		return nil, err
	}
	var out TransactionResponse
	return &out, c.http.PostJSON(ctx, c.http.Config().BaseURL, "/lend/v1/earn/withdraw", req, &out, false)
}
func (c *Client) Borrow(ctx context.Context, req any) (*TransactionResponse, error) {
	var out TransactionResponse
	return &out, c.http.PostJSON(ctx, c.http.Config().BaseURL, "/lend/v1/borrow", req, &out, false)
}
func (c *Client) Repay(ctx context.Context, req any) (*TransactionResponse, error) {
	var out TransactionResponse
	return &out, c.http.PostJSON(ctx, c.http.Config().BaseURL, "/lend/v1/repay", req, &out, false)
}

func validate(req EarnActionRequest) error {
	if err := juphttp.ValidatePublicKey(req.Owner); err != nil {
		return err
	}
	if err := juphttp.ValidatePublicKey(req.Mint); err != nil {
		return err
	}
	return juphttp.ValidateRawAmount(req.Amount)
}
