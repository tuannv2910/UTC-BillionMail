package blacklist

import (
	"billionmail-core/internal/model/entity"
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
)

const (
	// upsertChunkSize keeps each INSERT well under PostgreSQL's 65535 bind-parameter limit.
	upsertChunkSize = 500
	// lookupChunkSize bounds the size of IN (...) lists.
	lookupChunkSize = 1000
)

// ListFilter selects entries for paging. Zero values mean "no filter".
type ListFilter struct {
	Page     int
	PageSize int
	Keyword  string  // substring match on value
	Scope    Scope   // 0: all
	Reason   Reason  // 0: all
	Source   Source  // 0: all
	Status   *Status // nil: all (0 is a real status)
	Domain   string
}

// Patch holds the editable fields of an entry; nil fields are left unchanged.
type Patch struct {
	Reason    *Reason
	Note      *string
	ExpiresAt *int64
}

// UpsertResult counts what an upsert did.
type UpsertResult struct {
	Inserted int
	Updated  int
}

// StatCount is the number of entries with a given status and reason.
type StatCount struct {
	Status Status `json:"status"`
	Reason Reason `json:"reason"`
	Total  int    `json:"total"`
}

// Repository persists suppression entries.
type Repository interface {
	// Upsert inserts entries or, on (scope, value) conflict, overwrites their reason, source,
	// status, expiry and timestamps. Hit count, first-seen time and non-empty bounce details
	// already stored are preserved. Entries must be normalized and unique by (scope, value).
	Upsert(ctx context.Context, entries []*entity.EmailBlacklist) (UpsertResult, error)
	GetByID(ctx context.Context, id int64) (*entity.EmailBlacklist, error)
	GetByIDs(ctx context.Context, ids []int64) ([]*entity.EmailBlacklist, error)
	GetByValue(ctx context.Context, scope Scope, value string) (*entity.EmailBlacklist, error)
	List(ctx context.Context, filter ListFilter) (total int, list []*entity.EmailBlacklist, err error)
	Update(ctx context.Context, id int64, patch Patch, now int64) error
	SetStatus(ctx context.Context, ids []int64, status Status, now int64) (int, error)
	Delete(ctx context.Context, ids []int64) (int, error)
	// FindActive returns the entries among values that block sending at time now.
	FindActive(ctx context.Context, scope Scope, values []string, now int64) ([]*entity.EmailBlacklist, error)
	Stats(ctx context.Context) ([]StatCount, error)
}

type pgRepository struct {
	db gdb.DB
}

// NewRepository returns a PostgreSQL-backed Repository.
func NewRepository(db gdb.DB) Repository {
	return &pgRepository{db: db}
}

var upsertColumns = []string{
	"scope", "value", "domain", "reason", "source", "status", "hit_count", "dsn", "last_response",
	"last_task_id", "expires_at", "note", "first_seen_at", "last_seen_at", "create_time", "update_time",
}

// upsertConflictClause preserves stored bounce details when the incoming row has none.
const upsertConflictClause = `
ON CONFLICT (scope, value) DO UPDATE SET
	reason        = EXCLUDED.reason,
	source        = EXCLUDED.source,
	status        = EXCLUDED.status,
	expires_at    = EXCLUDED.expires_at,
	dsn           = COALESCE(NULLIF(EXCLUDED.dsn, ''), t.dsn),
	last_response = COALESCE(NULLIF(EXCLUDED.last_response, ''), t.last_response),
	last_task_id  = COALESCE(NULLIF(EXCLUDED.last_task_id, 0), t.last_task_id),
	note          = COALESCE(NULLIF(EXCLUDED.note, ''), t.note),
	last_seen_at  = EXCLUDED.last_seen_at,
	update_time   = EXCLUDED.update_time
RETURNING (xmax = 0) AS inserted`

func (r *pgRepository) Upsert(ctx context.Context, entries []*entity.EmailBlacklist) (UpsertResult, error) {
	var res UpsertResult
	for start := 0; start < len(entries); start += upsertChunkSize {
		chunk := entries[start:min(start+upsertChunkSize, len(entries))]

		placeholders := make([]string, 0, len(chunk))
		args := make([]any, 0, len(chunk)*len(upsertColumns))
		rowPlaceholder := "(" + strings.TrimSuffix(strings.Repeat("?,", len(upsertColumns)), ",") + ")"
		for _, e := range chunk {
			placeholders = append(placeholders, rowPlaceholder)
			args = append(args, e.Scope, e.Value, e.Domain, e.Reason, e.Source, e.Status, e.HitCount, e.Dsn,
				e.LastResponse, e.LastTaskId, e.ExpiresAt, e.Note, e.FirstSeenAt, e.LastSeenAt, e.CreateTime, e.UpdateTime)
		}

		query := fmt.Sprintf("INSERT INTO %s AS t (%s) VALUES %s %s",
			TableName, strings.Join(upsertColumns, ", "), strings.Join(placeholders, ", "), upsertConflictClause)
		rows, err := r.db.GetAll(ctx, query, args...)
		if err != nil {
			return res, fmt.Errorf("blacklist: upsert: %w", err)
		}
		for _, row := range rows {
			if row["inserted"].Bool() {
				res.Inserted++
			} else {
				res.Updated++
			}
		}
	}
	return res, nil
}

