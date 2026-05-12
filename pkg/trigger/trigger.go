package trigger

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"

	"github.com/0xdamahou/jup-go/internal/juphttp"
)

type Client struct{ http *juphttp.Client }

func NewClient(h *juphttp.Client) *Client { return &Client{http: h} }

type OrderType string
type TriggerCondition string

const (
	Single OrderType        = "single"
	OCO    OrderType        = "oco"
	OTOCO  OrderType        = "otoco"
	Above  TriggerCondition = "above"
	Below  TriggerCondition = "below"
)

type AuthChallengeRequest struct {
	WalletPubkey string `json:"walletPubkey"`
	Type         string `json:"type"`
}
type AuthChallengeResponse struct {
	Challenge string `json:"challenge"`
}
type AuthVerifyRequest struct {
	WalletPubkey string `json:"walletPubkey"`
	Type         string `json:"type"`
	Signature    string `json:"signature"`
}
type AuthVerifyResponse struct {
	Token string `json:"token"`
}

type CreateOrderRequest struct {
	WalletPubkey       string           `json:"walletPubkey"`
	InputMint          string           `json:"inputMint"`
	OutputMint         string           `json:"outputMint"`
	MakingAmount       string           `json:"makingAmount"`
	TakingAmount       string           `json:"takingAmount,omitempty"`
	TriggerPrice       string           `json:"triggerPrice,omitempty"`
	TriggerCondition   TriggerCondition `json:"triggerCondition,omitempty"`
	OrderType          OrderType        `json:"orderType,omitempty"`
	SlippageBps        *int             `json:"slippageBps,omitempty"`
	ExpiredAt          *int64           `json:"expiredAt,omitempty"`
	DepositRequestID   string           `json:"depositRequestId,omitempty"`
	DepositSignedTx    string           `json:"depositSignedTx,omitempty"`
	AdditionalJSONBody json.RawMessage  `json:"-"`
}
type OrderResponse struct {
	ID          string          `json:"id,omitempty"`
	Status      string          `json:"status,omitempty"`
	Transaction string          `json:"transaction,omitempty"`
	RequestID   string          `json:"requestId,omitempty"`
	Raw         json.RawMessage `json:"-"`
}

func (c *Client) AuthChallenge(ctx context.Context, req AuthChallengeRequest) (*AuthChallengeResponse, error) {
	var out AuthChallengeResponse
	return &out, c.http.PostJSON(ctx, c.http.Config().BaseURL, "/trigger/v2/auth/challenge", req, &out, true)
}
func (c *Client) AuthVerify(ctx context.Context, req AuthVerifyRequest) (*AuthVerifyResponse, error) {
	var out AuthVerifyResponse
	return &out, c.http.PostJSON(ctx, c.http.Config().BaseURL, "/trigger/v2/auth/verify", req, &out, false)
}
func (c *Client) CreatePriceOrder(ctx context.Context, req CreateOrderRequest) (*OrderResponse, error) {
	if err := validateOrder(req.WalletPubkey, req.InputMint, req.OutputMint, req.MakingAmount); err != nil {
		return nil, err
	}
	var out OrderResponse
	return &out, c.http.PostJSON(ctx, c.http.Config().BaseURL, "/trigger/v2/orders/price", req, &out, false)
}
func (c *Client) CancelPriceOrder(ctx context.Context, orderID string) (*OrderResponse, error) {
	if orderID == "" {
		return nil, errors.New("order id is required")
	}
	var out OrderResponse
	return &out, c.http.PostJSON(ctx, c.http.Config().BaseURL, "/trigger/v2/orders/price/cancel/"+url.PathEscape(orderID), map[string]string{}, &out, false)
}
func (c *Client) ConfirmCancelPriceOrder(ctx context.Context, orderID string, signedTransaction string, cancelRequestID string) (*OrderResponse, error) {
	if orderID == "" || signedTransaction == "" || cancelRequestID == "" {
		return nil, errors.New("order id, signed transaction, and cancel request id are required")
	}
	req := map[string]string{"signedTransaction": signedTransaction, "cancelRequestId": cancelRequestID}
	var out OrderResponse
	return &out, c.http.PostJSON(ctx, c.http.Config().BaseURL, "/trigger/v2/orders/price/confirm-cancel/"+url.PathEscape(orderID), req, &out, false)
}
func (c *Client) OpenOrders(ctx context.Context, wallet string) ([]OrderResponse, error) {
	q := url.Values{}
	if wallet != "" {
		q.Set("wallet", wallet)
	}
	var out []OrderResponse
	return out, c.http.GetJSON(ctx, c.http.Config().BaseURL, "/trigger/v2/orders", q, &out)
}
func (c *Client) History(ctx context.Context) ([]OrderResponse, error) {
	var out []OrderResponse
	return out, c.http.GetJSON(ctx, c.http.Config().BaseURL, "/trigger/v2/orders/history", nil, &out)
}

func validateOrder(wallet, input, output, amount string) error {
	for _, v := range []string{wallet, input, output} {
		if err := juphttp.ValidatePublicKey(v); err != nil {
			return err
		}
	}
	return juphttp.ValidateRawAmount(amount)
}
