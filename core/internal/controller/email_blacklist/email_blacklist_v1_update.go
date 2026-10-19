package email_blacklist

import (
	"billionmail-core/internal/consts"
	"billionmail-core/internal/service/blacklist"
	"billionmail-core/internal/service/public"
	"context"
	"fmt"

	"github.com/gogf/gf/v2/errors/gerror"

	"billionmail-core/api/email_blacklist/v1"
)

func (c *ControllerV1) UpdateBlacklist(ctx context.Context, req *v1.UpdateBlacklistReq) (res *v1.UpdateBlacklistRes, err error) {
	res = &v1.UpdateBlacklistRes{}

	reason := blacklist.Reason(req.Reason)
	in := blacklist.UpdateInput{Reason: &reason, Note: &req.Note}
	switch req.ExpireMode {
	case "permanent":
		var permanent int64
		in.ExpiresAt = &permanent
	case "until":
		in.ExpiresAt = &req.ExpiresAt
	}

	entry, err := blacklist.Default().Update(ctx, req.Id, in)
	if err = ignoreCacheSync(ctx, err); err != nil {
		res.SetError(gerror.New(errorMessage(ctx, err)))
		return res, nil
	}

	_ = public.WriteLog(ctx, public.LogParams{
		Type: consts.LOGTYPE.EmailBlacklist,
		Log:  "Update blacklist entry: " + entry.Value,
		Data: entry,
	})

	res.Data = entry
	res.SetSuccess(public.LangCtx(ctx, "Updated successfully"))
	return
}

func (c *ControllerV1) ReleaseBlacklist(ctx context.Context, req *v1.ReleaseBlacklistReq) (res *v1.ReleaseBlacklistRes, err error) {
	res = &v1.ReleaseBlacklistRes{}

	n, err := blacklist.Default().Release(ctx, req.Ids)
	if err = ignoreCacheSync(ctx, err); err != nil {
		res.SetError(gerror.New(errorMessage(ctx, err)))
		return res, nil
	}

	_ = public.WriteLog(ctx, public.LogParams{
		Type: consts.LOGTYPE.EmailBlacklist,
		Log:  fmt.Sprintf("Release %d blacklist entries", n),
		Data: req.Ids,
	})

	res.Data.Count = n
	res.SetSuccess(public.LangCtx(ctx, "Released successfully"))
	return
}

func (c *ControllerV1) DeleteBlacklist(ctx context.Context, req *v1.DeleteBlacklistReq) (res *v1.DeleteBlacklistRes, err error) {
	res = &v1.DeleteBlacklistRes{}

	n, err := blacklist.Default().Delete(ctx, req.Ids)
	if err = ignoreCacheSync(ctx, err); err != nil {
		res.SetError(gerror.New(errorMessage(ctx, err)))
		return res, nil
	}

	_ = public.WriteLog(ctx, public.LogParams{
		Type: consts.LOGTYPE.EmailBlacklist,
		Log:  fmt.Sprintf("Delete %d blacklist entries", n),
		Data: req.Ids,
	})

	res.Data.Count = n
	res.SetSuccess(public.LangCtx(ctx, "Deleted successfully"))
	return
}
