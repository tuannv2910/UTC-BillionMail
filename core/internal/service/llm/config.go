package llm

import (
	"errors"
	"net/http"
	"strings"
	"time"
)

const (
	defaultMaxTokens = 4096
	defaultTimeout   = 120 * time.Second
)

// Config holds the settings needed to construct a Provider.
type Config struct {
	Provider   string        // registered provider name, see the Provider* constants
	APIKey     string        // vendor API key
	BaseURL    string        // optional; empty uses the vendor's public endpoint
	Model      string        // default model ID for requests that do not set one
	MaxTokens  int           // default output token cap; 0 uses defaultMaxTokens
	Timeout    time.Duration // HTTP timeout; 0 uses defaultTimeout. Ignored when HTTPClient is set
	HTTPClient *http.Client  // optional; injected for proxies or tests
}

// withDefaults returns a copy of cfg with empty optional fields filled in.
func (cfg Config) withDefaults(defaultBaseURL string) Config {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	if cfg.MaxTokens <= 0 {
		cfg.MaxTokens = defaultMaxTokens
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultTimeout
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: cfg.Timeout}
	}
	return cfg
}

// validate checks the fields every provider requires.
func (cfg Config) validate() error {
	if cfg.APIKey == "" {
		return errors.Join(ErrInvalidConfig, errors.New("api key is required"))
	}
	if cfg.BaseURL == "" {
		return errors.Join(ErrInvalidConfig, errors.New("base url is required"))
	}
	if cfg.Model == "" {
		return errors.Join(ErrInvalidConfig, errors.New("model is required"))
	}
	return nil
}

// resolve merges per-request overrides with the provider defaults.
func (cfg Config) resolve(req *GenerateRequest) (model string, maxTokens int) {
	model, maxTokens = cfg.Model, cfg.MaxTokens
	if req.Model != "" {
		model = req.Model
	}
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}
	return model, maxTokens
}
