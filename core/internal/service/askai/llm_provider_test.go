package askai

import (
	"billionmail-core/internal/service/llm"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLLMConfigFromSupplier(t *testing.T) {
	model := &ModelInfo{ModelId: "m-1"}
	cases := []struct {
		supplier, baseURL  string
		provider, wantBase string
	}{
		{"OpenAI", "https://api.openai.com/v1", llm.ProviderOpenAI, "https://api.openai.com/v1"},
		{"Anthropic", "https://api.anthropic.com/v1/", llm.ProviderAnthropic, "https://api.anthropic.com/v1"},
		{"Gemini", "https://generativelanguage.googleapis.com/v1beta/openai/", llm.ProviderGemini, "https://generativelanguage.googleapis.com/v1beta"},
		{"DeepSeek", " https://api.deepseek.com/v1 ", llm.ProviderOpenAICompatible, "https://api.deepseek.com/v1"},
		{"my-proxy", "https://llm.internal/v1", llm.ProviderOpenAICompatible, "https://llm.internal/v1"},
	}
	for _, tc := range cases {
		t.Run(tc.supplier, func(t *testing.T) {
			cfg := llmConfigFromSupplier(&Supplier{SupplierName: tc.supplier, BaseUrl: tc.baseURL, ApiKey: "k"}, model)
			assert.Equal(t, tc.provider, cfg.Provider)
			assert.Equal(t, tc.wantBase, cfg.BaseURL)
			assert.Equal(t, "k", cfg.APIKey)
			assert.Equal(t, "m-1", cfg.Model)

			_, err := llm.New(cfg)
			assert.NoError(t, err, "config is accepted by the llm registry")
		})
	}
}
