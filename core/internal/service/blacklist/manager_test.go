package blacklist

import (
	"billionmail-core/internal/model/entity"
	"context"
	"errors"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// memRepository is an in-memory Repository that mirrors the PostgreSQL upsert semantics.
type memRepository struct {
	rows       map[int64]*entity.EmailBlacklist
	nextID     int64
	err        error      // returned by Upsert when set
	lastFilter ListFilter // filter received by the last List call
}

func newMemRepository() *memRepository {
	return &memRepository{rows: map[int64]*entity.EmailBlacklist{}}
}

func (r *memRepository) find(scope int, value string) *entity.EmailBlacklist {
	for _, e := range r.rows {
		if e.Scope == scope && e.Value == value {
			return e
		}
	}
	return nil
}

func (r *memRepository) Upsert(_ context.Context, entries []*entity.EmailBlacklist) (UpsertResult, error) {
	var res UpsertResult
	if r.err != nil {
		return res, r.err
	}
	for _, in := range entries {
		if cur := r.find(in.Scope, in.Value); cur != nil {
			cur.Reason, cur.Source, cur.Status, cur.ExpiresAt = in.Reason, in.Source, in.Status, in.ExpiresAt
			if in.Note != "" {
				cur.Note = in.Note
			}
			cur.LastSeenAt, cur.UpdateTime = in.LastSeenAt, in.UpdateTime
			res.Updated++
			continue
		}
		r.nextID++
		row := *in
		row.Id = r.nextID
		r.rows[row.Id] = &row
		res.Inserted++
	}
	return res, nil
}

func (r *memRepository) GetByID(_ context.Context, id int64) (*entity.EmailBlacklist, error) {
	if e, ok := r.rows[id]; ok {
		cp := *e
		return &cp, nil
	}
	return nil, ErrNotFound
}

func (r *memRepository) GetByIDs(_ context.Context, ids []int64) ([]*entity.EmailBlacklist, error) {
	out := make([]*entity.EmailBlacklist, 0)
	for _, id := range ids {
		if e, ok := r.rows[id]; ok {
			cp := *e
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *memRepository) GetByValue(_ context.Context, scope Scope, value string) (*entity.EmailBlacklist, error) {
	if e := r.find(int(scope), value); e != nil {
		cp := *e
		return &cp, nil
	}
	return nil, ErrNotFound
}

func (r *memRepository) List(_ context.Context, f ListFilter) (int, []*entity.EmailBlacklist, error) {
	r.lastFilter = f
	out := make([]*entity.EmailBlacklist, 0)
	for _, e := range r.rows {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Id < out[j].Id })
	return len(out), out, nil
}

func (r *memRepository) Update(_ context.Context, id int64, p Patch, now int64) error {
	e, ok := r.rows[id]
	if !ok {
		return ErrNotFound
	}
	if p.Reason != nil {
		e.Reason = int(*p.Reason)
	}
	if p.Note != nil {
		e.Note = *p.Note
	}
	if p.ExpiresAt != nil {
		e.ExpiresAt = *p.ExpiresAt
	}
	e.UpdateTime = now
	return nil
}

func (r *memRepository) SetStatus(_ context.Context, ids []int64, status Status, now int64) (int, error) {
	n := 0
	for _, id := range ids {
		if e, ok := r.rows[id]; ok && e.Status != int(status) {
			e.Status, e.UpdateTime = int(status), now
			n++
		}
	}
	return n, nil
}

func (r *memRepository) Delete(_ context.Context, ids []int64) (int, error) {
	n := 0
	for _, id := range ids {
		if _, ok := r.rows[id]; ok {
			delete(r.rows, id)
			n++
		}
	}
	return n, nil
}

func (r *memRepository) FindActive(_ context.Context, scope Scope, values []string, now int64) ([]*entity.EmailBlacklist, error) {
	out := make([]*entity.EmailBlacklist, 0)
	for _, v := range values {
		if e := r.find(int(scope), v); e != nil && isBlocking(e, now) {
			out = append(out, e)
		}
	}
	return out, nil
}

func (r *memRepository) Stats(context.Context) ([]StatCount, error) { return nil, nil }

// recordingCache records the values passed to Put and Remove.
type recordingCache struct {
	put, removed []string
	err          error
}

func (c *recordingCache) Put(_ context.Context, entries []*entity.EmailBlacklist) error {
	for _, e := range entries {
		c.put = append(c.put, e.Value)
	}
	return c.err
}

func (c *recordingCache) Remove(_ context.Context, entries []*entity.EmailBlacklist) error {
	for _, e := range entries {
		c.removed = append(c.removed, e.Value)
	}
	return c.err
}

var fixedNow = time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)

func newTestManager() (*Manager, *memRepository, *recordingCache) {
	repo, cache := newMemRepository(), &recordingCache{}
	m := NewManager(repo, cache)
	m.now = func() time.Time { return fixedNow }
	return m, repo, cache
}

func ttl(d time.Duration) *time.Duration { return &d }

func TestManager_AddNormalizesAndAppliesPolicy(t *testing.T) {
	m, _, cache := newTestManager()
	ctx := context.Background()

	e, err := m.Add(ctx, AddInput{Scope: ScopeEmail, Value: " Bob@Example.com ", Reason: ReasonInvalidMailbox})
	require.NoError(t, err)
	assert.Equal(t, "bob@example.com", e.Value)
	assert.Equal(t, "example.com", e.Domain)
	assert.Equal(t, int(StatusActive), e.Status)
	assert.Equal(t, int(SourceManual), e.Source, "source defaults to manual")
	assert.Zero(t, e.ExpiresAt, "hard bounce is permanent")

	full, err := m.Add(ctx, AddInput{Scope: ScopeEmail, Value: "full@example.com", Reason: ReasonMailboxFull})
	require.NoError(t, err)
	assert.Equal(t, fixedNow.Add(7*24*time.Hour).Unix(), full.ExpiresAt, "soft reason uses policy TTL")

	perm, err := m.Add(ctx, AddInput{Scope: ScopeEmail, Value: "p@example.com", Reason: ReasonMailboxFull, TTL: ttl(0)})
	require.NoError(t, err)
	assert.Zero(t, perm.ExpiresAt, "explicit zero TTL overrides policy")

	dom, err := m.Add(ctx, AddInput{Scope: ScopeDomain, Value: "@Mailinator.com", Reason: ReasonDisposable, Source: SourceImport})
	require.NoError(t, err)
	assert.Equal(t, "mailinator.com", dom.Value)
	assert.Equal(t, "mailinator.com", dom.Domain)

	assert.Equal(t, []string{"bob@example.com", "full@example.com", "p@example.com", "mailinator.com"}, cache.put)
}

func TestManager_AddRejectsInvalidInput(t *testing.T) {
	m, _, _ := newTestManager()
	ctx := context.Background()

	cases := map[string]struct {
		in   AddInput
		want error
	}{
		"bad email":  {AddInput{Scope: ScopeEmail, Value: "nope", Reason: ReasonManual}, ErrInvalidEmail},
		"bad domain": {AddInput{Scope: ScopeDomain, Value: "nope", Reason: ReasonManual}, ErrInvalidDomain},
		"bad scope":  {AddInput{Scope: 9, Value: "a@b.com", Reason: ReasonManual}, ErrInvalidScope},
		"bad reason": {AddInput{Scope: ScopeEmail, Value: "a@b.com", Reason: 99}, ErrInvalidReason},
		"bad source": {AddInput{Scope: ScopeEmail, Value: "a@b.com", Reason: ReasonManual, Source: 99}, ErrInvalidSource},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			res, err := m.AddBatch(ctx, []AddInput{tc.in})
			require.NoError(t, err)
			require.Len(t, res.Invalid, 1)
			assert.Contains(t, res.Invalid[0].Error, tc.want.Error())

			_, err = m.Add(ctx, tc.in)
			assert.ErrorIs(t, err, tc.want, "Add keeps the sentinel error")
		})
	}

	_, err := m.Add(ctx, AddInput{Scope: ScopeEmail, Value: "a@b.com", Reason: ReasonManual, TTL: ttl(-time.Hour)})
	assert.Error(t, err)
}

