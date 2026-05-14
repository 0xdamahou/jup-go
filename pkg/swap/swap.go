package swap

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strconv"
	"time"

	"github.com/0xdamahou/jup-go/internal/juphttp"
)

const pathPrefix = "/swap/v2"

const computeBudgetProgramID = "ComputeBudget111111111111111111111111111111"

const (
	computeBudgetSetComputeUnitLimit byte = 2
	computeBudgetSetComputeUnitPrice byte = 3
)

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

// BuildRequest requests raw swap instructions for custom transaction assembly.
type BuildRequest struct {
	InputMint                string
	OutputMint               string
	Amount                   string
	Taker                    string
	Receiver                 string
	SwapMode                 SwapMode
	SlippageBPS              *int
	ReferralAccount          string
	ReferralFeeBPS           *int
	Payer                    string
	PriorityFeeLamports      *uint64
	PriorityFeeMicrolamports *uint64
	MaxFeeLamports           *uint64
	Router                   string
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

// BuildResponse contains raw instructions for self-managed execution.
type BuildResponse struct {
	ComputeBudgetInstructions []Instruction   `json:"computeBudgetInstructions,omitempty"`
	ComputeBudget             ComputeBudget   `json:"computeBudget,omitempty"`
	EstimatedMaxFeeLamports   *uint64         `json:"estimatedMaxFeeLamports,omitempty"`
	Raw                       json.RawMessage `json:"-"`
}

// Instruction is a raw Jupiter instruction. Data is normally base64 encoded.
type Instruction struct {
	ProgramID string               `json:"programId"`
	Accounts  []InstructionAccount `json:"accounts,omitempty"`
	Data      string               `json:"data"`
}

// InstructionAccount is an account meta in a raw Jupiter instruction.
type InstructionAccount struct {
	Pubkey     string `json:"pubkey"`
	IsSigner   bool   `json:"isSigner"`
	IsWritable bool   `json:"isWritable"`
}

// ComputeBudget summarizes ComputeBudget instructions returned by /build.
type ComputeBudget struct {
	UnitLimit              uint32 `json:"unitLimit,omitempty"`
	UnitPriceMicrolamports uint64 `json:"unitPriceMicrolamports,omitempty"`
	HasUnitLimit           bool   `json:"hasUnitLimit"`
	HasUnitPrice           bool   `json:"hasUnitPrice"`
}

func (r *BuildResponse) UnmarshalJSON(data []byte) error {
	type alias BuildResponse
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*r = BuildResponse(a)
	r.Raw = append(r.Raw[:0], data...)
	budget, err := ParseComputeBudgetInstructions(r.ComputeBudgetInstructions)
	if err != nil {
		return err
	}
	r.ComputeBudget = budget
	if budget.HasUnitLimit && budget.HasUnitPrice {
		fee := EstimatePriorityFeeLamports(budget.UnitLimit, budget.UnitPriceMicrolamports)
		r.EstimatedMaxFeeLamports = &fee
	}
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
	var out BuildResponse
	if err := c.http.GetJSON(ctx, c.http.Config().BaseURL, pathPrefix+"/build", q, &out); err != nil {
		return nil, err
	}
	if req.PriorityFeeMicrolamports != nil {
		if err := out.OverrideComputeUnitPrice(*req.PriorityFeeMicrolamports); err != nil {
			return nil, err
		}
	}
	if req.MaxFeeLamports != nil {
		if err := out.GuardMaxFee(*req.MaxFeeLamports); err != nil {
			return nil, err
		}
	}
	return &out, nil
}

// GuardMaxFee rejects a build response whose estimated priority fee exceeds maxFeeLamports.
func (r *BuildResponse) GuardMaxFee(maxFeeLamports uint64) error {
	if r.ComputeBudget.HasUnitPrice && !r.ComputeBudget.HasUnitLimit {
		return errors.New("cannot guard priority fee: compute unit price is set but compute unit limit is missing")
	}
	if r.EstimatedMaxFeeLamports == nil {
		return nil
	}
	if *r.EstimatedMaxFeeLamports > maxFeeLamports {
		return fmt.Errorf("priority fee estimate %d lamports exceeds max %d lamports", *r.EstimatedMaxFeeLamports, maxFeeLamports)
	}
	return nil
}

// OverrideComputeUnitPrice replaces SetComputeUnitPrice in computeBudgetInstructions.
func (r *BuildResponse) OverrideComputeUnitPrice(microlamports uint64) error {
	if len(r.ComputeBudgetInstructions) == 0 {
		return errors.New("cannot override compute unit price: no computeBudgetInstructions returned")
	}
	replaced := false
	for i := range r.ComputeBudgetInstructions {
		data, enc, err := decodeInstructionData(r.ComputeBudgetInstructions[i].Data)
		if err != nil || len(data) == 0 || data[0] != computeBudgetSetComputeUnitPrice {
			continue
		}
		r.ComputeBudgetInstructions[i].Data = encodeInstructionData(encodeSetComputeUnitPrice(microlamports), enc)
		replaced = true
	}
	if !replaced {
		return errors.New("cannot override compute unit price: SetComputeUnitPrice instruction missing")
	}
	budget, err := ParseComputeBudgetInstructions(r.ComputeBudgetInstructions)
	if err != nil {
		return err
	}
	r.ComputeBudget = budget
	if budget.HasUnitLimit && budget.HasUnitPrice {
		fee := EstimatePriorityFeeLamports(budget.UnitLimit, budget.UnitPriceMicrolamports)
		r.EstimatedMaxFeeLamports = &fee
	} else {
		r.EstimatedMaxFeeLamports = nil
	}
	return r.refreshRaw()
}

func (r *BuildResponse) refreshRaw() error {
	var obj map[string]json.RawMessage
	if len(r.Raw) > 0 {
		if err := json.Unmarshal(r.Raw, &obj); err != nil {
			return err
		}
	}
	if obj == nil {
		obj = map[string]json.RawMessage{}
	}
	rawInstructions, err := json.Marshal(r.ComputeBudgetInstructions)
	if err != nil {
		return err
	}
	obj["computeBudgetInstructions"] = rawInstructions
	rawBudget, err := json.Marshal(r.ComputeBudget)
	if err != nil {
		return err
	}
	obj["computeBudget"] = rawBudget
	if r.EstimatedMaxFeeLamports != nil {
		rawFee, err := json.Marshal(*r.EstimatedMaxFeeLamports)
		if err != nil {
			return err
		}
		obj["estimatedMaxFeeLamports"] = rawFee
	} else {
		delete(obj, "estimatedMaxFeeLamports")
	}
	raw, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	r.Raw = raw
	return nil
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

// ParseComputeBudgetInstructions extracts CU limit and CU price from ComputeBudget instructions.
func ParseComputeBudgetInstructions(instructions []Instruction) (ComputeBudget, error) {
	var budget ComputeBudget
	for _, ix := range instructions {
		if ix.ProgramID != computeBudgetProgramID {
			continue
		}
		data, _, err := decodeInstructionData(ix.Data)
		if err != nil {
			return budget, err
		}
		if len(data) == 0 {
			continue
		}
		switch data[0] {
		case computeBudgetSetComputeUnitLimit:
			if len(data) < 5 {
				return budget, errors.New("invalid SetComputeUnitLimit instruction data")
			}
			budget.UnitLimit = binary.LittleEndian.Uint32(data[1:5])
			budget.HasUnitLimit = true
		case computeBudgetSetComputeUnitPrice:
			if len(data) < 9 {
				return budget, errors.New("invalid SetComputeUnitPrice instruction data")
			}
			budget.UnitPriceMicrolamports = binary.LittleEndian.Uint64(data[1:9])
			budget.HasUnitPrice = true
		}
	}
	return budget, nil
}

// EstimatePriorityFeeLamports returns ceil(unitLimit * microlamports / 1_000_000).
func EstimatePriorityFeeLamports(unitLimit uint32, microlamports uint64) uint64 {
	product := new(big.Int).Mul(new(big.Int).SetUint64(uint64(unitLimit)), new(big.Int).SetUint64(microlamports))
	divisor := big.NewInt(1_000_000)
	quotient, remainder := new(big.Int).QuoRem(product, divisor, new(big.Int))
	if remainder.Sign() > 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	return quotient.Uint64()
}

type instructionEncoding int

const (
	instructionEncodingBase64 instructionEncoding = iota
	instructionEncodingBase58
)

func decodeInstructionData(raw string) ([]byte, instructionEncoding, error) {
	data, err := base64.StdEncoding.DecodeString(raw)
	if err == nil {
		return data, instructionEncodingBase64, nil
	}
	data, err = base64.RawStdEncoding.DecodeString(raw)
	if err == nil {
		return data, instructionEncodingBase64, nil
	}
	data, err = decodeBase58(raw)
	if err == nil {
		return data, instructionEncodingBase58, nil
	}
	return nil, instructionEncodingBase64, fmt.Errorf("decode instruction data: %w", err)
}

func encodeInstructionData(data []byte, enc instructionEncoding) string {
	if enc == instructionEncodingBase58 {
		return encodeBase58(data)
	}
	return base64.StdEncoding.EncodeToString(data)
}

func encodeSetComputeUnitPrice(microlamports uint64) []byte {
	data := make([]byte, 9)
	data[0] = computeBudgetSetComputeUnitPrice
	binary.LittleEndian.PutUint64(data[1:], microlamports)
	return data
}

func decodeBase58(s string) ([]byte, error) {
	const alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	result := big.NewInt(0)
	base := big.NewInt(58)
	for _, r := range s {
		idx := int64(bytes.IndexRune([]byte(alphabet), r))
		if idx < 0 {
			return nil, fmt.Errorf("invalid base58 character %q", r)
		}
		result.Mul(result, base)
		result.Add(result, big.NewInt(idx))
	}
	decoded := result.Bytes()
	leadingZeros := 0
	for leadingZeros < len(s) && s[leadingZeros] == '1' {
		leadingZeros++
	}
	return append(bytes.Repeat([]byte{0}, leadingZeros), decoded...), nil
}

func encodeBase58(data []byte) string {
	const alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	x := new(big.Int).SetBytes(data)
	base := big.NewInt(58)
	zero := big.NewInt(0)
	mod := new(big.Int)
	var out []byte
	for x.Cmp(zero) > 0 {
		x.DivMod(x, base, mod)
		out = append(out, alphabet[mod.Int64()])
	}
	for _, b := range data {
		if b != 0 {
			break
		}
		out = append(out, alphabet[0])
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return string(out)
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
