package token

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"

	"github.com/0xdamahou/jup-go/internal/juphttp"
)

type Client struct{ http *juphttp.Client }

func NewClient(h *juphttp.Client) *Client { return &Client{http: h} }

// Token contains commonly used Tokens V2 metadata and keeps the full payload extensible.
type Token struct {
	ID           string          `json:"id,omitempty"`
	Mint         string          `json:"mint,omitempty"`
	Name         string          `json:"name,omitempty"`
	Symbol       string          `json:"symbol,omitempty"`
	Icon         string          `json:"icon,omitempty"`
	Decimals     int             `json:"decimals,omitempty"`
	Verified     bool            `json:"verified,omitempty"`
	OrganicScore float64         `json:"organicScore,omitempty"`
	Audit        json.RawMessage `json:"audit,omitempty"`
	Raw          json.RawMessage `json:"-"`
}

// SearchRequest searches by mint, symbol, or text. Comma-separated mint searches are supported.
type SearchRequest struct {
	Query string
}

func (c *Client) Search(ctx context.Context, req SearchRequest) ([]Token, error) {
	if strings.TrimSpace(req.Query) == "" {
		return nil, errors.New("query is required")
	}
	q := url.Values{"query": []string{req.Query}}
	var out []Token
	return out, c.http.GetJSON(ctx, c.http.Config().LiteBaseURL, "/tokens/v2/search", q, &out)
}

func (c *Client) ByTag(ctx context.Context, tag string) ([]Token, error) {
	if tag == "" {
		return nil, errors.New("tag is required")
	}
	q := url.Values{"query": []string{tag}}
	var out []Token
	return out, c.http.GetJSON(ctx, c.http.Config().LiteBaseURL, "/tokens/v2/tag", q, &out)
}

func (c *Client) Category(ctx context.Context, category, interval string) ([]Token, error) {
	if category == "" || interval == "" {
		return nil, errors.New("category and interval are required")
	}
	var out []Token
	return out, c.http.GetJSON(ctx, c.http.Config().LiteBaseURL, "/tokens/v2/"+category+"/"+interval, nil, &out)
}

func (c *Client) Recent(ctx context.Context) ([]Token, error) {
	var out []Token
	return out, c.http.GetJSON(ctx, c.http.Config().LiteBaseURL, "/tokens/v2/recent", nil, &out)
}
