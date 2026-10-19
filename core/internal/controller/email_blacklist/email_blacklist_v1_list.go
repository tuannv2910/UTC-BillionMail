package email_blacklist

import (
	"billionmail-core/internal/service/blacklist"
	"billionmail-core/internal/service/public"
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"billionmail-core/api/email_blacklist/v1"
)

func (c *ControllerV1) ListBlacklist(ctx context.Context, req *v1.ListBlacklistReq) (res *v1.ListBlacklistRes, err error) {
	res = &v1.ListBlacklistRes{}

	filter := blacklist.ListFilter{
		Page:     req.Page,
		PageSize: req.PageSize,
		Keyword:  strings.ToLower(strings.TrimSpace(req.Keyword)),
		Scope:    blacklist.Scope(req.Scope),
		Reason:   blacklist.Reason(req.Reason),
		Source:   blacklist.Source(req.Source),
	}
	if req.Status >= 0 {
		status := blacklist.Status(req.Status)
		filter.Status = &status
	}

	total, list, err := blacklist.Default().List(ctx, filter)
	if err != nil {
		res.SetError(gerror.New(errorMessage(ctx, err)))
		return res, nil
	}

	res.Data.Total = total
	res.Data.List = list
	res.SetSuccess(public.LangCtx(ctx, "Success"))
	return
}

func (c *ControllerV1) StatsBlacklist(ctx context.Context, req *v1.StatsBlacklistReq) (res *v1.StatsBlacklistRes, err error) {
	res = &v1.StatsBlacklistRes{}

	stats, err := blacklist.Default().Stats(ctx)
	if err != nil {
		res.SetError(gerror.New(errorMessage(ctx, err)))
		return res, nil
	}

	res.Data.ByReason = make([]v1.ReasonCount, 0)
	for _, s := range stats {
		switch s.Status {
		case blacklist.StatusActive:
			res.Data.Active += s.Total
			res.Data.ByReason = append(res.Data.ByReason, v1.ReasonCount{Reason: int(s.Reason), Total: s.Total})
		case blacklist.StatusObserving:
			res.Data.Observing += s.Total
		case blacklist.StatusReleased:
			res.Data.Released += s.Total
		}
	}
	res.SetSuccess(public.LangCtx(ctx, "Success"))
	return
}
