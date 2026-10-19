package inbox_placement

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestClient(t *testing.T, admin string, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return NewClient(srv.URL+"/", "key-1", admin, 5*time.Second)
}

func TestDisabled(t *testing.T) {
	c := NewClient("", "", "", time.Second)
	if c.Enabled() || c.AdminEnabled() {
		t.Fatal("empty URL must disable the client")
	}
	if _, err := c.GetTest(context.Background(), "x"); !errors.Is(err, ErrDisabled) {
		t.Fatalf("got %v want ErrDisabled", err)
	}
}

func TestCreateTest(t *testing.T) {
	c := newTestClient(t, "", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/placement-tests" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer key-1" {
			t.Errorf("authorization=%q", r.Header.Get("Authorization"))
		}
		var in CreateInput
		_ = json.NewDecoder(r.Body).Decode(&in)
		if in.Reference != "template:7" || in.Domain != "abc.com" {
			t.Errorf("input=%+v", in)
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"test_id":"t1","status":"pending","seeds":3,"poll_after_sec":20}`))
	})
	out, err := c.CreateTest(context.Background(), CreateInput{Domain: "abc.com", Reference: "template:7"})
	if err != nil || out.TestID != "t1" || out.Seeds != 3 {
		t.Fatalf("out=%+v err=%v", out, err)
	}
}

func TestAPIError(t *testing.T) {
	c := newTestClient(t, "", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error":{"code":"no_active_seed","message":"no active seed in this group"}}`))
	})
	_, err := c.CreateTest(context.Background(), CreateInput{})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "no_active_seed" || apiErr.Status != 409 || err.Error() != "no active seed in this group" {
		t.Fatalf("err=%v", err)
	}
}

func TestListTestsConvertsTimes(t *testing.T) {
	c := newTestClient(t, "", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("reference") != "template:12" || q.Get("page") != "2" || q.Get("page_size") != "10" {
			t.Errorf("query=%v", q)
		}
		_, _ = w.Write([]byte(`{"total":11,"items":[{"test_id":"t1","status":"completed","reference":"template:12",
			"created_at":"2026-10-16T08:00:00Z","sent_at":null,"completed_at":"2026-10-16T08:05:00Z",
			"summary":{"total":1,"inbox":1,"delivered":1,"inbox_rate":1},
			"results":[{"seed":"seed_01","provider":"gmail","status":"inbox","received_at":"2026-10-16T08:00:09Z","latency_ms":7000,"spf":"pass"}]}]}`))
	})
	list, total, err := c.ListTests(context.Background(), ListFilter{Page: 2, PageSize: 10, TemplateID: 12})
	if err != nil || total != 11 || len(list) != 1 {
		t.Fatalf("list=%v total=%d err=%v", list, total, err)
	}
	tt := list[0]
	if tt.TemplateId != 12 || tt.CreatedAt != 1792137600 || tt.SentAt != 0 || tt.CompletedAt != 1792137900 {
		t.Fatalf("test=%+v", tt)
	}
	if tt.Summary.InboxRate != 1 || tt.Results[0].ReceivedAt != 1792137609 || *tt.Results[0].LatencyMS != 7000 {
		t.Fatalf("results=%+v", tt.Results)
	}
}

func TestAdminCallsUseAdminToken(t *testing.T) {
	if _, err := NewClient("http://x", "k", "", time.Second).ListSeeds(context.Background()); !errors.Is(err, ErrAdminDisabled) {
		t.Fatalf("got %v want ErrAdminDisabled", err)
	}
	c := newTestClient(t, "adm", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Admin-Token") != "adm" || r.Header.Get("Authorization") != "" {
			t.Errorf("headers=%v", r.Header)
		}
		switch r.URL.Path {
		case "/v1/admin/seeds":
			_, _ = w.Write([]byte(`{"items":[{"id":1,"email":"s1@gmail.com","is_active":true,"health_status":"ok","created_at":"2026-10-16T08:00:00Z"}]}`))
		case "/v1/admin/seeds/1":
			b, _ := io.ReadAll(r.Body)
			if string(b) != `{"is_active":false}` {
				t.Errorf("patch body=%s", b)
			}
		}
	})
	seeds, err := c.ListSeeds(context.Background())
	if err != nil || len(seeds) != 1 || seeds[0].Email != "s1@gmail.com" || seeds[0].CreatedAt != 1792137600 {
		t.Fatalf("seeds=%v err=%v", seeds, err)
	}
	off := false
	if err := c.UpdateSeed(context.Background(), 1, SeedPatch{IsActive: &off}); err != nil {
		t.Fatal(err)
	}
}

func TestTemplateIDFromReference(t *testing.T) {
	for ref, want := range map[string]int{"template:5": 5, "template:x": 0, "other:5": 0, "": 0} {
		if got := TemplateIDFromReference(ref); got != want {
			t.Errorf("%q -> %d want %d", ref, got, want)
		}
	}
}
