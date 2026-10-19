package askai

import (
	"billionmail-core/internal/service/llm"
	"errors"
	"strings"
	"time"
)

var (
	ErrNoActiveModel         = errors.New("no AI model is enabled; configure one in Settings > AI Model")
	ErrSupplierNotConfigured = errors.New("AI supplier is not configured or disabled")
	ErrModelNotFound         = errors.New("AI model not found or disabled")
)

// NewLLMProvider builds an llm.Provider from a configured supplier and model, so features can use
// the API keys entered under Settings > AI Model. With an empty supplier or model, the first
// enabled model is used. It returns the provider and the resolved model.
func NewLLMProvider(supplierName, modelId string) (llm.Provider, *ModelInfo, error) {
	if supplierName == "" || modelId == "" {
		models := Models("")
		if len(models) == 0 {
			return nil, nil, ErrNoActiveModel
		}
		supplierName, modelId = models[0].SupplierName, models[0].ModelId
	}

	supplier, err := GetSupplierConfig(supplierName)
	if err != nil || !supplier.Status || supplier.ApiKey == "" || supplier.BaseUrl == "" {
		return nil, nil, ErrSupplierNotConfigured
	}
	model := GetModelInfo(supplierName, modelId)
	if model == nil || !model.Status {
		return nil, nil, ErrModelNotFound
	}

	provider, err := llm.New(llmConfigFromSupplier(supplier, model))
	if err != nil {
		return nil, nil, err
	}
	return llm.WithRetry(provider, 3, 2*time.Second), model, nil
}

// llmConfigFromSupplier maps a supplier to the native llm provider. Supplier base URLs point at the
// vendors' OpenAI-compatible endpoints, which the native Anthropic and Gemini providers do not use.
func llmConfigFromSupplier(s *Supplier, m *ModelInfo) llm.Config {
	baseURL := strings.TrimRight(strings.TrimSpace(s.BaseUrl), "/")
	cfg := llm.Config{APIKey: s.ApiKey, Model: m.ModelId, BaseURL: baseURL}

	switch strings.ToLower(s.SupplierName) {
	case "openai":
		cfg.Provider = llm.ProviderOpenAI
	case "anthropic":
		// https://api.anthropic.com/v1 serves both /chat/completions and the native /messages.
		cfg.Provider = llm.ProviderAnthropic
	case "gemini":
		// .../v1beta/openai is the compatibility layer; the native API lives at .../v1beta.
		cfg.Provider = llm.ProviderGemini
		cfg.BaseURL = strings.TrimSuffix(baseURL, "/openai")
	default:
		cfg.Provider = llm.ProviderOpenAICompatible
	}
	return cfg
}
