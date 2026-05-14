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
	User            string      `json:"user"`
	InputMint       string      `json:"inputMint"`
	OutputMint      string      `json:"outputMint"`
	Params          OrderParams `json:"params,omitempty"`
	Amount          string      `json:"-"`
	NumberOfOrders  int         `json:"-"`
	IntervalSeconds int64       `json:"-"`
	StartAt         *int64      `json:"-"`
	MinPrice        string      `json:"-"`
	MaxPrice        string      `json:"-"`
}
type OrderParams struct {
	Time *TimeParams `json:"time,omitempty"`
}
type TimeParams struct {
	InAmount       string `json:"inAmount"`
	NumberOfOrders int    `json:"numberOfOrders"`
	Interval       int64  `json:"interval"`
	MinPrice       any    `json:"minPrice,omitempty"`
	MaxPrice       any    `json:"maxPrice,omitempty"`
	StartAt        *int64 `json:"startAt,omitempty"`
}
type GetOrdersRequest struct {
	User            string
	Page            int
	OrderStatus     OrderStatus
	RecurringType   RecurringType
	Mint            string
	IncludeFailedTx *bool
}
type OrderStatus string
type RecurringType string

const (
	OrderStatusActive  OrderStatus   = "active"
	OrderStatusHistory OrderStatus   = "history"
	RecurringTypeTime  RecurringType = "time"
	RecurringTypeAll   RecurringType = "all"
)

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
	req = req.normalized()
	if req.Params.Time == nil {
		return nil, errors.New("params.time is required")
	}
	if err := validate(req.User, req.InputMint, req.OutputMint, req.Params.Time.InAmount); err != nil {
		return nil, err
	}
	if req.Params.Time.NumberOfOrders < 2 {
		return nil, errors.New("numberOfOrders must be at least 2")
	}
	if req.Params.Time.Interval <= 0 {
		return nil, errors.New("params.time.interval must be positive")
	}
	if req.Params.Time.StartAt != nil && *req.Params.Time.StartAt <= time.Now().Unix() {
		return nil, errors.New("startAt must be in the future")
	}
	if err := validatePriceRange(req.Params.Time.MinPrice, req.Params.Time.MaxPrice); err != nil {
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
	return c.GetOrders(ctx, GetOrdersRequest{User: user, Page: page, OrderStatus: OrderStatusActive, RecurringType: RecurringTypeTime})
}
func (c *Client) History(ctx context.Context, user string, page int) ([]OrderResponse, error) {
	return c.GetOrders(ctx, GetOrdersRequest{User: user, Page: page, OrderStatus: OrderStatusHistory, RecurringType: RecurringTypeTime})
}
func (c *Client) GetOrders(ctx context.Context, req GetOrdersRequest) ([]OrderResponse, error) {
	if req.User == "" {
		return nil, errors.New("user is required")
	}
	if err := juphttp.ValidatePublicKey(req.User); err != nil {
		return nil, err
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.OrderStatus == "" {
		req.OrderStatus = OrderStatusActive
	}
	if req.OrderStatus != OrderStatusActive && req.OrderStatus != OrderStatusHistory {
		return nil, errors.New("orderStatus must be active or history")
	}
	if req.RecurringType == "" {
		req.RecurringType = RecurringTypeTime
	}
	if req.RecurringType != RecurringTypeTime && req.RecurringType != RecurringTypeAll {
		return nil, errors.New("recurringType must be time or all")
	}
	q := url.Values{
		"user":          []string{req.User},
		"page":          []string{strconv.Itoa(req.Page)},
		"orderStatus":   []string{string(req.OrderStatus)},
		"recurringType": []string{string(req.RecurringType)},
	}
	if req.Mint != "" {
		if err := juphttp.ValidatePublicKey(req.Mint); err != nil {
			return nil, err
		}
		q.Set("mint", req.Mint)
	}
	if req.IncludeFailedTx != nil {
		q.Set("includeFailedTx", strconv.FormatBool(*req.IncludeFailedTx))
	}
	var out []OrderResponse
	return out, c.http.GetJSON(ctx, c.http.Config().BaseURL, "/recurring/v1/getRecurringOrders", q, &out)
}

func validate(user, input, output, amount string) error {
	for _, v := range []string{user, input, output} {
		if err := juphttp.ValidatePublicKey(v); err != nil {
			return err
		}
	}
	return juphttp.ValidateRawAmount(amount)
}

func (r CreateOrderRequest) normalized() CreateOrderRequest {
	if r.Params.Time == nil && (r.Amount != "" || r.NumberOfOrders != 0 || r.IntervalSeconds != 0 || r.StartAt != nil || r.MinPrice != "" || r.MaxPrice != "") {
		r.Params.Time = &TimeParams{
			InAmount:       r.Amount,
			NumberOfOrders: r.NumberOfOrders,
			Interval:       r.IntervalSeconds,
			StartAt:        r.StartAt,
		}
		if r.MinPrice != "" {
			r.Params.Time.MinPrice = r.MinPrice
		}
		if r.MaxPrice != "" {
			r.Params.Time.MaxPrice = r.MaxPrice
		}
	}
	return r
}

func validatePriceRange(minPrice, maxPrice any) error {
	var min, max float64
	if minPrice != nil {
		var err error
		min, err = parsePositivePrice(minPrice)
		if err != nil {
			return errors.New("minPrice must be a positive number")
		}
	}
	if maxPrice != nil {
		var err error
		max, err = parsePositivePrice(maxPrice)
		if err != nil {
			return errors.New("maxPrice must be a positive number")
		}
	}
	if minPrice != nil && maxPrice != nil && min > max {
		return errors.New("minPrice must be less than or equal to maxPrice")
	}
	return nil
}

func parsePositivePrice(v any) (float64, error) {
	switch x := v.(type) {
	case string:
		if x == "" {
			return 0, nil
		}
		f, err := strconv.ParseFloat(x, 64)
		if err != nil || f <= 0 {
			return 0, errors.New("invalid price")
		}
		return f, nil
	case float64:
		if x <= 0 {
			return 0, errors.New("invalid price")
		}
		return x, nil
	case int:
		if x <= 0 {
			return 0, errors.New("invalid price")
		}
		return float64(x), nil
	default:
		return 0, errors.New("invalid price")
	}
}
