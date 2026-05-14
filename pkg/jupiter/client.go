package jupiter

import (
	"log/slog"
	"net/http"

	"github.com/0xdamahou/jup-go/internal/juphttp"
	"github.com/0xdamahou/jup-go/pkg/lend"
	"github.com/0xdamahou/jup-go/pkg/perps"
	"github.com/0xdamahou/jup-go/pkg/portfolio"
	"github.com/0xdamahou/jup-go/pkg/prediction"
	"github.com/0xdamahou/jup-go/pkg/price"
	"github.com/0xdamahou/jup-go/pkg/recurring"
	"github.com/0xdamahou/jup-go/pkg/swap"
	"github.com/0xdamahou/jup-go/pkg/token"
	"github.com/0xdamahou/jup-go/pkg/trigger"
)

type Config = juphttp.Config
type APIError = juphttp.APIError

// Client is the top-level Jupiter SDK client.
type Client struct {
	Swap       *swap.Client
	Token      *token.Client
	Price      *price.Client
	Trigger    *trigger.Client
	Recurring  *recurring.Client
	Lend       *lend.Client
	Portfolio  *portfolio.Client
	Perps      *perps.Client
	Prediction *prediction.Client
}

// Option customizes the top-level client.
type Option func(*options)

type options struct {
	httpClient *http.Client
	logger     *slog.Logger
}

func WithHTTPClient(h *http.Client) Option { return func(o *options) { o.httpClient = h } }
func WithLogger(l *slog.Logger) Option     { return func(o *options) { o.logger = l } }

// NewClient creates a Jupiter SDK client and all subclients.
func NewClient(cfg Config, opts ...Option) *Client {
	var o options
	for _, opt := range opts {
		opt(&o)
	}
	httpOpts := []juphttp.Option{}
	if o.httpClient != nil {
		httpOpts = append(httpOpts, juphttp.WithHTTPClient(o.httpClient))
	}
	if o.logger != nil {
		httpOpts = append(httpOpts, juphttp.WithLogger(o.logger))
	}
	h := juphttp.NewClient(cfg, httpOpts...)
	return &Client{
		Swap:       swap.NewClient(h),
		Token:      token.NewClient(h),
		Price:      price.NewClient(h),
		Trigger:    trigger.NewClient(h),
		Recurring:  recurring.NewClient(h),
		Lend:       lend.NewClient(h),
		Portfolio:  portfolio.NewClient(h),
		Perps:      perps.NewClient(h),
		Prediction: prediction.NewClient(h),
	}
}

func ConfigFromEnv() Config { return juphttp.ConfigFromEnv() }
func LoadConfigFile(path string) (Config, error) {
	return juphttp.LoadConfigFile(path)
}
func RedactSecret(s string) string { return juphttp.RedactSecret(s) }
func IsRetryable(err error) bool   { return juphttp.IsRetryable(err) }
