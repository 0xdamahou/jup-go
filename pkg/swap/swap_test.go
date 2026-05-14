package swap

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/0xdamahou/jup-go/internal/juphttp"
)

const (
	testMintA = "So11111111111111111111111111111111111111112"
	testMintB = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
	testUser  = "11111111111111111111111111111111"
)

func TestGetOrderQueryEncoding(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/swap/v2/order" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("inputMint") != testMintA || q.Get("outputMint") != testMintB || q.Get("amount") != "1000" || q.Get("taker") != testUser {
			t.Fatalf("query = %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"inputMint":"` + testMintA + `","outputMint":"` + testMintB + `","inAmount":"1000","outAmount":"900"}`))
	}))
	defer srv.Close()
	c := NewClient(juphttp.NewClient(juphttp.Config{BaseURL: srv.URL, Timeout: time.Second}))
	out, err := c.GetOrder(context.Background(), GetOrderRequest{InputMint: testMintA, OutputMint: testMintB, Amount: "1000", Taker: testUser})
	if err != nil {
		t.Fatal(err)
	}
	if out.InAmount != "1000" || out.OutAmount != "900" {
		t.Fatalf("unexpected output: %+v", out)
	}
}

func TestGetBuildQueryEncodingAndComputeBudgetGuard(t *testing.T) {
	priorityFee := uint64(5000)
	maxFee := uint64(1000)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/swap/v2/build" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("priorityFeeLamports") != "5000" || q.Get("router") != "iris" || q.Get("payer") != testUser {
			t.Fatalf("query = %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{
			"computeBudgetInstructions": [
				{"programId":"ComputeBudget111111111111111111111111111111","data":"` + computeUnitLimitData(1_400_000) + `"},
				{"programId":"ComputeBudget111111111111111111111111111111","data":"` + base64.StdEncoding.EncodeToString(encodeSetComputeUnitPrice(2_000_000)) + `"}
			],
			"swapInstruction": {"programId":"swap","data":"abc"}
		}`))
	}))
	defer srv.Close()
	c := NewClient(juphttp.NewClient(juphttp.Config{BaseURL: srv.URL, Timeout: time.Second}))
	_, err := c.GetBuild(context.Background(), BuildRequest{
		InputMint:           testMintA,
		OutputMint:          testMintB,
		Amount:              "1000",
		Taker:               testUser,
		Payer:               testUser,
		PriorityFeeLamports: &priorityFee,
		Router:              "iris",
		MaxFeeLamports:      &maxFee,
	})
	if err == nil {
		t.Fatal("expected max fee guard error")
	}
}

func TestGetBuildOverridesComputeUnitPrice(t *testing.T) {
	override := uint64(10)
	maxFee := uint64(20)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"computeBudgetInstructions": [
				{"programId":"ComputeBudget111111111111111111111111111111","data":"` + computeUnitLimitData(1_000_000) + `"},
				{"programId":"ComputeBudget111111111111111111111111111111","data":"` + base64.StdEncoding.EncodeToString(encodeSetComputeUnitPrice(9_000_000)) + `"}
			],
			"swapInstruction": {"programId":"swap","data":"abc"}
		}`))
	}))
	defer srv.Close()
	c := NewClient(juphttp.NewClient(juphttp.Config{BaseURL: srv.URL, Timeout: time.Second}))
	out, err := c.GetBuild(context.Background(), BuildRequest{
		InputMint:                testMintA,
		OutputMint:               testMintB,
		Amount:                   "1000",
		Taker:                    testUser,
		PriorityFeeMicrolamports: &override,
		MaxFeeLamports:           &maxFee,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.ComputeBudget.UnitPriceMicrolamports != override {
		t.Fatalf("unit price = %d", out.ComputeBudget.UnitPriceMicrolamports)
	}
	if out.EstimatedMaxFeeLamports == nil || *out.EstimatedMaxFeeLamports != 10 {
		t.Fatalf("fee = %v", out.EstimatedMaxFeeLamports)
	}
	var raw struct {
		ComputeBudgetInstructions []Instruction   `json:"computeBudgetInstructions"`
		SwapInstruction           json.RawMessage `json:"swapInstruction"`
	}
	if err := json.Unmarshal(out.Raw, &raw); err != nil {
		t.Fatal(err)
	}
	if len(raw.ComputeBudgetInstructions) != 2 || len(raw.SwapInstruction) == 0 {
		t.Fatalf("raw json was not preserved: %s", string(out.Raw))
	}
	budget, err := ParseComputeBudgetInstructions(raw.ComputeBudgetInstructions)
	if err != nil {
		t.Fatal(err)
	}
	if budget.UnitPriceMicrolamports != override {
		t.Fatalf("raw unit price = %d", budget.UnitPriceMicrolamports)
	}
}

func TestParseComputeBudgetInstructions(t *testing.T) {
	budget, err := ParseComputeBudgetInstructions([]Instruction{
		{ProgramID: computeBudgetProgramID, Data: computeUnitLimitData(500_000)},
		{ProgramID: computeBudgetProgramID, Data: base64.StdEncoding.EncodeToString(encodeSetComputeUnitPrice(250_000))},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !budget.HasUnitLimit || !budget.HasUnitPrice || budget.UnitLimit != 500_000 || budget.UnitPriceMicrolamports != 250_000 {
		t.Fatalf("budget = %+v", budget)
	}
	if got := EstimatePriorityFeeLamports(budget.UnitLimit, budget.UnitPriceMicrolamports); got != 125_000 {
		t.Fatalf("fee = %d", got)
	}
}

func computeUnitLimitData(limit uint32) string {
	data := make([]byte, 5)
	data[0] = computeBudgetSetComputeUnitLimit
	binary.LittleEndian.PutUint32(data[1:], limit)
	return base64.StdEncoding.EncodeToString(data)
}
