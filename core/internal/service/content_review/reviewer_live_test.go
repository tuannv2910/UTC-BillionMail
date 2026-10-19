package content_review

import (
	"billionmail-core/internal/service/llm"
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Calls a real model to check the prompt's quality. Skipped unless configured, e.g.:
//
//	BM_TEST_LLM_PROVIDER=gemini BM_TEST_LLM_KEY=... BM_TEST_LLM_MODEL=... \
//	  go test -run TestReviewLive -v ./internal/service/content_review/
func TestReviewLive(t *testing.T) {
	if testing.Short() {
		t.Skip("live test skipped in -short mode")
	}
	cfg := llm.Config{
		Provider: os.Getenv("BM_TEST_LLM_PROVIDER"),
		APIKey:   os.Getenv("BM_TEST_LLM_KEY"),
		Model:    os.Getenv("BM_TEST_LLM_MODEL"),
		BaseURL:  os.Getenv("BM_TEST_LLM_BASE_URL"),
	}
	if cfg.Provider == "" || cfg.APIKey == "" || cfg.Model == "" {
		t.Skip("BM_TEST_LLM_PROVIDER, BM_TEST_LLM_KEY and BM_TEST_LLM_MODEL not set")
	}
	provider, err := llm.New(cfg)
	require.NoError(t, err)
	reviewer := NewReviewer(llm.WithRetry(provider, 3, 2*time.Second))

	cases := []struct {
		name     string
		in       Input
		minScore float64
		maxScore float64
	}{
		{
			name: "clean newsletter",
			in: Input{
				Subject:            "Your October product update",
				Content:            `<p>Hi {{.Subscriber.Email}},</p><p>This month we shipped dark mode and faster exports. Read the details on our <a href="https://example.com/blog/october">blog</a>.</p><p>Thanks for being a customer.</p>`,
				UnsubscribeManaged: true,
			},
			minScore: 8, maxScore: 10,
		},
		{
			name: "spammy and phishing",
			in: Input{
				Subject:  "RE: URGENT!!! Your account will be SUSPENDED",
				Content:  `<p>Dear customer, verify your bank password within 24 hours or lose access.</p><p>Also EARN $5000 A WEEK from home, 100% GUARANTEED!!!</p><a href="http://192.168.4.20/login">www.yourbank.com</a>`,
				Language: "vi",
			},
			minScore: 0, maxScore: 4,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			res, err := reviewer.Review(ctx, tc.in)
			require.NoError(t, err)

			out, _ := json.MarshalIndent(res, "", "  ")
			t.Logf("%s", out)
			assert.GreaterOrEqual(t, res.Score, tc.minScore)
			assert.LessOrEqual(t, res.Score, tc.maxScore)
		})
	}
}
