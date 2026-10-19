package v1

import (
	"billionmail-core/utility/types/api_v1"

	"github.com/gogf/gf/v2/frame/g"
)

// Summary aggregates the seed results of one placement test.
type Summary struct {
	Total        int     `json:"total"`
	Pending      int     `json:"pending"`
	Inbox        int     `json:"inbox"`
	Promotions   int     `json:"promotions"`
	Social       int     `json:"social"`
	Updates      int     `json:"updates"`
	Forums       int     `json:"forums"`
	Spam         int     `json:"spam"`
	Trash        int     `json:"trash"`
	Other        int     `json:"other"`
	Missing      int     `json:"missing"`
	Error        int     `json:"error"`
	SendFailed   int     `json:"send_failed"`
	Delivered    int     `json:"delivered"`
	InboxRate    float64 `json:"inbox_rate"`
	InboxAnyRate float64 `json:"inbox_any_rate"`
	SpamRate     float64 `json:"spam_rate"`
}

// SeedResult is the outcome of one seed mailbox.
type SeedResult struct {
	Seed       string   `json:"seed"`
	Provider   string   `json:"provider"`
	Status     string   `json:"status"`
	SentAt     int64    `json:"sent_at" dc:"Unix seconds, 0 when unset"`
	ReceivedAt int64    `json:"received_at" dc:"Unix seconds, 0 when unset"`
	LatencyMS  *int64   `json:"latency_ms,omitempty"`
	SPF        string   `json:"spf,omitempty"`
	DKIM       string   `json:"dkim,omitempty"`
	DMARC      string   `json:"dmarc,omitempty"`
	Labels     []string `json:"labels,omitempty"`
	Error      string   `json:"error,omitempty"`
}

// PlacementTest is one inbox placement test.
type PlacementTest struct {
	TestID      string       `json:"test_id"`
	Domain      string       `json:"domain"`
	From        string       `json:"from"`
	Subject     string       `json:"subject"`
	Status      string       `json:"status" dc:"pending|sending|checking|completed|failed|canceled"`
	SeedGroup   string       `json:"seed_group"`
	Reference   string       `json:"reference,omitempty"`
	TemplateId  int          `json:"template_id" dc:"Template ID parsed from reference (0 when not linked)"`
	TimeoutSec  int          `json:"timeout_sec"`
	Error       string       `json:"error,omitempty"`
	CreatedAt   int64        `json:"created_at" dc:"Unix seconds, 0 when unset"`
	SentAt      int64        `json:"sent_at" dc:"Unix seconds, 0 when unset"`
	CompletedAt int64        `json:"completed_at" dc:"Unix seconds, 0 when unset"`
	Summary     Summary      `json:"summary"`
	Results     []SeedResult `json:"results"`
}

// Seed is one seed mailbox (secrets are never returned).
type Seed struct {
	Id              int64  `json:"id"`
	Provider        string `json:"provider"`
	Email           string `json:"email"`
	SeedGroup       string `json:"seed_group"`
	Addressing      string `json:"addressing"`
	OAuthClient     string `json:"oauth_client,omitempty"`
	IsActive        bool   `json:"is_active"`
	HealthStatus    string `json:"health_status"`
	HealthFailCount int    `json:"health_fail_count"`
	LastHealthAt    int64  `json:"last_health_at" dc:"Unix seconds, 0 when unset"`
	LastError       string `json:"last_error,omitempty"`
	CreatedAt       int64  `json:"created_at" dc:"Unix seconds, 0 when unset"`
}

// StatusReq reports whether the placement service is configured and reachable.
type StatusReq struct {
	g.Meta        `path:"/inbox_placement/status" method:"get" tags:"InboxPlacement" summary:"Inbox placement service status"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
}

type StatusRes struct {
	api_v1.StandardRes
	Data struct {
		Enabled   bool `json:"enabled" dc:"INBOX_PLACEMENT_URL is configured"`
		Reachable bool `json:"reachable" dc:"Service answered its readiness check"`
		AdminSeed bool `json:"admin_seed" dc:"Seed administration is available (admin token configured)"`
	} `json:"data"`
}

// CreateTestReq starts a placement test from a template.
type CreateTestReq struct {
	g.Meta        `path:"/inbox_placement/create" method:"post" tags:"InboxPlacement" summary:"Create inbox placement test"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	TemplateId    int    `json:"template_id" v:"required|min:1" dc:"Email template ID"`
	Addresser     string `json:"addresser" v:"required|email" dc:"Sender mailbox"`
	FullName      string `json:"full_name" dc:"Sender display name"`
	Subject       string `json:"subject" v:"required|max-length:998" dc:"Email subject"`
	SeedGroup     string `json:"seed_group" dc:"Seed group (default: default)"`
	TimeoutSec    int    `json:"timeout_sec" v:"min:0|max:1800" dc:"Seconds to wait before marking missing (120-1800, 0 = service default)"`
}

