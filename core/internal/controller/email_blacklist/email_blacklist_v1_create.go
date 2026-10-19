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

func (c *ControllerV1) CreateBlacklist(ctx context.Context, req *v1.CreateBlacklistReq) (res *v1.CreateBlacklistRes, err error) {
	res = &v1.CreateBlacklistRes{}

	entry, err := blacklist.Default().Add(ctx, blacklist.AddInput{
		Scope:  blacklist.Scope(req.Scope),
		Value:  req.Value,
		Reason: blacklist.Reason(req.Reason),
		Source: blacklist.SourceManual,
		Note:   req.Note,
		TTL:    ttlFromDays(req.TtlDays),
	})
	if err = ignoreCacheSync(ctx, err); err != nil {
		res.SetError(gerror.New(errorMessage(ctx, err)))
		return res, nil
	}

	_ = public.WriteLog(ctx, public.LogParams{
		Type: consts.LOGTYPE.EmailBlacklist,
		Log:  "Add to blacklist: " + entry.Value,
		Data: entry,
	})

	res.Data = entry
	res.SetSuccess(public.LangCtx(ctx, "Added to blacklist"))
	return
}

func (c *ControllerV1) ImportBlacklist(ctx context.Context, req *v1.ImportBlacklistReq) (res *v1.ImportBlacklistRes, err error) {
	res = &v1.ImportBlacklistRes{}
	if len(req.Values) > maxImportValues {
		res.SetError(gerror.New(public.LangCtx(ctx, "At most {} entries can be imported at once", maxImportValues)))
		return res, nil
	}

	inputs := make([]blacklist.AddInput, 0, len(req.Values))
	for _, value := range req.Values {
		inputs = append(inputs, blacklist.AddInput{
			Scope:  blacklist.Scope(req.Scope),
			Value:  value,
			Reason: blacklist.Reason(req.Reason),
			Source: blacklist.SourceImport,
			Note:   req.Note,
			TTL:    ttlFromDays(req.TtlDays),
		})
	}

	result, err := blacklist.Default().AddBatch(ctx, inputs)
	if err = ignoreCacheSync(ctx, err); err != nil {
		res.SetError(gerror.New(errorMessage(ctx, err)))
		return res, nil
	}

	res.Data.Inserted = result.Inserted
	res.Data.Updated = result.Updated
	res.Data.Invalid = make([]v1.InvalidValue, 0, len(result.Invalid))
	for _, item := range result.Invalid {
		res.Data.Invalid = append(res.Data.Invalid, v1.InvalidValue{Value: item.Value, Error: item.Error})
	}

	_ = public.WriteLog(ctx, public.LogParams{
		Type: consts.LOGTYPE.EmailBlacklist,
		Log: fmt.Sprintf("Import blacklist: %d added, %d updated, %d invalid",
			result.Inserted, result.Updated, len(result.Invalid)),
	})

	res.SetSuccess(public.LangCtx(ctx, "Import completed"))
	return
}
