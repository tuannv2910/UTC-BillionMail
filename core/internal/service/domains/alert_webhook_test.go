package domains

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateAlertWebhook(t *testing.T) {
	cases := []struct {
		s  AlertWebhookSettings
		ok bool
	}{
		{AlertWebhookSettings{}, true},
		{AlertWebhookSettings{Enabled: true, URL: "https://open.feishu.cn/hook/x", Format: "feishu"}, true},
		{AlertWebhookSettings{Enabled: true, URL: "ftp://x"}, false},
		{AlertWebhookSettings{Enabled: true, URL: "not a url"}, false},
		{AlertWebhookSettings{Enabled: true, URL: "https://x.com", Format: "teams"}, false},
	}
	for _, c := range cases {
		if err := ValidateAlertWebhook(c.s); (err == nil) != c.ok {
			t.Errorf("ValidateAlertWebhook(%+v) err=%v, want ok=%v", c.s, err, c.ok)
		}
	}
}

func TestBuildWebhookPayload_Formats(t *testing.T) {
	ev := AlertEvent{Event: "blacklist", Title: "T", Message: "M", Domain: "example.com"}

	check := func(format, path string) {
		b, err := BuildWebhookPayload(format, ev)
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]interface{}
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatal(err)
		}
		if _, ok := m[path]; !ok {
			t.Errorf("%s payload missing %q: %s", format, path, b)
		}
	}
	check(WebhookFormatJSON, "domain")
	check(WebhookFormatFeishu, "msg_type")
	check(WebhookFormatWeCom, "msgtype")
	check(WebhookFormatDingTalk, "msgtype")
	check(WebhookFormatSlack, "text")
	check(WebhookFormatGotify, "priority")
}

func TestSendAlertWebhook(t *testing.T) {
	var got map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		if r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	err := SendAlertWebhook(context.Background(), AlertWebhookSettings{Enabled: true, URL: srv.URL, Format: "slack"}, AlertEvent{Title: "Hi", Message: "there"})
	if err != nil {
		t.Fatal(err)
	}
	if got["text"] != "Hi\nthere" {
		t.Errorf("unexpected body %v", got)
	}

	failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer failing.Close()
	if err := SendAlertWebhook(context.Background(), AlertWebhookSettings{Enabled: true, URL: failing.URL}, AlertEvent{Title: "x"}); err == nil {
		t.Error("expected error on HTTP 500")
	}
}