func (r *pgRepository) model(ctx context.Context) *gdb.Model {
	return r.db.Model(TableName).Ctx(ctx).Safe()
}

func (r *pgRepository) GetByID(ctx context.Context, id int64) (*entity.EmailBlacklist, error) {
	return r.getOne(r.model(ctx).Where("id", id))
}

func (r *pgRepository) GetByValue(ctx context.Context, scope Scope, value string) (*entity.EmailBlacklist, error) {
	return r.getOne(r.model(ctx).Where("scope", int(scope)).Where("value", value))
}

func (r *pgRepository) getOne(m *gdb.Model) (*entity.EmailBlacklist, error) {
	var e *entity.EmailBlacklist
	if err := m.Scan(&e); err != nil {
		return nil, fmt.Errorf("blacklist: get: %w", err)
	}
	if e == nil {
		return nil, ErrNotFound
	}
	return e, nil
}

func (r *pgRepository) GetByIDs(ctx context.Context, ids []int64) ([]*entity.EmailBlacklist, error) {
	list := make([]*entity.EmailBlacklist, 0, len(ids))
	if len(ids) == 0 {
		return list, nil
	}
	if err := r.model(ctx).WhereIn("id", ids).OrderAsc("id").Scan(&list); err != nil {
		return nil, fmt.Errorf("blacklist: get by ids: %w", err)
	}
	return list, nil
}

func (r *pgRepository) List(ctx context.Context, f ListFilter) (int, []*entity.EmailBlacklist, error) {
	m := r.model(ctx)
	if f.Keyword != "" {
		m = m.WhereLike("value", "%"+f.Keyword+"%")
	}
	if f.Scope != 0 {
		m = m.Where("scope", int(f.Scope))
	}
	if f.Reason != 0 {
		m = m.Where("reason", int(f.Reason))
	}
	if f.Source != 0 {
		m = m.Where("source", int(f.Source))
	}
	if f.Status != nil {
		m = m.Where("status", int(*f.Status))
	}
	if f.Domain != "" {
		m = m.Where("domain", f.Domain)
	}

	total, err := m.Count()
	if err != nil {
		return 0, nil, fmt.Errorf("blacklist: count: %w", err)
	}

	list := make([]*entity.EmailBlacklist, 0)
	if total == 0 {
		return 0, list, nil
	}
	if err := m.Page(f.Page, f.PageSize).OrderDesc("last_seen_at").OrderDesc("id").Scan(&list); err != nil {
		return 0, nil, fmt.Errorf("blacklist: list: %w", err)
	}
	return total, list, nil
}

func (r *pgRepository) Update(ctx context.Context, id int64, p Patch, now int64) error {
	data := map[string]any{"update_time": now}
	if p.Reason != nil {
		data["reason"] = int(*p.Reason)
	}
	if p.Note != nil {
		data["note"] = *p.Note
	}
	if p.ExpiresAt != nil {
		data["expires_at"] = *p.ExpiresAt
	}

	result, err := r.model(ctx).Where("id", id).Data(data).Update()
	if err != nil {
		return fmt.Errorf("blacklist: update: %w", err)
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgRepository) SetStatus(ctx context.Context, ids []int64, status Status, now int64) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	result, err := r.model(ctx).
		WhereIn("id", ids).
		WhereNot("status", int(status)).
		Data(map[string]any{"status": int(status), "update_time": now}).
		Update()
	if err != nil {
		return 0, fmt.Errorf("blacklist: set status: %w", err)
	}
	n, err := result.RowsAffected()
	return int(n), err
}

func (r *pgRepository) Delete(ctx context.Context, ids []int64) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	result, err := r.model(ctx).WhereIn("id", ids).Delete()
	if err != nil {
		return 0, fmt.Errorf("blacklist: delete: %w", err)
	}
	n, err := result.RowsAffected()
	return int(n), err
}

func (r *pgRepository) FindActive(ctx context.Context, scope Scope, values []string, now int64) ([]*entity.EmailBlacklist, error) {
	found := make([]*entity.EmailBlacklist, 0)
	for start := 0; start < len(values); start += lookupChunkSize {
		chunk := values[start:min(start+lookupChunkSize, len(values))]
		var part []*entity.EmailBlacklist
		err := r.model(ctx).
			Where("scope", int(scope)).
			WhereIn("value", chunk).
			Where("status", int(StatusActive)).
			Where("(expires_at = 0 OR expires_at > ?)", now).
			Scan(&part)
		if err != nil {
			return nil, fmt.Errorf("blacklist: find active: %w", err)
		}
		found = append(found, part...)
	}
	return found, nil
}

func (r *pgRepository) Stats(ctx context.Context) ([]StatCount, error) {
	stats := make([]StatCount, 0)
	err := r.model(ctx).
		Fields("status, reason, COUNT(1) AS total").
		Group("status, reason").
		OrderAsc("status").
		OrderAsc("reason").
		Scan(&stats)
	if err != nil {
		return nil, fmt.Errorf("blacklist: stats: %w", err)
	}
	return stats, nil
}
