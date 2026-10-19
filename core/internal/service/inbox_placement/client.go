// Package inbox_placement is the client of the standalone inbox placement
// service (seed mailboxes + Gmail checks). BillionMail only proxies to it:
// the service sends the test through BillionMail's Postfix (SMTP sender) and
// reports where each seed received the message.
package inbox_placement

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	v1 "billionmail-core/api/inbox_placement/v1"
)

// Env variables configuring the client.
const (
	EnvURL        = "INBOX_PLACEMENT_URL"         // e.g. http://inbox-placement:8080
	EnvAPIKey     = "INBOX_PLACEMENT_API_KEY"     // one of the service PLACEMENT_API_KEYS
	EnvAdminToken = "INBOX_PLACEMENT_ADMIN_TOKEN" // service PLACEMENT_ADMIN_TOKEN (seed admin)
)

// ReferencePrefix links a test to a BillionMail template: "template:<id>".
const ReferencePrefix = "template:"

var (
	// ErrDisabled means INBOX_PLACEMENT_URL is not configured.
	ErrDisabled = errors.New("inbox placement service is not configured")
	// ErrAdminDisabled means INBOX_PLACEMENT_ADMIN_TOKEN is not configured.
	ErrAdminDisabled = errors.New("seed administration is not configured")
)

// APIError is an error answered by the placement service.
type APIError struct {
	Status  int
	Code    string
	Message string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("inbox placement service returned %d", e.Status)
}

// Client calls the placement service REST API.
type Client struct {
	baseURL    string
	apiKey     string
	adminToken string
	http       *http.Client
}

// NewClient builds a client.
func NewClient(baseURL, apiKey, adminToken string, timeout time.Duration) *Client {
	return &Client{
		baseURL:    strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		apiKey:     strings.TrimSpace(apiKey),
		adminToken: strings.TrimSpace(adminToken),
		http:       &http.Client{Timeout: timeout},
	}
}

var (
	defaultOnce   sync.Once
	defaultClient *Client
)

// Default returns the client configured from the environment.
func Default() *Client {
	defaultOnce.Do(func() {
		defaultClient = NewClient(os.Getenv(EnvURL), os.Getenv(EnvAPIKey), os.Getenv(EnvAdminToken), 20*time.Second)
	})
	return defaultClient
}

// Enabled reports whether the service URL is configured.
func (c *Client) Enabled() bool { return c.baseURL != "" }

// AdminEnabled reports whether seed administration is possible.
func (c *Client) AdminEnabled() bool { return c.Enabled() && c.adminToken != "" }

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body any, admin bool, out any) error {
	if !c.Enabled() {
		return ErrDisabled
	}
	if admin && c.adminToken == "" {
		return ErrAdminDisabled
	}
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	var rd io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, rd)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if admin {
		req.Header.Set("X-Admin-Token", c.adminToken)
	} else {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("inbox placement service unreachable: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		apiErr := &APIError{Status: resp.StatusCode}
		var e struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal(raw, &e) == nil {
			apiErr.Code, apiErr.Message = e.Error.Code, e.Error.Message
		}
		return apiErr
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("decode inbox placement response: %w", err)
	}
	return nil
}

// Ping checks the service readiness endpoint.
func (c *Client) Ping(ctx context.Context) error {
	return c.do(ctx, http.MethodGet, "/readyz", nil, nil, false, nil)
}

// CreateInput is a new test.
type CreateInput struct {
	Domain     string `json:"domain"`
	From       string `json:"from"`
	Subject    string `json:"subject"`
	HTML       string `json:"html"`
	SeedGroup  string `json:"seed_group,omitempty"`
	TimeoutSec int    `json:"timeout_sec,omitempty"`
	Reference  string `json:"reference,omitempty"`
}

// CreateOutput is the created test.
type CreateOutput struct {
	TestID       string `json:"test_id"`
	Status       string `json:"status"`
	Seeds        int    `json:"seeds"`
	PollAfterSec int    `json:"poll_after_sec"`
}

