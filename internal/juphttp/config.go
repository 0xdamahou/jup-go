package juphttp

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultBaseURL    = "https://api.jup.ag"
	DefaultLiteURL    = "https://lite-api.jup.ag"
	DefaultTimeout    = 10 * time.Second
	DefaultMaxRetries = 2
	DefaultBackoff    = 250 * time.Millisecond
	DefaultUserAgent  = "jup-go/0.1"
)

const unsetMaxRetries = -1

// Config contains shared Jupiter client settings.
type Config struct {
	APIKey          string        `json:"apiKey"`
	BaseURL         string        `json:"baseURL"`
	LiteBaseURL     string        `json:"liteBaseURL"`
	SolanaRPCURL    string        `json:"solanaRPCURL"`
	Timeout         time.Duration `json:"timeout"`
	MaxRetries      int           `json:"maxRetries"`
	RetryBackoff    time.Duration `json:"retryBackoff"`
	UserAgent       string        `json:"userAgent"`
	ReferralAccount string        `json:"referralAccount"`
	ReferralFeeBPS  int           `json:"referralFeeBps"`
	Payer           string        `json:"payer"`
}

// WithDefaults fills unset fields with production-safe defaults.
func (c Config) WithDefaults() Config {
	if c.BaseURL == "" {
		c.BaseURL = DefaultBaseURL
	}
	if c.LiteBaseURL == "" {
		c.LiteBaseURL = DefaultLiteURL
	}
	if c.Timeout <= 0 {
		c.Timeout = DefaultTimeout
	}
	if c.MaxRetries < 0 {
		c.MaxRetries = 0
	}
	if c.RetryBackoff <= 0 {
		c.RetryBackoff = DefaultBackoff
	}
	if c.UserAgent == "" {
		c.UserAgent = DefaultUserAgent
	}
	return c
}

// ConfigFromEnv loads common settings from environment variables.
func ConfigFromEnv() Config {
	timeout := parseDurationEnv("JUPITER_TIMEOUT", 0)
	backoff := parseDurationEnv("JUPITER_RETRY_BACKOFF", 0)
	retries := DefaultMaxRetries
	if raw := os.Getenv("JUPITER_MAX_RETRIES"); raw != "" {
		retries, _ = strconv.Atoi(raw)
	}
	fee, _ := strconv.Atoi(os.Getenv("JUPITER_REFERRAL_FEE_BPS"))
	return Config{
		APIKey:          os.Getenv("JUPITER_API_KEY"),
		BaseURL:         os.Getenv("JUPITER_BASE_URL"),
		LiteBaseURL:     os.Getenv("JUPITER_LITE_BASE_URL"),
		SolanaRPCURL:    os.Getenv("SOLANA_RPC_URL"),
		Timeout:         timeout,
		MaxRetries:      retries,
		RetryBackoff:    backoff,
		UserAgent:       os.Getenv("JUPITER_USER_AGENT"),
		ReferralAccount: os.Getenv("JUPITER_REFERRAL_ACCOUNT"),
		ReferralFeeBPS:  fee,
		Payer:           os.Getenv("JUPITER_PAYER"),
	}.WithDefaults()
}

func parseDurationEnv(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	d, err := time.ParseDuration(raw)
	if err == nil {
		return d
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}

// LoadConfigFile loads a JSON file or a simple flat YAML file.
func LoadConfigFile(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	if strings.HasSuffix(path, ".json") {
		cfg := Config{MaxRetries: unsetMaxRetries}
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return Config{}, err
		}
		if cfg.MaxRetries == unsetMaxRetries {
			cfg.MaxRetries = DefaultMaxRetries
		}
		return cfg.WithDefaults(), nil
	}
	return parseFlatYAML(string(raw))
}

func parseFlatYAML(raw string) (Config, error) {
	values := map[string]string{}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			return Config{}, errors.New("unsupported yaml config: expected flat key: value entries")
		}
		values[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"'`)
	}
	cfg := Config{
		APIKey:          values["apiKey"],
		BaseURL:         values["baseURL"],
		LiteBaseURL:     values["liteBaseURL"],
		SolanaRPCURL:    values["solanaRPCURL"],
		UserAgent:       values["userAgent"],
		ReferralAccount: values["referralAccount"],
		Payer:           values["payer"],
		MaxRetries:      DefaultMaxRetries,
	}
	if v := values["timeout"]; v != "" {
		cfg.Timeout, _ = time.ParseDuration(v)
	}
	if v := values["retryBackoff"]; v != "" {
		cfg.RetryBackoff, _ = time.ParseDuration(v)
	}
	if v := values["maxRetries"]; v != "" {
		cfg.MaxRetries, _ = strconv.Atoi(v)
	}
	cfg.ReferralFeeBPS, _ = strconv.Atoi(values["referralFeeBps"])
	return cfg.WithDefaults(), nil
}

// RedactSecret returns a value suitable for logs.
func RedactSecret(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 8 {
		return "[redacted]"
	}
	return s[:4] + "..." + s[len(s)-4:]
}
