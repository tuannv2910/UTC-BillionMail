package blacklist

import (
	"billionmail-core/internal/model/entity"
	"context"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"
)

const (
	defaultPageSize = 20
	maxPageSize     = 500
	maxNoteLength   = 255
)

// Cache mirrors active entries for fast lookups at send time. The database stays the source of truth.
type Cache interface {
	// Put stores entries that currently block sending.
	Put(ctx context.Context, entries []*entity.EmailBlacklist) error
	// Remove drops entries so they no longer block sending.
	Remove(ctx context.Context, entries []*entity.EmailBlacklist) error
}

// NoopCache is a Cache that does nothing, used until a real cache is configured.
type NoopCache struct{}

func (NoopCache) Put(context.Context, []*entity.EmailBlacklist) error    { return nil }
func (NoopCache) Remove(context.Context, []*entity.EmailBlacklist) error { return nil }

// Manager implements the suppression list use cases on top of a Repository and a Cache.
// Writes go to the repository first and are then mirrored to the cache; a cache failure is
// reported as ErrCacheSync while the database change is kept.
type Manager struct {
	repo  Repository
	cache Cache
	now   func() time.Time
}

// NewManager returns a Manager. A nil cache disables caching.
func NewManager(repo Repository, cache Cache) *Manager {
	if cache == nil {
		cache = NoopCache{}
	}
	return &Manager{repo: repo, cache: cache, now: time.Now}
}

// AddInput describes one entry to block.
type AddInput struct {
	Scope  Scope
	Value  string // email or domain, normalized by the manager
	Reason Reason
	Source Source         // 0 defaults to SourceManual
	Note   string         // truncated to 255 characters
	TTL    *time.Duration // nil uses the reason's policy; 0 blocks permanently
}

// InvalidItem reports an input rejected by AddBatch.
type InvalidItem struct {
	Index int    `json:"index"`
	Value string `json:"value"`
	Error string `json:"error"`
}

// AddResult summarizes an AddBatch call.
type AddResult struct {
	Inserted int           `json:"inserted"`
	Updated  int           `json:"updated"`
	Invalid  []InvalidItem `json:"invalid"`
}

// Add blocks a single email or domain and returns the stored entry.
func (m *Manager) Add(ctx context.Context, in AddInput) (*entity.EmailBlacklist, error) {
	// Validate here so callers get the sentinel error (ErrInvalidEmail ...), not just its text.
	if _, err := m.buildEntry(in, m.now()); err != nil {
		return nil, err
	}
	_, err := m.AddBatch(ctx, []AddInput{in})
	if err != nil && !errors.Is(err, ErrCacheSync) {
		return nil, err
	}

	value, _, _ := normalizeValue(in.Scope, in.Value)
	entry, getErr := m.repo.GetByValue(ctx, in.Scope, value)
	if getErr != nil {
		return nil, getErr
	}
	return entry, err
}

// AddBatch blocks many entries at once, e.g. from a CSV import. Invalid inputs are skipped and
// reported; duplicates within the batch collapse to the last occurrence. Existing entries in any
// status become active with the new reason and expiry.
func (m *Manager) AddBatch(ctx context.Context, inputs []AddInput) (AddResult, error) {
	res := AddResult{Invalid: make([]InvalidItem, 0)}
	now := m.now()
	nowUnix := now.Unix()

	entries := make([]*entity.EmailBlacklist, 0, len(inputs))
	position := make(map[string]int, len(inputs))
	for i, in := range inputs {
		e, err := m.buildEntry(in, now)
		if err != nil {
			res.Invalid = append(res.Invalid, InvalidItem{Index: i, Value: in.Value, Error: err.Error()})
			continue
		}
		key := fmt.Sprintf("%d|%s", e.Scope, e.Value)
		if p, dup := position[key]; dup {
			entries[p] = e
			continue
		}
		position[key] = len(entries)
		entries = append(entries, e)
	}
	if len(entries) == 0 {
		return res, nil
	}

	upserted, err := m.repo.Upsert(ctx, entries)
	if err != nil {
		return res, err
	}
	res.Inserted, res.Updated = upserted.Inserted, upserted.Updated

	if err := m.cache.Put(ctx, activeAt(entries, nowUnix)); err != nil {
		return res, errors.Join(ErrCacheSync, err)
	}
	return res, nil
}