type CreateTestRes struct {
	api_v1.StandardRes
	Data struct {
		TestId       string `json:"test_id"`
		Status       string `json:"status"`
		Seeds        int    `json:"seeds"`
		PollAfterSec int    `json:"poll_after_sec"`
	} `json:"data"`
}

// ListTestsReq lists placement tests.
type ListTestsReq struct {
	g.Meta        `path:"/inbox_placement/list" method:"get" tags:"InboxPlacement" summary:"List inbox placement tests"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	Page          int    `json:"page" v:"min:1" dc:"Page number" default:"1"`
	PageSize      int    `json:"page_size" v:"min:1|max:100" dc:"Page size" default:"10"`
	TemplateId    int    `json:"template_id" dc:"Filter by template ID"`
	Domain        string `json:"domain" dc:"Filter by sending domain"`
	Status        string `json:"status" dc:"Filter by status"`
}

type ListTestsRes struct {
	api_v1.StandardRes
	Data struct {
		Total int              `json:"total"`
		List  []*PlacementTest `json:"list"`
	} `json:"data"`
}

// GetTestReq returns one placement test.
type GetTestReq struct {
	g.Meta        `path:"/inbox_placement/detail" method:"get" tags:"InboxPlacement" summary:"Get inbox placement test"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	TestId        string `json:"test_id" v:"required" dc:"Test ID"`
}

type GetTestRes struct {
	api_v1.StandardRes
	Data *PlacementTest `json:"data"`
}

// CancelTestReq cancels a running placement test.
type CancelTestReq struct {
	g.Meta        `path:"/inbox_placement/cancel" method:"post" tags:"InboxPlacement" summary:"Cancel inbox placement test"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	TestId        string `json:"test_id" v:"required" dc:"Test ID"`
}

type CancelTestRes struct {
	api_v1.StandardRes
}

// ListSeedsReq lists seed mailboxes.
type ListSeedsReq struct {
	g.Meta        `path:"/inbox_placement/seed_list" method:"get" tags:"InboxPlacement" summary:"List seed mailboxes"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
}

type ListSeedsRes struct {
	api_v1.StandardRes
	Data struct {
		Total int     `json:"total"`
		List  []*Seed `json:"list"`
	} `json:"data"`
}

// AddSeedReq adds a Gmail seed from an existing refresh token.
type AddSeedReq struct {
	g.Meta        `path:"/inbox_placement/seed_add" method:"post" tags:"InboxPlacement" summary:"Add seed mailbox with refresh token"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	Email         string `json:"email" v:"required|email" dc:"Seed mailbox"`
	RefreshToken  string `json:"refresh_token" v:"required" dc:"Google OAuth refresh token"`
	SeedGroup     string `json:"seed_group" dc:"Seed group"`
	OAuthClient   string `json:"oauth_client" dc:"OAuth client name configured in the placement service"`
}

type AddSeedRes struct {
	api_v1.StandardRes
	Data struct {
		Id int64 `json:"id"`
	} `json:"data"`
}

// SeedOAuthReq returns the Google consent URL to connect a new seed.
type SeedOAuthReq struct {
	g.Meta        `path:"/inbox_placement/seed_oauth" method:"post" tags:"InboxPlacement" summary:"Start Google consent for a seed"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	SeedGroup     string `json:"seed_group" dc:"Seed group"`
	OAuthClient   string `json:"oauth_client" dc:"OAuth client name"`
}

type SeedOAuthRes struct {
	api_v1.StandardRes
	Data struct {
		Url string `json:"url"`
	} `json:"data"`
}

// UpdateSeedReq enables/disables a seed or moves it to another group.
type UpdateSeedReq struct {
	g.Meta        `path:"/inbox_placement/seed_update" method:"post" tags:"InboxPlacement" summary:"Update seed mailbox"`
	Authorization string  `json:"authorization" dc:"Authorization" in:"header"`
	Id            int64   `json:"id" v:"required|min:1" dc:"Seed ID"`
	IsActive      *bool   `json:"is_active" dc:"Enable or disable"`
	SeedGroup     *string `json:"seed_group" dc:"Seed group"`
}

type UpdateSeedRes struct {
	api_v1.StandardRes
}

// CheckSeedReq runs a health check on one seed now.
type CheckSeedReq struct {
	g.Meta        `path:"/inbox_placement/seed_check" method:"post" tags:"InboxPlacement" summary:"Health check seed mailbox"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	Id            int64  `json:"id" v:"required|min:1" dc:"Seed ID"`
}

type CheckSeedRes struct {
	api_v1.StandardRes
	Data *Seed `json:"data"`
}
