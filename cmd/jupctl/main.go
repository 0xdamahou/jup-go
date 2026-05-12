package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/0xdamahou/jup-go/pkg/jupiter"
	"github.com/0xdamahou/jup-go/pkg/price"
	"github.com/0xdamahou/jup-go/pkg/swap"
	"github.com/0xdamahou/jup-go/pkg/token"
)

type globalFlags struct {
	config    string
	apiKeyEnv string
	jsonOut   bool
	dryRun    bool
	timeout   time.Duration
	baseURL   string
	liteURL   string
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	g, rest, err := parseGlobal(args)
	if err != nil {
		return err
	}
	if len(rest) < 2 {
		return fmt.Errorf("usage: jupctl [--config file] [--json] [--dry-run] <group> <command>")
	}
	cfg := jupiter.ConfigFromEnv()
	if g.config != "" {
		cfg, err = jupiter.LoadConfigFile(g.config)
		if err != nil {
			return err
		}
	}
	if g.apiKeyEnv != "" {
		cfg.APIKey = os.Getenv(g.apiKeyEnv)
	}
	if g.timeout > 0 {
		cfg.Timeout = g.timeout
	}
	if g.baseURL != "" {
		cfg.BaseURL = g.baseURL
	}
	if g.liteURL != "" {
		cfg.LiteBaseURL = g.liteURL
	}
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	client := jupiter.NewClient(cfg)
	switch rest[0] + " " + rest[1] {
	case "swap order":
		return swapOrder(ctx, client, rest[2:], g)
	case "swap build":
		return swapBuild(ctx, client, rest[2:], g)
	case "swap execute":
		return swapExecute(ctx, client, rest[2:], g)
	case "token search":
		return tokenSearch(ctx, client, rest[2:], g)
	case "price get":
		return priceGet(ctx, client, rest[2:], g)
	case "trigger create", "trigger cancel", "recurring create", "recurring cancel":
		if g.dryRun {
			return print(g, map[string]any{"dryRun": true, "command": rest[0] + " " + rest[1]})
		}
		return fmt.Errorf("%s requires signing/vault workflow parameters; use SDK package for production flow", rest[0]+" "+rest[1])
	case "lend markets":
		out, err := client.Lend.EarnTokens(ctx)
		if err != nil {
			return err
		}
		return print(g, out)
	case "portfolio holdings":
		fs := flag.NewFlagSet("portfolio holdings", flag.ContinueOnError)
		wallet := fs.String("wallet", "", "wallet public key")
		if err := fs.Parse(rest[2:]); err != nil {
			return err
		}
		out, err := client.Portfolio.Holdings(ctx, *wallet)
		if err != nil {
			return err
		}
		return print(g, out)
	case "perps markets":
		out, err := client.Perps.Markets(ctx)
		if err != nil {
			return err
		}
		return print(g, out)
	case "prediction markets":
		out, err := client.Prediction.Markets(ctx)
		if err != nil {
			return err
		}
		return print(g, out)
	default:
		return fmt.Errorf("unknown command %q", rest[0]+" "+rest[1])
	}
}

func parseGlobal(args []string) (globalFlags, []string, error) {
	fs := flag.NewFlagSet("jupctl", flag.ContinueOnError)
	var g globalFlags
	fs.StringVar(&g.config, "config", "", "config file")
	fs.StringVar(&g.apiKeyEnv, "api-key-env", "JUPITER_API_KEY", "api key environment variable")
	fs.BoolVar(&g.jsonOut, "json", false, "print JSON")
	fs.BoolVar(&g.dryRun, "dry-run", false, "validate without submitting")
	fs.DurationVar(&g.timeout, "timeout", 0, "request timeout")
	fs.StringVar(&g.baseURL, "base-url", "", "Jupiter API base URL")
	fs.StringVar(&g.liteURL, "lite-base-url", "", "Jupiter lite API base URL")
	if err := fs.Parse(args); err != nil {
		return g, nil, err
	}
	return g, fs.Args(), nil
}

func swapOrder(ctx context.Context, c *jupiter.Client, args []string, g globalFlags) error {
	fs := flag.NewFlagSet("swap order", flag.ContinueOnError)
	input := fs.String("input-mint", "", "input mint")
	output := fs.String("output-mint", "", "output mint")
	amount := fs.String("amount", "", "raw amount")
	taker := fs.String("taker", "", "taker wallet")
	if err := fs.Parse(args); err != nil {
		return err
	}
	out, err := c.Swap.GetOrder(ctx, swap.GetOrderRequest{InputMint: *input, OutputMint: *output, Amount: *amount, Taker: *taker})
	if err != nil {
		return err
	}
	return print(g, out)
}

func swapBuild(ctx context.Context, c *jupiter.Client, args []string, g globalFlags) error {
	fs := flag.NewFlagSet("swap build", flag.ContinueOnError)
	input := fs.String("input-mint", "", "input mint")
	output := fs.String("output-mint", "", "output mint")
	amount := fs.String("amount", "", "raw amount")
	taker := fs.String("taker", "", "taker wallet")
	if err := fs.Parse(args); err != nil {
		return err
	}
	out, err := c.Swap.GetBuild(ctx, swap.BuildRequest{InputMint: *input, OutputMint: *output, Amount: *amount, Taker: *taker})
	if err != nil {
		return err
	}
	return print(g, out)
}

func swapExecute(ctx context.Context, c *jupiter.Client, args []string, g globalFlags) error {
	fs := flag.NewFlagSet("swap execute", flag.ContinueOnError)
	tx := fs.String("signed-transaction", "", "signed transaction base64")
	requestID := fs.String("request-id", "", "order request id")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if g.dryRun {
		return print(g, map[string]any{"dryRun": true, "hasSignedTransaction": *tx != "", "requestId": *requestID})
	}
	out, err := c.Swap.Execute(ctx, swap.ExecuteRequest{SignedTransaction: *tx, RequestID: *requestID})
	if err != nil {
		return err
	}
	return print(g, out)
}

func tokenSearch(ctx context.Context, c *jupiter.Client, args []string, g globalFlags) error {
	fs := flag.NewFlagSet("token search", flag.ContinueOnError)
	query := fs.String("query", "", "query")
	if err := fs.Parse(args); err != nil {
		return err
	}
	out, err := c.Token.Search(ctx, token.SearchRequest{Query: *query})
	if err != nil {
		return err
	}
	return print(g, out)
}

func priceGet(ctx context.Context, c *jupiter.Client, args []string, g globalFlags) error {
	fs := flag.NewFlagSet("price get", flag.ContinueOnError)
	var ids repeated
	fs.Func("id", "mint id, repeatable", func(v string) error {
		ids = append(ids, v)
		return nil
	})
	if err := fs.Parse(args); err != nil {
		return err
	}
	out, err := c.Price.Get(ctx, price.GetRequest{IDs: ids})
	if err != nil {
		return err
	}
	return print(g, out)
}

type repeated []string

func print(g globalFlags, v any) error {
	if g.jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	}
	fmt.Printf("%+v\n", v)
	return nil
}
