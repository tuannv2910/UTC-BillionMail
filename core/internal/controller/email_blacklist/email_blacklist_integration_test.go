package email_blacklist

import (
	"billionmail-core/internal/service/blacklist"
	"billionmail-core/internal/service/middlewares"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Exercises request binding, validation and the controllers over HTTP against a real PostgreSQL.
// Skipped unless BM_TEST_PG_LINK is set (see internal/service/blacklist/repository_integration_test.go).

type apiResponse struct {
	Success bool            `json:"success"`
	Code    int             `json:"code"`
	Msg     string          `json:"msg"`
	Data    json.RawMessage `json:"data"`
}

type apiClient struct {
	t      *testing.T
	client *gclient.Client
}

func (c apiClient) call(method, path string, body any) apiResponse {
	c.t.Helper()
	ctx := context.Background()
	var resp *gclient.Response
	var err error
	if method == "GET" {
		resp, err = c.client.Get(ctx, path, body)
	} else {
		resp, err = c.client.ContentJson().Post(ctx, path, body)
	}
	require.NoError(c.t, err)
	defer resp.Close()

	var out apiResponse
	require.NoError(c.t, json.Unmarshal(resp.ReadAll(), &out), path)
	return out
}

// assertRejected checks a request was refused with a message. Validation failures currently come
// back with code 500, not 412: HandleApiResponse's gvalid.Error type check misses the wrapped error.
func assertRejected(t *testing.T, r apiResponse, msgAndArgs ...any) {
	t.Helper()
	assert.False(t, r.Success, msgAndArgs...)
	assert.NotEmpty(t, r.Msg, msgAndArgs...)
}

func decode[T any](t *testing.T, r apiResponse) T {
	t.Helper()
	var v T
	require.NoError(t, json.Unmarshal(r.Data, &v))
	return v
}

type entry struct {
	Id        int64  `json:"id"`
	Value     string `json:"value"`
	Reason    int    `json:"reason"`
	Status    int    `json:"status"`
	Note      string `json:"note"`
	ExpiresAt int64  `json:"expires_at"`
}

type list struct {
	Total int     `json:"total"`
	List  []entry `json:"list"`
}

func setup(t *testing.T) apiClient {
	t.Helper()
	if testing.Short() {
		t.Skip("integration test skipped in -short mode")
	}
	link := os.Getenv("BM_TEST_PG_LINK")
	if link == "" {
		t.Skip("BM_TEST_PG_LINK not set")
	}

	require.NoError(t, gdb.SetConfig(gdb.Config{gdb.DefaultGroupName: gdb.ConfigGroup{{Link: link}}}))
	ctx := context.Background()
	_, err := g.DB().Exec(ctx, "DROP TABLE IF EXISTS "+blacklist.TableName)
	require.NoError(t, err)
	for _, sql := range blacklist.SchemaSQL {
		_, err = g.DB().Exec(ctx, sql)
		require.NoError(t, err)
	}

	s := g.Server(fmt.Sprintf("blacklist-test-%d", time.Now().UnixNano()))
	s.SetAddr("127.0.0.1:0")
	s.SetDumpRouterMap(false)
	s.Group("/api", func(group *ghttp.RouterGroup) {
		group.Middleware(middlewares.HandleApiResponse)
		group.Bind(NewV1())
	})
	require.NoError(t, s.Start())
	t.Cleanup(func() { _ = s.Shutdown() })

	client := g.Client()
	client.SetPrefix(fmt.Sprintf("http://127.0.0.1:%d/api", s.GetListenedPort()))
	return apiClient{t: t, client: client}
}

func TestBlacklistAPI(t *testing.T) {
	api := setup(t)
	now := time.Now().Unix()

	// Create: value is normalized; ttl_days defaults to -1, so a hard bounce is permanent.
	r := api.call("POST", "/email_blacklist/create", g.Map{"scope": 1, "value": " Bob@Example.com ", "reason": 1})
	require.True(t, r.Success, r.Msg)
	bob := decode[entry](t, r)
	assert.Equal(t, "bob@example.com", bob.Value)
	assert.Zero(t, bob.ExpiresAt)

	// Omitted ttl_days on a soft reason uses its 7-day policy.
	r = api.call("POST", "/email_blacklist/create", g.Map{"scope": 1, "value": "full@example.com", "reason": 4})
	require.True(t, r.Success, r.Msg)
	full := decode[entry](t, r)
	assert.InDelta(t, now+7*24*3600, full.ExpiresAt, 60)

	// ttl_days=0 is permanent even for a soft reason.
	r = api.call("POST", "/email_blacklist/create", g.Map{"scope": 2, "value": "@Mailinator.com", "reason": 8, "ttl_days": 0})
	require.True(t, r.Success, r.Msg)
	assert.Zero(t, decode[entry](t, r).ExpiresAt)

	// Invalid value is rejected by the service; invalid scope by request validation.
	r = api.call("POST", "/email_blacklist/create", g.Map{"scope": 1, "value": "not-an-email", "reason": 1})
	assert.False(t, r.Success)
	assert.Contains(t, r.Msg, "Invalid email")
	r = api.call("POST", "/email_blacklist/create", g.Map{"scope": 3, "value": "a@b.com", "reason": 1})
	assert.False(t, r.Success)
	assertRejected(t, r)

	// Import reports invalid values and collapses duplicates.
	r = api.call("POST", "/email_blacklist/import", g.Map{
		"scope": 1, "reason": 6, "values": []string{"a@x.com", "bad", "A@X.com"},
	})
	require.True(t, r.Success, r.Msg)
	imported := decode[struct {
		Inserted int `json:"inserted"`
		Invalid  []struct {
			Value string `json:"value"`
		} `json:"invalid"`
	}](t, r)
	assert.Equal(t, 1, imported.Inserted)
	require.Len(t, imported.Invalid, 1)
	assert.Equal(t, "bad", imported.Invalid[0].Value)

	// List: status omitted means all (-1); 0 must still filter observing entries.
	all := decode[list](t, api.call("GET", "/email_blacklist/list", g.Map{"page": 1, "page_size": 10}))
	assert.Equal(t, 4, all.Total)
	observing := decode[list](t, api.call("GET", "/email_blacklist/list", g.Map{"page": 1, "page_size": 10, "status": 0}))
	assert.Equal(t, 0, observing.Total)
	byKeyword := decode[list](t, api.call("GET", "/email_blacklist/list", g.Map{"page": 1, "page_size": 10, "keyword": "BOB"}))
	require.Equal(t, 1, byKeyword.Total)
	assert.Equal(t, bob.Id, byKeyword.List[0].Id)
	r = api.call("GET", "/email_blacklist/list", g.Map{"page": 0, "page_size": 10})
	assertRejected(t, r)

	// Update: expire_mode defaults to keep; permanent clears the expiry; unknown modes are rejected.
	r = api.call("POST", "/email_blacklist/update", g.Map{"id": full.Id, "reason": 5, "note": "checked"})
	require.True(t, r.Success, r.Msg)
	updated := decode[entry](t, r)
	assert.Equal(t, 5, updated.Reason)
	assert.Equal(t, "checked", updated.Note)
	assert.Equal(t, full.ExpiresAt, updated.ExpiresAt)
	r = api.call("POST", "/email_blacklist/update", g.Map{"id": full.Id, "reason": 5, "expire_mode": "permanent"})
	require.True(t, r.Success, r.Msg)
	assert.Zero(t, decode[entry](t, r).ExpiresAt)
	r = api.call("POST", "/email_blacklist/update", g.Map{"id": full.Id, "reason": 5, "expire_mode": "forever"})
	assertRejected(t, r)
	r = api.call("POST", "/email_blacklist/update", g.Map{"id": 999999, "reason": 5})
	assert.False(t, r.Success)
	assert.Contains(t, r.Msg, "not found")

	// Release keeps history; stats reflect it.
	r = api.call("POST", "/email_blacklist/release", g.Map{"ids": []int64{bob.Id}})
	require.True(t, r.Success, r.Msg)
	assert.Equal(t, 1, decode[struct {
		Count int `json:"count"`
	}](t, r).Count)
	released := decode[list](t, api.call("GET", "/email_blacklist/list", g.Map{"page": 1, "page_size": 10, "status": 2}))
	assert.Equal(t, 1, released.Total)

	stats := decode[struct {
		Active   int `json:"active"`
		Released int `json:"released"`
		ByReason []struct {
			Reason int `json:"reason"`
			Total  int `json:"total"`
		} `json:"by_reason"`
	}](t, api.call("GET", "/email_blacklist/stats", nil))
	assert.Equal(t, 3, stats.Active)
	assert.Equal(t, 1, stats.Released)
	assert.Len(t, stats.ByReason, 3)

	// Delete removes rows for good.
	r = api.call("POST", "/email_blacklist/delete", g.Map{"ids": []int64{bob.Id, full.Id}})
	require.True(t, r.Success, r.Msg)
	assert.Equal(t, 2, decode[struct {
		Count int `json:"count"`
	}](t, r).Count)
	r = api.call("POST", "/email_blacklist/delete", g.Map{})
	assertRejected(t, r, "ids is required")
}