func (m *Manager) buildEntry(in AddInput, now time.Time) (*entity.EmailBlacklist, error) {
	if !in.Scope.Valid() {
		return nil, ErrInvalidScope
	}
	if !in.Reason.Valid() {
		return nil, ErrInvalidReason
	}
	source := in.Source
	if source == 0 {
		source = SourceManual
	}
	if !source.Valid() {
		return nil, ErrInvalidSource
	}
	value, domain, err := normalizeValue(in.Scope, in.Value)
	if err != nil {
		return nil, err
	}

	ttl := in.Reason.Policy().TTL
	if in.TTL != nil {
		ttl = *in.TTL
	}
	if ttl < 0 {
		return nil, fmt.Errorf("blacklist: ttl must not be negative")
	}
	var expiresAt int64
	if ttl > 0 {
		expiresAt = now.Add(ttl).Unix()
	}

	nowUnix := now.Unix()
	return &entity.EmailBlacklist{
		Scope:       int(in.Scope),
		Value:       value,
		Domain:      domain,
		Reason:      int(in.Reason),
		Source:      int(source),
		Status:      int(StatusActive),
		HitCount:    1,
		ExpiresAt:   expiresAt,
		Note:        truncate(in.Note, maxNoteLength),
		FirstSeenAt: nowUnix,
		LastSeenAt:  nowUnix,
		CreateTime:  nowUnix,
		UpdateTime:  nowUnix,
	}, nil
}

// Get returns one entry by id.
func (m *Manager) Get(ctx context.Context, id int64) (*entity.EmailBlacklist, error) {
	return m.repo.GetByID(ctx, id)
}

// List returns a page of entries. Page defaults to 1 and page size to 20 (max 500).
func (m *Manager) List(ctx context.Context, f ListFilter) (int, []*entity.EmailBlacklist, error) {
	if f.Scope != 0 && !f.Scope.Valid() {
		return 0, nil, ErrInvalidScope
	}
	if f.Reason != 0 && !f.Reason.Valid() {
		return 0, nil, ErrInvalidReason
	}
	if f.Source != 0 && !f.Source.Valid() {
		return 0, nil, ErrInvalidSource
	}
	if f.Status != nil && !f.Status.Valid() {
		return 0, nil, ErrInvalidStatus
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.PageSize <= 0 {
		f.PageSize = defaultPageSize
	}
	f.PageSize = min(f.PageSize, maxPageSize)
	return m.repo.List(ctx, f)
}

// UpdateInput holds the editable fields; nil fields are left unchanged.
type UpdateInput struct {
	Reason    *Reason
	Note      *string
	ExpiresAt *int64 // unix time; 0 makes the block permanent
}

// Update edits an entry and refreshes the cache to match its new state.
func (m *Manager) Update(ctx context.Context, id int64, in UpdateInput) (*entity.EmailBlacklist, error) {
	now := m.now().Unix()
	if in.Reason != nil && !in.Reason.Valid() {
		return nil, ErrInvalidReason
	}
	if in.ExpiresAt != nil && *in.ExpiresAt != 0 && *in.ExpiresAt <= now {
		return nil, fmt.Errorf("blacklist: expires_at must be 0 or in the future")
	}
	patch := Patch{Reason: in.Reason, ExpiresAt: in.ExpiresAt}
	if in.Note != nil {
		note := truncate(*in.Note, maxNoteLength)
		patch.Note = &note
	}

	if err := m.repo.Update(ctx, id, patch, now); err != nil {
		return nil, err
	}
	entry, err := m.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if isBlocking(entry, now) {
		err = m.cache.Put(ctx, []*entity.EmailBlacklist{entry})
	} else {
		err = m.cache.Remove(ctx, []*entity.EmailBlacklist{entry})
	}
	if err != nil {
		return entry, errors.Join(ErrCacheSync, err)
	}
	return entry, nil
}

// Release unblocks entries but keeps them for history. Returns how many changed.
func (m *Manager) Release(ctx context.Context, ids []int64) (int, error) {
	entries, err := m.repo.GetByIDs(ctx, ids)
	if err != nil {
		return 0, err
	}
	n, err := m.repo.SetStatus(ctx, ids, StatusReleased, m.now().Unix())
	if err != nil {
		return 0, err
	}
	if err := m.cache.Remove(ctx, entries); err != nil {
		return n, errors.Join(ErrCacheSync, err)
	}
	return n, nil
}

// Delete removes entries permanently. Prefer Release to keep history.
func (m *Manager) Delete(ctx context.Context, ids []int64) (int, error) {
	entries, err := m.repo.GetByIDs(ctx, ids)
	if err != nil {
		return 0, err
	}
	n, err := m.repo.Delete(ctx, ids)
	if err != nil {
		return 0, err
	}
	if err := m.cache.Remove(ctx, entries); err != nil {
		return n, errors.Join(ErrCacheSync, err)
	}
	return n, nil
}

// Stats returns entry counts grouped by status and reason.
func (m *Manager) Stats(ctx context.Context) ([]StatCount, error) {
	return m.repo.Stats(ctx)
}

// isBlocking reports whether e blocks sending at unix time now.
func isBlocking(e *entity.EmailBlacklist, now int64) bool {
	return Status(e.Status) == StatusActive && (e.ExpiresAt == 0 || e.ExpiresAt > now)
}

func activeAt(entries []*entity.EmailBlacklist, now int64) []*entity.EmailBlacklist {
	active := make([]*entity.EmailBlacklist, 0, len(entries))
	for _, e := range entries {
		if isBlocking(e, now) {
			active = append(active, e)
		}
	}
	return active
}

func truncate(s string, maxRunes int) string {
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	return string([]rune(s)[:maxRunes])
}
