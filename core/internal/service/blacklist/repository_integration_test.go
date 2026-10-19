package blacklist

import (
	"billionmail-core/internal/model/entity"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Runs against a real PostgreSQL. Example:
//
//	docker run -d --name bm-pg-test -e POSTGRES_PASSWORD=test -p 55432:5432 postgres:17.4-alpine
//	BM_TEST_PG_LINK='pgsql:postgres:test@tcp(127.0.0.1:55432)/postgres' go test ./internal/service/blacklist/
const pgLinkEnv = "BM_TEST_PG_LINK"

func newTestRepository(t *testing.T) (Repository, gdb.DB) {
	t.Helper()
	if testing.Short() {
		t.Skip("integration test skipped in -short mode")
	}
	link := os.Getenv(pgLinkEnv)
	if link == "" {
		t.Skipf("%s not set", pgLinkEnv)
	}

	db, err := gdb.New(gdb.ConfigNode{Link: link})
	require.NoError(t, err)
	ctx := context.Background()
	_, err = db.Exec(ctx, "DROP TABLE IF EXISTS "+TableName)
	require.NoError(t, err)
	for _, sql := range SchemaSQL {
		_, err = db.Exec(ctx, sql)
		require.NoError(t, err)
	}
	t.Cleanup(func() { _ = db.Close(ctx) })
	return NewRepository(db), db
}

func row(scope Scope, value string, reason Reason, status Status, now int64) *entity.EmailBlacklist {
	domain := value
	if scope == ScopeEmail {
		_, domain, _ = NormalizeEmail(value)
	}
	return &entity.EmailBlacklist{
		Scope: int(scope), Value: value, Domain: domain, Reason: int(reason), Source: int(SourceManual),
		Status: int(status), HitCount: 1, FirstSeenAt: now, LastSeenAt: now, CreateTime: now, UpdateTime: now,
	}
}

func TestRepository_UpsertPreservesHistory(t *testing.T) {
	repo, _ := newTestRepository(t)
	ctx := context.Background()

	first := row(ScopeEmail, "a@example.com", ReasonInvalidMailbox, StatusActive, 1000)
	first.HitCount, first.Dsn, first.LastResponse, first.LastTaskId, first.Note = 4, "5.1.1", "550 user unknown", 7, "from bounce"
	res, err := repo.Upsert(ctx, []*entity.EmailBlacklist{first, row(ScopeDomain, "mailinator.com", ReasonDisposable, StatusActive, 1000)})
	require.NoError(t, err)
	assert.Equal(t, UpsertResult{Inserted: 2}, res)

	// A manual re-add without bounce details must not wipe them.
	again := row(ScopeEmail, "a@example.com", ReasonComplaint, StatusActive, 2000)
	again.ExpiresAt = 9999
	res, err = repo.Upsert(ctx, []*entity.EmailBlacklist{again})
	require.NoError(t, err)
	assert.Equal(t, UpsertResult{Updated: 1}, res)

	got, err := repo.GetByValue(ctx, ScopeEmail, "a@example.com")
	require.NoError(t, err)
	assert.Equal(t, int(ReasonComplaint), got.Reason)
	assert.Equal(t, int64(9999), got.ExpiresAt)
	assert.Equal(t, int64(2000), got.LastSeenAt)
	assert.Equal(t, int64(2000), got.UpdateTime)
	assert.Equal(t, 4, got.HitCount, "hit count preserved")
	assert.Equal(t, int64(1000), got.FirstSeenAt, "first seen preserved")
	assert.Equal(t, int64(1000), got.CreateTime, "create time preserved")
	assert.Equal(t, "5.1.1", got.Dsn)
	assert.Equal(t, "550 user unknown", got.LastResponse)
	assert.Equal(t, 7, got.LastTaskId)
	assert.Equal(t, "from bounce", got.Note)

	// Same value with a different scope is a separate entry.
	_, err = repo.GetByValue(ctx, ScopeDomain, "a@example.com")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestRepository_UpsertLargeBatch(t *testing.T) {
	repo, db := newTestRepository(t)
	ctx := context.Background()

	entries := make([]*entity.EmailBlacklist, 0, 1203)
	for i := 0; i < cap(entries); i++ {
		entries = append(entries, row(ScopeEmail, fmt.Sprintf("bulk%d@example.com", i), ReasonManual, StatusActive, 1000))
	}
	res, err := repo.Upsert(ctx, entries)
	require.NoError(t, err)
	assert.Equal(t, 1203, res.Inserted)

	count, err := db.Model(TableName).Ctx(ctx).Count()
	require.NoError(t, err)
	assert.Equal(t, 1203, count)

	values := make([]string, 0, 1500)
	for i := 0; i < 1500; i++ {
		values = append(values, fmt.Sprintf("bulk%d@example.com", i))
	}
	found, err := repo.FindActive(ctx, ScopeEmail, values, 1000)
	require.NoError(t, err)
	assert.Len(t, found, 1203, "lookups span several IN chunks")
}

func TestRepository_GetNotFound(t *testing.T) {
	repo, _ := newTestRepository(t)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 42)
	assert.ErrorIs(t, err, ErrNotFound)
	_, err = repo.GetByValue(ctx, ScopeEmail, "none@example.com")
	assert.ErrorIs(t, err, ErrNotFound)
	list, err := repo.GetByIDs(ctx, []int64{1, 2})
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestRepository_ListFilters(t *testing.T) {
	repo, _ := newTestRepository(t)
	ctx := context.Background()

	a := row(ScopeEmail, "alice@foo.com", ReasonInvalidMailbox, StatusActive, 100)
	b := row(ScopeEmail, "bob@foo.com", ReasonMailboxFull, StatusObserving, 300)
	c := row(ScopeEmail, "carol@bar.com", ReasonInvalidMailbox, StatusReleased, 200)
	d := row(ScopeDomain, "spam.io", ReasonDisposable, StatusActive, 400)
	d.Source = int(SourceImport)
	_, err := repo.Upsert(ctx, []*entity.EmailBlacklist{a, b, c, d})
	require.NoError(t, err)

	values := func(list []*entity.EmailBlacklist) []string {
		out := make([]string, 0, len(list))
		for _, e := range list {
			out = append(out, e.Value)
		}
		return out
	}
	observing, released := StatusObserving, StatusReleased

	cases := []struct {
		name   string
		filter ListFilter
		total  int
		want   []string
	}{
		{"all ordered by last seen", ListFilter{Page: 1, PageSize: 10}, 4, []string{"spam.io", "bob@foo.com", "carol@bar.com", "alice@foo.com"}},
		{"page 2", ListFilter{Page: 2, PageSize: 3}, 4, []string{"alice@foo.com"}},
		{"keyword", ListFilter{Page: 1, PageSize: 10, Keyword: "foo"}, 2, []string{"bob@foo.com", "alice@foo.com"}},
		{"scope domain", ListFilter{Page: 1, PageSize: 10, Scope: ScopeDomain}, 1, []string{"spam.io"}},
		{"reason", ListFilter{Page: 1, PageSize: 10, Reason: ReasonInvalidMailbox}, 2, []string{"carol@bar.com", "alice@foo.com"}},
		{"status zero is a real filter", ListFilter{Page: 1, PageSize: 10, Status: &observing}, 1, []string{"bob@foo.com"}},
		{"status released", ListFilter{Page: 1, PageSize: 10, Status: &released}, 1, []string{"carol@bar.com"}},
		{"source", ListFilter{Page: 1, PageSize: 10, Source: SourceImport}, 1, []string{"spam.io"}},
		{"domain", ListFilter{Page: 1, PageSize: 10, Domain: "bar.com"}, 1, []string{"carol@bar.com"}},
		{"no match", ListFilter{Page: 1, PageSize: 10, Keyword: "zzz"}, 0, []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			total, list, err := repo.List(ctx, tc.filter)
			require.NoError(t, err)
			assert.Equal(t, tc.total, total)
			assert.Equal(t, tc.want, values(list))
		})
	}
}

func TestRepository_UpdateSetStatusDelete(t *testing.T) {
	repo, _ := newTestRepository(t)
	ctx := context.Background()
	_, err := repo.Upsert(ctx, []*entity.EmailBlacklist{
		row(ScopeEmail, "a@example.com", ReasonManual, StatusActive, 100),
		row(ScopeEmail, "b@example.com", ReasonManual, StatusActive, 100),
	})
	require.NoError(t, err)
	a, _ := repo.GetByValue(ctx, ScopeEmail, "a@example.com")
	b, _ := repo.GetByValue(ctx, ScopeEmail, "b@example.com")

	reason, note, exp := ReasonSpamTrap, "trap", int64(5000)
	require.NoError(t, repo.Update(ctx, a.Id, Patch{Reason: &reason, Note: &note, ExpiresAt: &exp}, 300))
	got, _ := repo.GetByID(ctx, a.Id)
	assert.Equal(t, int(ReasonSpamTrap), got.Reason)
	assert.Equal(t, "trap", got.Note)
	assert.Equal(t, int64(5000), got.ExpiresAt)
	assert.Equal(t, int64(300), got.UpdateTime)
	assert.ErrorIs(t, repo.Update(ctx, 999999, Patch{Note: &note}, 300), ErrNotFound)

	n, err := repo.SetStatus(ctx, []int64{a.Id, b.Id}, StatusReleased, 400)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	n, err = repo.SetStatus(ctx, []int64{a.Id, b.Id}, StatusReleased, 400)
	require.NoError(t, err)
	assert.Zero(t, n, "rows already in the target status are not counted")

	list, err := repo.GetByIDs(ctx, []int64{b.Id, a.Id})
	require.NoError(t, err)
	assert.Len(t, list, 2)

	n, err = repo.Delete(ctx, []int64{a.Id, 999999})
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	_, err = repo.GetByID(ctx, a.Id)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestRepository_FindActive(t *testing.T) {
	repo, _ := newTestRepository(t)
	ctx := context.Background()
	const now = int64(10_000)

	permanent := row(ScopeEmail, "perm@example.com", ReasonInvalidMailbox, StatusActive, 1)
	future := row(ScopeEmail, "future@example.com", ReasonMailboxFull, StatusActive, 1)
	future.ExpiresAt = now + 1
	expired := row(ScopeEmail, "expired@example.com", ReasonMailboxFull, StatusActive, 1)
	expired.ExpiresAt = now
	observing := row(ScopeEmail, "obs@example.com", ReasonMailboxFull, StatusObserving, 1)
	released := row(ScopeEmail, "rel@example.com", ReasonManual, StatusReleased, 1)
	domain := row(ScopeDomain, "example.com", ReasonDisposable, StatusActive, 1)
	_, err := repo.Upsert(ctx, []*entity.EmailBlacklist{permanent, future, expired, observing, released, domain})
	require.NoError(t, err)

	found, err := repo.FindActive(ctx, ScopeEmail, []string{
		"perm@example.com", "future@example.com", "expired@example.com", "obs@example.com", "rel@example.com", "clean@example.com",
	}, now)
	require.NoError(t, err)
	got := make([]string, 0, len(found))
	for _, e := range found {
		got = append(got, e.Value)
	}
	assert.ElementsMatch(t, []string{"perm@example.com", "future@example.com"}, got)

	found, err = repo.FindActive(ctx, ScopeDomain, []string{"example.com"}, now)
	require.NoError(t, err)
	assert.Len(t, found, 1)

	found, err = repo.FindActive(ctx, ScopeEmail, nil, now)
	require.NoError(t, err)
	assert.Empty(t, found)
}

func TestRepository_Stats(t *testing.T) {
	repo, _ := newTestRepository(t)
	ctx := context.Background()

	stats, err := repo.Stats(ctx)
	require.NoError(t, err)
	assert.Empty(t, stats)

	_, err = repo.Upsert(ctx, []*entity.EmailBlacklist{
		row(ScopeEmail, "a@example.com", ReasonInvalidMailbox, StatusActive, 1),
		row(ScopeEmail, "b@example.com", ReasonInvalidMailbox, StatusActive, 1),
		row(ScopeEmail, "c@example.com", ReasonMailboxFull, StatusObserving, 1),
	})
	require.NoError(t, err)

	stats, err = repo.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, []StatCount{
		{Status: StatusObserving, Reason: ReasonMailboxFull, Total: 1},
		{Status: StatusActive, Reason: ReasonInvalidMailbox, Total: 2},
	}, stats)
}

func TestManager_WithPostgres(t *testing.T) {
	repo, _ := newTestRepository(t)
	ctx := context.Background()
	m := NewManager(repo, nil)

	res, err := m.AddBatch(ctx, []AddInput{
		{Scope: ScopeEmail, Value: "X@Example.com", Reason: ReasonComplaint, Source: SourceImport},
		{Scope: ScopeEmail, Value: "x@example.com", Reason: ReasonSpamTrap, Source: SourceImport},
		{Scope: ScopeEmail, Value: "y@example.com", Reason: ReasonMailboxFull, Source: SourceImport},
	})
	require.NoError(t, err, "in-batch duplicates must not trip ON CONFLICT twice")
	assert.Equal(t, 2, res.Inserted)

	x, err := repo.GetByValue(ctx, ScopeEmail, "x@example.com")
	require.NoError(t, err)
	assert.Equal(t, int(ReasonSpamTrap), x.Reason)

	n, err := m.Release(ctx, []int64{x.Id})
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	active, err := repo.FindActive(ctx, ScopeEmail, []string{"x@example.com", "y@example.com"}, time.Now().Unix())
	require.NoError(t, err)
	require.Len(t, active, 1)
	assert.Equal(t, "y@example.com", active[0].Value)
}