func TestManager_AddBatch(t *testing.T) {
	m, repo, cache := newTestManager()
	ctx := context.Background()

	res, err := m.AddBatch(ctx, []AddInput{
		{Scope: ScopeEmail, Value: "a@example.com", Reason: ReasonInvalidMailbox, Note: "first"},
		{Scope: ScopeEmail, Value: "bad-address", Reason: ReasonManual},
		{Scope: ScopeEmail, Value: "A@EXAMPLE.com", Reason: ReasonComplaint, Note: "last wins"},
		{Scope: ScopeEmail, Value: "b@example.com", Reason: ReasonSpamTrap},
	})
	require.NoError(t, err)
	assert.Equal(t, 2, res.Inserted, "duplicates collapse before reaching the database")
	require.Len(t, res.Invalid, 1)
	assert.Equal(t, 1, res.Invalid[0].Index)

	a, err := repo.GetByValue(ctx, ScopeEmail, "a@example.com")
	require.NoError(t, err)
	assert.Equal(t, int(ReasonComplaint), a.Reason)
	assert.Equal(t, "last wins", a.Note)
	assert.ElementsMatch(t, []string{"a@example.com", "b@example.com"}, cache.put)

	// Re-adding a released entry reactivates it.
	_, err = m.Release(ctx, []int64{a.Id})
	require.NoError(t, err)
	res, err = m.AddBatch(ctx, []AddInput{{Scope: ScopeEmail, Value: "a@example.com", Reason: ReasonManual}})
	require.NoError(t, err)
	assert.Equal(t, 1, res.Updated)
	a, _ = repo.GetByValue(ctx, ScopeEmail, "a@example.com")
	assert.Equal(t, int(StatusActive), a.Status)
}

func TestManager_AddBatchAllInvalidSkipsDatabase(t *testing.T) {
	m, repo, _ := newTestManager()
	repo.err = errors.New("must not be called")

	res, err := m.AddBatch(context.Background(), []AddInput{{Scope: ScopeEmail, Value: "x", Reason: ReasonManual}})
	require.NoError(t, err)
	assert.Len(t, res.Invalid, 1)
}

