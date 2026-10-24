package domains

import (
	"billionmail-core/internal/model"
	"billionmail-core/internal/service/public"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Webhook alert channel for blacklist notifications (#382). It lets users
// forward alerts to Feishu/Lark, WeCom, DingTalk, Slack, Gotify or any HTTP
// endpoint without configuring an SMTP account.

const AlertWebhookOptionName = "blacklist_alert_webhook"

// Supported payload formats.
const (
	WebhookFormatJSON     = "json"
	WebhookFormatFeishu   = "feishu"
	WebhookFormatWeCom    = "wecom"
	WebhookFormatDingTalk = "dingtalk"
	WebhookFormatSlack    = "slack"
	WebhookFormatGotify   = "gotify"
)

var webhookFormats = map[string]bool{
	WebhookFormatJSON: true, WebhookFormatFeishu: true, WebhookFormatWeCom: true,
	WebhookFormatDingTalk: true, WebhookFormatSlack: true, WebhookFormatGotify: true,
}

type AlertWebhookSettings struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url"`
	Format  string `json:"format"`
}

// AlertEvent is the generic JSON payload ("json" format).
type AlertEvent struct {
	Event       string                  `json:"event"`
	Title       string                  `json:"title"`
	Message     string                  `json:"message"`
	Domain      string                  `json:"domain,omitempty"`
	IP          string                  `json:"ip,omitempty"`
	Blacklisted int                     `json:"blacklisted,omitempty"`
	Tested      int                     `json:"tested,omitempty"`
	Lists       []model.BlacklistDetail `json:"lists,omitempty"`
	Time        int64                   `json:"time"`
	Extra       map[string]interface{}  `json:"extra,omitempty"`
}

var webhookHTTPClient = &http.Client{Timeout: 10 * time.Second}

// ValidateAlertWebhook checks URL scheme and format.
func ValidateAlertWebhook(s AlertWebhookSettings) error {
	if !s.Enabled && strings.TrimSpace(s.URL) == "" {
		return nil
	}
	u, err := url.Parse(strings.TrimSpace(s.URL))
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("webhook URL must be an absolute http(s) URL")
	}
	if s.Format != "" && !webhookFormats[s.Format] {
		return fmt.Errorf("unsupported webhook format: %s", s.Format)
	}
	return nil
}

// LoadAlertWebhookSettings reads the stored webhook settings.
func LoadAlertWebhookSettings(ctx context.Context) AlertWebhookSettings {
	var s AlertWebhookSettings
	if err := public.OptionsMgrInstance.GetOption(ctx, AlertWebhookOptionName, &s); err != nil {
		return AlertWebhookSettings{Format: WebhookFormatJSON}
	}
	if s.Format == "" {
		s.Format = WebhookFormatJSON
	}
	return s
}

// SaveAlertWebhookSettings validates and stores webhook settings.
func SaveAlertWebhookSettings(ctx context.Context, s AlertWebhookSettings) error {
	s.URL = strings.TrimSpace(s.URL)
	if s.Format == "" {
		s.Format = WebhookFormatJSON
	}
	if err := ValidateAlertWebhook(s); err != nil {
		return err
	}
	return public.OptionsMgrInstance.SetOption(ctx, AlertWebhookOptionName, s)
}

// BuildWebhookPayload renders the request body for the given format.
func BuildWebhookPayload(format string, ev AlertEvent) ([]byte, error) {
	text := ev.Title
	if ev.Message != "" {
		text += "\n" + ev.Message
	}

	var body interface{}
	switch format {
	case WebhookFormatFeishu:
		body = map[string]interface{}{"msg_type": "text", "content": map[string]string{"text": text}}
	case WebhookFormatWeCom:
		body = map[string]interface{}{"msgtype": "text", "text": map[string]string{"content": text}}
	case WebhookFormatDingTalk:
		body = map[string]interface{}{"msgtype": "text", "text": map[string]string{"content": text}}
	case WebhookFormatSlack:
		body = map[string]string{"text": text}
	case WebhookFormatGotify:
		body = map[string]interface{}{"title": ev.Title, "message": ev.Message, "priority": 8}
	default:
		body = ev
	}
	return json.Marshal(body)
}

// SendAlertWebhook posts an alert to the configured webhook.
func SendAlertWebhook(ctx context.Context, s AlertWebhookSettings, ev AlertEvent) error {
	if err := ValidateAlertWebhook(s); err != nil {
		return err
	}
	if ev.Time == 0 {
		ev.Time = time.Now().Unix()
	}
	payload, err := BuildWebhookPayload(s.Format, ev)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.URL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "BillionMail-Alert")

	resp, err := webhookHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64*1024))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned HTTP %d", resp.StatusCode)
	}
	return nil
}

// sendBlacklistWebhook dispatches a blacklist alert when the webhook channel is enabled.
func sendBlacklistWebhook(ctx context.Context, ip, domain string, result *model.BlacklistCheckResult) {
	s := LoadAlertWebhookSettings(ctx)
	if !s.Enabled || s.URL == "" {
		return
	}

	ev := AlertEvent{
		Event:  "blacklist",
		Title:  fmt.Sprintf("Blacklist Alert - %s", domain),
		Domain: domain,
		IP:     ip,
		Time:   time.Now().Unix(),
	}
	var names []string
	if result != nil {
		ev.Blacklisted = result.Blacklisted
		ev.Tested = result.Tested
		ev.Lists = result.BlackList
		for _, d := range result.BlackList {
			names = append(names, d.Blacklist)
		}
	}
	ev.Message = fmt.Sprintf("IP %s of domain %s is listed on %d of %d blacklists: %s",
		ip, domain, ev.Blacklisted, ev.Tested, strings.Join(names, ", "))

	if err := SendAlertWebhook(ctx, s, ev); err != nil {
		g.Log().Errorf(ctx, "Failed to send blacklist webhook alert for domain %s: %v", domain, err)
		return
	}
	g.Log().Infof(ctx, "Blacklist webhook alert sent for domain: %s", domain)
}
