package recurring

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/0xdamahou/jup-go/internal/juphttp"
)

type Client struct{ http *juphttp.Client }

func NewClient(h *juphttp.Client) *Client { return &Client{http: h} }

type CreateOrderRequest struct {
	User               string          `json:"user"`
	InputMint          string          `json:"inputMint"`
	OutputMint         string          `json:"outputMint"`
	Amount             string          `json:"amount"`
	NumberOfOrders     int             `json:"numberOfOrders"`
	IntervalSeconds    int64           `json:"intervalSeconds,omitempty"`
	StartAt            *int64          `json:"startAt,omitempty"`
	MinPrice           string          `json:"minPrice,omitempty"`
	MaxPrice           string          `json:"maxPrice,omitempty"`
	AdditionalJSONBody json.RawMessage `json:"-"`
}
type OrderResponse struct {
	ID          string          `json:"id,omitempty"`
	Transaction string          `json:"transaction,omitempty"`
	RequestID   string          `json:"requestId,omitempty"`
	Status      string          `json:"status,omitempty"`
	Raw         json.RawMessage `json:"-"`
}
type ExecuteRequest struct {
	SignedTransaction string `json:"signedTransaction"`
	RequestID         string `json:"requestId,omitempty"`
}

func (c *Client) CreateOrder(ctx context.Context, req CreateOrderRequest) (*OrderResponse, error) {
	if err := validate(req.User, req.InputMint, req.OutputMint, req.Amount); err != nil {
		return nil, err
	}
	if req.NumberOfOrders < 2 {
		return nil, errors.New("numberOfOrders must be at least 2")
	}
	if req.IntervalSeconds <= 0 {
		return nil, errors.New("intervalSeconds must be positive")
	}
	if req.StartAt != nil && *req.StartAt <= time.Now().Unix() {
		return nil, errors.New("startAt must be in the future")
	}
	if err := validatePriceRange(req.MinPrice, req.MaxPrice); err != nil {
		return nil, err
	}
	var out OrderResponse
	return &out, c.http.PostJSON(ctx, c.http.Config().BaseURL, "/recurring/v1/createOrder", req, &out, false)
}
func (c *Client) Execute(ctx context.Context, req ExecuteRequest) (*OrderResponse, error) {
	var out OrderResponse
	return &out, c.http.PostJSON(ctx, c.http.Config().BaseURL, "/recurring/v1/execute", req, &out, false)
}
func (c *Client) CancelOrder(ctx context.Context, orderID, user string) (*OrderResponse, error) {
	if orderID == "" || user == "" {
		return nil, errors.New("order id and user are required")
	}
	req := map[string]string{"order": orderID, "user": user}
	var out OrderResponse
	return &out, c.http.PostJSON(ctx, c.http.Config().BaseURL, "/recurring/v1/cancelOrder", req, &out, false)
}
func (c *Client) Orders(ctx context.Context, user string, page int) ([]OrderResponse, error) {
	if user == "" {
		return nil, errors.New("user is required")
	}
	q := url.Values{"user": []string{user}}
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	var out []OrderResponse
	return out, c.http.GetJSON(ctx, c.http.Config().BaseURL, "/recurring/v1/getRecurringOrders", q, &out)
}
func (c *Client) History(ctx context.Context, user string, page int) ([]OrderResponse, error) {
	if user == "" {
		return nil, errors.New("user is required")
	}
	if err := juphttp.ValidatePublicKey(user); err != nil {
		return nil, err
	}
	q := url.Values{"user": []string{user}}
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	var out []OrderResponse
	return out, c.http.GetJSON(ctx, c.http.Config().BaseURL, "/recurring/v1/getRecurringOrderHistory", q, &out)
}

func validate(user, input, output, amount string) error {
	for _, v := range []string{user, input, output} {
		if err := juphttp.ValidatePublicKey(v); err != nil {
			return err
		}
	}
	return juphttp.ValidateRawAmount(amount)
}

func validatePriceRange(minPrice, maxPrice string) error {
	var min, max float64
	var err error
	if minPrice != "" {
		min, err = strconv.ParseFloat(minPrice, 64)
		if err != nil || min <= 0 {
			return errors.New("minPrice must be a positive number")
		}
	}
	if maxPrice != "" {
		max, err = strconv.ParseFloat(maxPrice, 64)
		if err != nil || max <= 0 {
			return errors.New("maxPrice must be a positive number")
		}
	}
	if minPrice != "" && maxPrice != "" && min > max {
		return errors.New("minPrice must be less than or equal to maxPrice")
	}
	return nil
}
