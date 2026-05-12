package swap

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"time"

	"github.com/0xdamahou/jup-go/internal/juphttp"
)

const pathPrefix = "/swap/v2"

// SwapMode identifies exact-in or exact-out swap semantics.
type SwapMode string

const (
	ExactIn  SwapMode = "ExactIn"
	ExactOut SwapMode = "ExactOut"
)

// Client provides Jupiter Swap API V2 methods.
type Client struct{ http *juphttp.Client }

func NewClient(h *juphttp.Client) *Client { return &Client{http: h} }

// GetOrderRequest requests a quote and, when Taker is set, an assembled transaction.
type GetOrderRequest struct {
	InputMint           string
	OutputMint          string
	Amount              string
	Taker               string
	Receiver            string
	SwapMode            SwapMode
	SlippageBPS         *int
	ReferralAccount     string
	ReferralFeeBPS      *int
	Payer               string
	PriorityFeeLamports *uint64
	Router              string
}

// OrderResponse is intentionally extensible because Jupiter may add route fields.
type OrderResponse struct {
	RequestID   string          `json:"requestId,omitempty"`
	Mode        string          `json:"mode,omitempty"`
	Router      string          `json:"router,omitempty"`
	InputMint   string          `json:"inputMint,omitempty"`
	OutputMint  string          `json:"outputMint,omitempty"`
	InAmount    string          `json:"inAmount,omitempty"`
	OutAmount   string          `json:"outAmount,omitempty"`
	OtherAmount string          `json:"otherAmountThreshold,omitempty"`
	Tx          string          `json:"tx,omitempty"`
	ExpireAt    *time.Time      `json:"expireAt,omitempty"`
	Raw         json.RawMessage `json:"-"`
}

func (r *OrderResponse) UnmarshalJSON(data []byte) error {
	type alias OrderResponse
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*r = OrderResponse(a)
	r.Raw = append(r.Raw[:0], data...)
	return nil
}

// GetOrder calls GET /swap/v2/order.
func (c *Client) GetOrder(ctx context.Context, req GetOrderRequest) (*OrderResponse, error) {
	if err := validateOrder(req.InputMint, req.OutputMint, req.Amount); err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("inputMint", req.InputMint)
	q.Set("outputMint", req.OutputMint)
	q.Set("amount", req.Amount)
	add(q, "taker", req.Taker)
	add(q, "receiver", req.Receiver)
	if req.SwapMode != "" {
		q.Set("swapMode", string(req.SwapMode))
	}
	addInt(q, "slippageBps", req.SlippageBPS)
	add(q, "referralAccount", first(req.ReferralAccount, c.http.Config().ReferralAccount))
	addInt(q, "referralFee", firstInt(req.ReferralFeeBPS, c.http.Config().ReferralFeeBPS))
	add(q, "payer", first(req.Payer, c.http.Config().Payer))
	addUint(q, "priorityFeeLamports", req.PriorityFeeLamports)
	add(q, "router", req.Router)
	var out OrderResponse
	if err := c.http.GetJSON(ctx, c.http.Config().BaseURL, pathPrefix+"/order", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// BuildRequest requests raw swap instructions for custom transaction assembly.
type BuildRequest = GetOrderRequest

// BuildResponse contains raw instructions for self-managed execution.
type BuildResponse struct {
	Raw json.RawMessage `json:"-"`
}

func (r *BuildResponse) UnmarshalJSON(data []byte) error {
	r.Raw = append(r.Raw[:0], data...)
	return nil
}

// GetBuild calls GET /swap/v2/build.
func (c *Client) GetBuild(ctx context.Context, req BuildRequest) (*BuildResponse, error) {
	if err := validateOrder(req.InputMint, req.OutputMint, req.Amount); err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("inputMint", req.InputMint)
	q.Set("outputMint", req.OutputMint)
	q.Set("amount", req.Amount)
	add(q, "taker", req.Taker)
	if req.SwapMode != "" {
		q.Set("swapMode", string(req.SwapMode))
	}
	addInt(q, "slippageBps", req.SlippageBPS)
	var out BuildResponse
	return &out, c.http.GetJSON(ctx, c.http.Config().BaseURL, pathPrefix+"/build", q, &out)
}

// ExecuteRequest submits a locally signed /order transaction to Jupiter managed execution.
type ExecuteRequest struct {
	SignedTransaction string `json:"signedTransaction"`
	RequestID         string `json:"requestId,omitempty"`
}

// ExecuteResponse is the managed execution result.
type ExecuteResponse struct {
	Status    string `json:"status,omitempty"`
	Signature string `json:"signature,omitempty"`
	Slot      uint64 `json:"slot,omitempty"`
	Code      int    `json:"code,omitempty"`
	Error     string `json:"error,omitempty"`
	Raw       json.RawMessage
}

func (r *ExecuteResponse) UnmarshalJSON(data []byte) error {
	type alias ExecuteResponse
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*r = ExecuteResponse(a)
	r.Raw = append(r.Raw[:0], data...)
	return nil
}

// Execute calls POST /swap/v2/execute. It is not retried automatically.
func (c *Client) Execute(ctx context.Context, req ExecuteRequest) (*ExecuteResponse, error) {
	var out ExecuteResponse
	return &out, c.http.PostJSON(ctx, c.http.Config().BaseURL, pathPrefix+"/execute", req, &out, false)
}

func validateOrder(inputMint, outputMint, amount string) error {
	if err := juphttp.ValidatePublicKey(inputMint); err != nil {
		return err
	}
	if err := juphttp.ValidatePublicKey(outputMint); err != nil {
		return err
	}
	return juphttp.ValidateRawAmount(amount)
}

func add(q url.Values, k, v string) {
	if v != "" {
		q.Set(k, v)
	}
}
func addInt(q url.Values, k string, v *int) {
	if v != nil {
		q.Set(k, strconv.Itoa(*v))
	}
}
func addUint(q url.Values, k string, v *uint64) {
	if v != nil {
		q.Set(k, strconv.FormatUint(*v, 10))
	}
}
func first(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
func firstInt(v *int, fallback int) *int {
	if v != nil {
		return v
	}
	if fallback == 0 {
		return nil
	}
	return &fallback
}