func TestManager_NoteIsTruncated(t *testing.T) {
	m, _, _ := newTestManager()
	long := ""
	for i := 0; i < 300; i++ {
		long += "é"
	}
	e, err := m.Add(context.Background(), AddInput{Scope: ScopeEmail, Value: "n@example.com", Reason: ReasonManual, Note: long})
	require.NoError(t, err)
	assert.Equal(t, maxNoteLength, len([]rune(e.Note)))
}

func TestManager_Update(t *testing.T) {
	m, _, cache := newTestManager()
	ctx := context.Background()
	e, err := m.Add(ctx, AddInput{Scope: ScopeEmail, Value: "u@example.com", Reason: ReasonManual})
	require.NoError(t, err)
	cache.put = nil

	reason, note := ReasonSpamTrap, "confirmed trap"
	future := fixedNow.Add(time.Hour).Unix()
	updated, err := m.Update(ctx, e.Id, UpdateInput{Reason: &reason, Note: &note, ExpiresAt: &future})
	require.NoError(t, err)
	assert.Equal(t, int(ReasonSpamTrap), updated.Reason)
	assert.Equal(t, "confirmed trap", updated.Note)
	assert.Equal(t, future, updated.ExpiresAt)
	assert.Equal(t, []string{"u@example.com"}, cache.put, "active entry is refreshed in cache")

	past := fixedNow.Add(-time.Hour).Unix()
	_, err = m.Update(ctx, e.Id, UpdateInput{ExpiresAt: &past})
	assert.Error(t, err)

	bad := Reason(99)
	_, err = m.Update(ctx, e.Id, UpdateInput{Reason: &bad})
	assert.ErrorIs(t, err, ErrInvalidReason)

	_, err = m.Update(ctx, 12345, UpdateInput{Note: &note})
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestManager_UpdateReleasedEntryRemovesFromCache(t *testing.T) {
	m, _, cache := newTestManager()
	ctx := context.Background()
	e, err := m.Add(ctx, AddInput{Scope: ScopeEmail, Value: "r@example.com", Reason: ReasonManual})
	require.NoError(t, err)
	_, err = m.Release(ctx, []int64{e.Id})
	require.NoError(t, err)
	cache.put, cache.removed = nil, nil

	note := "still released"
	_, err = m.Update(ctx, e.Id, UpdateInput{Note: &note})
	require.NoError(t, err)
	assert.Empty(t, cache.put)
	assert.Equal(t, []string{"r@example.com"}, cache.removed)
}

func TestManager_ReleaseAndDelete(t *testing.T) {
	m, repo, cache := newTestManager()
	ctx := context.Background()
	a, _ := m.Add(ctx, AddInput{Scope: ScopeEmail, Value: "a@example.com", Reason: ReasonManual})
	b, _ := m.Add(ctx, AddInput{Scope: ScopeEmail, Value: "b@example.com", Reason: ReasonManual})

	n, err := m.Release(ctx, []int64{a.Id, 999})
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, []string{"a@example.com"}, cache.removed)

	n, err = m.Release(ctx, []int64{a.Id})
	require.NoError(t, err)
	assert.Zero(t, n, "releasing twice changes nothing")

	n, err = m.Delete(ctx, []int64{b.Id})
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	_, err = repo.GetByID(ctx, b.Id)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestManager_CacheFailureKeepsDatabaseChange(t *testing.T) {
	m, repo, cache := newTestManager()
	cache.err = errors.New("redis down")

	e, err := m.Add(context.Background(), AddInput{Scope: ScopeEmail, Value: "c@example.com", Reason: ReasonManual})
	assert.ErrorIs(t, err, ErrCacheSync)
	require.NotNil(t, e)
	_, getErr := repo.GetByValue(context.Background(), ScopeEmail, "c@example.com")
	assert.NoError(t, getErr)
}

func TestManager_ListValidatesAndClampsPaging(t *testing.T) {
	m, repo, _ := newTestManager()
	ctx := context.Background()

	_, _, err := m.List(ctx, ListFilter{Reason: 99})
	assert.ErrorIs(t, err, ErrInvalidReason)
	bad := Status(7)
	_, _, err = m.List(ctx, ListFilter{Status: &bad})
	assert.ErrorIs(t, err, ErrInvalidStatus)

	for i := 0; i < 3; i++ {
		_, err := m.Add(ctx, AddInput{Scope: ScopeEmail, Value: fmt.Sprintf("l%d@example.com", i), Reason: ReasonManual})
		require.NoError(t, err)
	}
	total, list, err := m.List(ctx, ListFilter{PageSize: 10_000})
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, list, 3)
	assert.Equal(t, 1, repo.lastFilter.Page)
	assert.Equal(t, maxPageSize, repo.lastFilter.PageSize)

	_, _, err = m.List(ctx, ListFilter{})
	require.NoError(t, err)
	assert.Equal(t, defaultPageSize, repo.lastFilter.PageSize)
}
