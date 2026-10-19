package email_blacklist

import (
	"billionmail-core/internal/service/blacklist"
	"billionmail-core/internal/service/public"
	"context"
	"errors"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

const maxImportValues = 10000

// ttlFromDays maps the API's ttl_days to a Manager TTL: -1 uses the reason default, 0 is permanent.
func ttlFromDays(days int) *time.Duration {
	if days < 0 {
		return nil
	}
	ttl := time.Duration(days) * 24 * time.Hour
	return &ttl
}

// errorMessage turns blacklist errors into a translated message.
func errorMessage(ctx context.Context, err error) string {
	switch {
	case errors.Is(err, blacklist.ErrNotFound):
		return public.LangCtx(ctx, "Blacklist entry not found")
	case errors.Is(err, blacklist.ErrInvalidEmail):
		return public.LangCtx(ctx, "Invalid email address")
	case errors.Is(err, blacklist.ErrInvalidDomain):
		return public.LangCtx(ctx, "Invalid domain")
	case errors.Is(err, blacklist.ErrInvalidReason):
		return public.LangCtx(ctx, "Invalid reason")
	case errors.Is(err, blacklist.ErrInvalidScope):
		return public.LangCtx(ctx, "Invalid scope")
	}
	g.Log().Error(ctx, "email blacklist:", err)
	return public.LangCtx(ctx, "Operation failed: {}", err.Error())
}

// ignoreCacheSync treats a cache sync failure as success: the database is the source of truth
// and the cache is rebuilt from it.
func ignoreCacheSync(ctx context.Context, err error) error {
	if errors.Is(err, blacklist.ErrCacheSync) {
		g.Log().Warning(ctx, "email blacklist cache sync:", err)
		return nil
	}
	return err
}
