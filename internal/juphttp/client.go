package juphttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxBodyBytes = 8 << 20

// Client is the shared HTTP transport for all Jupiter API modules.
type Client struct {
	cfg    Config
	http   *http.Client
	logger *slog.Logger
}

// NewClient creates a shared Jupiter HTTP client.
func NewClient(cfg Config, opts ...Option) *Client {
	cfg = cfg.WithDefaults()
	c := &Client{
		cfg: cfg,
		http: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: slog.Default(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Option customizes the shared client.
type Option func(*Client)

// WithHTTPClient injects an HTTP client, typically for tests.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) {
		if h != nil {
			c.http = h
		}
	}
}

// WithLogger injects a structured logger.
func WithLogger(l *slog.Logger) Option {
	return func(c *Client) {
		if l != nil {
			c.logger = l
		}
	}
}

func (c *Client) Config() Config { return c.cfg }

// GetJSON performs an authenticated GET and decodes JSON.
func (c *Client) GetJSON(ctx context.Context, base, path string, q url.Values, out any) error {
	return c.doJSON(ctx, http.MethodGet, base, path, q, nil, out, true)
}

// PostJSON performs an authenticated POST and decodes JSON.
func (c *Client) PostJSON(ctx context.Context, base, path string, in, out any, retrySafe bool) error {
	return c.doJSON(ctx, http.MethodPost, base, path, nil, in, out, retrySafe)
}

// PatchJSON performs an authenticated PATCH and decodes JSON.
func (c *Client) PatchJSON(ctx context.Context, base, path string, in, out any) error {
	return c.doJSON(ctx, http.MethodPatch, base, path, nil, in, out, false)
}

func (c *Client) doJSON(ctx context.Context, method, base, path string, q url.Values, in, out any, retrySafe bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	var bodyBytes []byte
	var err error
	if in != nil {
		bodyBytes, err = json.Marshal(in)
		if err != nil {
			return err
		}
	}
	attempts := 1
	if retrySafe {
		attempts += c.cfg.MaxRetries
	}
	var last error
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			if err := sleepBackoff(ctx, c.cfg.RetryBackoff, attempt); err != nil {
				return err
			}
		}
		err := c.once(ctx, method, base, path, q, bodyBytes, out, attempt)
		if err == nil {
			return nil
		}
		last = err
		if !isRetryableErr(err) || attempt == attempts-1 {
			break
		}
	}
	return last
}

func (c *Client) once(ctx context.Context, method, base, path string, q url.Values, body []byte, out any, attempt int) error {
	u, err := url.Parse(strings.TrimRight(base, "/") + "/" + strings.TrimLeft(path, "/"))
	if err != nil {
		return err
	}
	if len(q) > 0 {
		u.RawQuery = q.Encode()
	}
	var r io.Reader
	if body != nil {
		r = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), r)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.cfg.APIKey != "" {
		req.Header.Set("x-api-key", c.cfg.APIKey)
	}

	start := time.Now()
	resp, err := c.http.Do(req)
	duration := time.Since(start)
	if err != nil {
		c.log(method, path, 0, duration, attempt, "", err)
		if isNetRetryable(err) {
			return &APIError{Endpoint: path, Message: err.Error(), Retryable: true}
		}
		return err
	}
	defer resp.Body.Close()

	requestID := resp.Header.Get("x-request-id")
	limited := io.LimitReader(resp.Body, maxBodyBytes+1)
	raw, err := io.ReadAll(limited)
	if err != nil {
		return err
	}
	if len(raw) > maxBodyBytes {
		return &APIError{StatusCode: resp.StatusCode, Endpoint: path, RequestID: requestID, Message: "response body too large"}
	}
	c.log(method, path, resp.StatusCode, duration, attempt, requestID, nil)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decodeAPIError(resp.StatusCode, path, requestID, raw)
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		dec = json.NewDecoder(bytes.NewReader(raw))
		if err2 := dec.Decode(out); err2 != nil {
			return fmt.Errorf("decode %s: %w", path, err)
		}
	}
	return nil
}

func decodeAPIError(status int, endpoint, requestID string, raw []byte) error {
	var payload struct {
		Code    any    `json:"code"`
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(raw, &payload)
	msg := payload.Message
	if msg == "" {
		msg = payload.Error
	}
	if msg == "" {
		msg = strings.TrimSpace(string(raw))
	}
	if msg == "" {
		msg = http.StatusText(status)
	}
	code := ""
	if payload.Code != nil {
		code = fmt.Sprint(payload.Code)
	}
	return &APIError{
		StatusCode: status,
		Code:       code,
		Message:    msg,
		RequestID:  requestID,
		Endpoint:   endpoint,
		Body:       string(raw),
		Retryable:  status == http.StatusTooManyRequests || status >= 500,
	}
}

func sleepBackoff(ctx context.Context, base time.Duration, attempt int) error {
	d := base * time.Duration(1<<min(attempt-1, 5))
	jitter := time.Duration(rand.Int63n(int64(max(d/2, time.Millisecond))))
	t := time.NewTimer(d + jitter)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func isRetryableErr(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Retryable
	}
	return isNetRetryable(err)
}

func isNetRetryable(err error) bool {
	var ne net.Error
	return errors.As(err, &ne) && ne.Timeout()
}

func (c *Client) log(method, endpoint string, status int, duration time.Duration, retry int, requestID string, err error) {
	if c.logger == nil {
		return
	}
	attrs := []any{"method", method, "endpoint", endpoint, "status", status, "duration_ms", duration.Milliseconds(), "retry", retry}
	if requestID != "" {
		attrs = append(attrs, "request_id", requestID)
	}
	if err != nil {
		attrs = append(attrs, "error", err.Error())
	}
	c.logger.Debug("jupiter api request", attrs...)
}
