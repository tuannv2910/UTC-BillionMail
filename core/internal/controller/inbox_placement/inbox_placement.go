package inbox_placement

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/errors/gerror"

	"billionmail-core/internal/service/inbox_placement"
	"billionmail-core/internal/service/public"
	"billionmail-core/utility/types/api_v1"
)

// fail sets a user-facing error on res. Errors answered by the placement
// service already carry a readable message.
func fail(ctx context.Context, res *api_v1.StandardRes, err error) {
	var apiErr *inbox_placement.APIError
	switch {
	case errors.Is(err, inbox_placement.ErrDisabled):
		res.Code = 503
		res.SetError(gerror.New(public.LangCtx(ctx, "Inbox placement service is not configured")))
	case errors.Is(err, inbox_placement.ErrAdminDisabled):
		res.Code = 503
		res.SetError(gerror.New(public.LangCtx(ctx, "Seed administration is not configured")))
	case errors.As(err, &apiErr):
		res.Code = apiErr.Status
		res.SetError(gerror.New(apiErr.Error()))
	default:
		res.Code = 500
		res.SetError(gerror.New(public.LangCtx(ctx, "Inbox placement service is unavailable")))
	}
}