// CreateTest starts a test.
func (c *Client) CreateTest(ctx context.Context, in CreateInput) (*CreateOutput, error) {
	var out CreateOutput
	if err := c.do(ctx, http.MethodPost, "/v1/placement-tests", nil, in, false, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// remote mirrors the placement service JSON (times are RFC 3339).
type remoteResult struct {
	Seed       string     `json:"seed"`
	Provider   string     `json:"provider"`
	Status     string     `json:"status"`
	SentAt     *time.Time `json:"sent_at"`
	ReceivedAt *time.Time `json:"received_at"`
	LatencyMS  *int64     `json:"latency_ms"`
	SPF        string     `json:"spf"`
	DKIM       string     `json:"dkim"`
	DMARC      string     `json:"dmarc"`
	Labels     []string   `json:"labels"`
	Error      string     `json:"error"`
}

type remoteTest struct {
	TestID      string         `json:"test_id"`
	Domain      string         `json:"domain"`
	From        string         `json:"from"`
	Subject     string         `json:"subject"`
	Status      string         `json:"status"`
	SeedGroup   string         `json:"seed_group"`
	Reference   string         `json:"reference"`
	TimeoutSec  int            `json:"timeout_sec"`
	Error       string         `json:"error"`
	CreatedAt   *time.Time     `json:"created_at"`
	SentAt      *time.Time     `json:"sent_at"`
	CompletedAt *time.Time     `json:"completed_at"`
	Summary     v1.Summary     `json:"summary"`
	Results     []remoteResult `json:"results"`
}

type remoteSeed struct {
	ID              int64      `json:"id"`
	Provider        string     `json:"provider"`
	Email           string     `json:"email"`
	SeedGroup       string     `json:"seed_group"`
	Addressing      string     `json:"addressing"`
	OAuthClient     string     `json:"oauth_client"`
	IsActive        bool       `json:"is_active"`
	HealthStatus    string     `json:"health_status"`
	HealthFailCount int        `json:"health_fail_count"`
	LastHealthAt    *time.Time `json:"last_health_at"`
	LastError       string     `json:"last_error"`
	CreatedAt       *time.Time `json:"created_at"`
}

func unix(t *time.Time) int64 {
	if t == nil || t.IsZero() {
		return 0
	}
	return t.Unix()
}

// TemplateIDFromReference extracts the template id from "template:<id>".
func TemplateIDFromReference(ref string) int {
	if !strings.HasPrefix(ref, ReferencePrefix) {
		return 0
	}
	id, _ := strconv.Atoi(strings.TrimPrefix(ref, ReferencePrefix))
	return id
}

func (r *remoteTest) toV1() *v1.PlacementTest {
	t := &v1.PlacementTest{
		TestID:      r.TestID,
		Domain:      r.Domain,
		From:        r.From,
		Subject:     r.Subject,
		Status:      r.Status,
		SeedGroup:   r.SeedGroup,
		Reference:   r.Reference,
		TemplateId:  TemplateIDFromReference(r.Reference),
		TimeoutSec:  r.TimeoutSec,
		Error:       r.Error,
		CreatedAt:   unix(r.CreatedAt),
		SentAt:      unix(r.SentAt),
		CompletedAt: unix(r.CompletedAt),
		Summary:     r.Summary,
		Results:     make([]v1.SeedResult, 0, len(r.Results)),
	}
	for _, x := range r.Results {
		t.Results = append(t.Results, v1.SeedResult{
			Seed: x.Seed, Provider: x.Provider, Status: x.Status,
			SentAt: unix(x.SentAt), ReceivedAt: unix(x.ReceivedAt), LatencyMS: x.LatencyMS,
			SPF: x.SPF, DKIM: x.DKIM, DMARC: x.DMARC, Labels: x.Labels, Error: x.Error,
		})
	}
	return t
}

func (r *remoteSeed) toV1() *v1.Seed {
	return &v1.Seed{
		Id: r.ID, Provider: r.Provider, Email: r.Email, SeedGroup: r.SeedGroup, Addressing: r.Addressing,
		OAuthClient: r.OAuthClient, IsActive: r.IsActive, HealthStatus: r.HealthStatus,
		HealthFailCount: r.HealthFailCount, LastHealthAt: unix(r.LastHealthAt), LastError: r.LastError,
		CreatedAt: unix(r.CreatedAt),
	}
}

// GetTest returns one test.
func (c *Client) GetTest(ctx context.Context, testID string) (*v1.PlacementTest, error) {
	var out remoteTest
	if err := c.do(ctx, http.MethodGet, "/v1/placement-tests/"+url.PathEscape(testID), nil, nil, false, &out); err != nil {
		return nil, err
	}
	return out.toV1(), nil
}

// ListFilter narrows ListTests.
type ListFilter struct {
	Page       int
	PageSize   int
	TemplateID int
	Domain     string
	Status     string
}

// ListTests returns tests, newest first, and the total count.
func (c *Client) ListTests(ctx context.Context, f ListFilter) ([]*v1.PlacementTest, int, error) {
	q := url.Values{}
	q.Set("page", strconv.Itoa(max(f.Page, 1)))
	q.Set("page_size", strconv.Itoa(max(f.PageSize, 1)))
	if f.TemplateID > 0 {
		q.Set("reference", ReferencePrefix+strconv.Itoa(f.TemplateID))
	}
	if f.Domain != "" {
		q.Set("domain", f.Domain)
	}
	if f.Status != "" {
		q.Set("status", f.Status)
	}
	var out struct {
		Items []remoteTest `json:"items"`
		Total int          `json:"total"`
	}
	if err := c.do(ctx, http.MethodGet, "/v1/placement-tests", q, nil, false, &out); err != nil {
		return nil, 0, err
	}
	list := make([]*v1.PlacementTest, 0, len(out.Items))
	for i := range out.Items {
		list = append(list, out.Items[i].toV1())
	}
	return list, out.Total, nil
}

// CancelTest cancels a running test.
func (c *Client) CancelTest(ctx context.Context, testID string) error {
	return c.do(ctx, http.MethodPost, "/v1/placement-tests/"+url.PathEscape(testID)+"/cancel", nil, nil, false, nil)
}

// ListSeeds returns every seed.
func (c *Client) ListSeeds(ctx context.Context) ([]*v1.Seed, error) {
	var out struct {
		Items []remoteSeed `json:"items"`
	}
	if err := c.do(ctx, http.MethodGet, "/v1/admin/seeds", nil, nil, true, &out); err != nil {
		return nil, err
	}
	list := make([]*v1.Seed, 0, len(out.Items))
	for i := range out.Items {
		list = append(list, out.Items[i].toV1())
	}
	return list, nil
}

// AddSeedInput adds a seed with a refresh token.
type AddSeedInput struct {
	Email        string `json:"email"`
	RefreshToken string `json:"refresh_token"`
	SeedGroup    string `json:"seed_group,omitempty"`
	OAuthClient  string `json:"oauth_client,omitempty"`
}

// AddSeed creates or replaces a seed.
func (c *Client) AddSeed(ctx context.Context, in AddSeedInput) (int64, error) {
	var out struct {
		ID int64 `json:"id"`
	}
	if err := c.do(ctx, http.MethodPost, "/v1/admin/seeds", nil, in, true, &out); err != nil {
		return 0, err
	}
	return out.ID, nil
}

// SeedOAuthURL returns the Google consent URL for a new seed.
func (c *Client) SeedOAuthURL(ctx context.Context, group, client string) (string, error) {
	q := url.Values{"provider": {"gmail"}}
	if group != "" {
		q.Set("group", group)
	}
	if client != "" {
		q.Set("client", client)
	}
	var out struct {
		URL string `json:"url"`
	}
	if err := c.do(ctx, http.MethodPost, "/v1/admin/seeds/oauth/start", q, nil, true, &out); err != nil {
		return "", err
	}
	return out.URL, nil
}

// SeedPatch holds optional seed changes.
type SeedPatch struct {
	IsActive  *bool   `json:"is_active,omitempty"`
	SeedGroup *string `json:"seed_group,omitempty"`
}

// UpdateSeed patches a seed.
func (c *Client) UpdateSeed(ctx context.Context, id int64, p SeedPatch) error {
	return c.do(ctx, http.MethodPatch, "/v1/admin/seeds/"+strconv.FormatInt(id, 10), nil, p, true, nil)
}

// CheckSeed runs a health check on one seed and returns it.
func (c *Client) CheckSeed(ctx context.Context, id int64) (*v1.Seed, error) {
	var out remoteSeed
	if err := c.do(ctx, http.MethodPost, "/v1/admin/seeds/"+strconv.FormatInt(id, 10)+"/health-check", nil, nil, true, &out); err != nil {
		return nil, err
	}
	return out.toV1(), nil
}

// OAuthCallback forwards Google's consent redirect (code + state) to the
// service and returns its HTML answer. The service validates the signed state.
func (c *Client) OAuthCallback(ctx context.Context, rawQuery string) (int, []byte, error) {
	if !c.Enabled() {
		return 0, nil, ErrDisabled
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/admin/seeds/oauth/callback?"+rawQuery, nil)
	if err != nil {
		return 0, nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("inbox placement service unreachable: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	return resp.StatusCode, body, err
}
