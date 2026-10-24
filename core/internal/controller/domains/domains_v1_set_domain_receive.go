package domains

import (
	"billionmail-core/api/domains/v1"
	"billionmail-core/internal/consts"
	"billionmail-core/internal/service/public"
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// SetDomainReceive marks a domain as send-only (receive_enabled = 0) or as a
// normal local mail domain. Postfix reads the flag through
// pgsql_virtual_domains_maps.cf, so no reload is needed (#303).
func (c *ControllerV1) SetDomainReceive(ctx context.Context, req *v1.SetDomainReceiveReq) (res *v1.SetDomainReceiveRes, err error) {
	res = &v1.SetDomainReceiveRes{}

	domain := strings.ToLower(strings.TrimSpace(req.Domain))
	result, err := g.DB().Model("domain").Ctx(ctx).
		Where("domain", domain).
		Update(g.Map{"receive_enabled": req.ReceiveEnabled})
	if err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to update domain: {}", err.Error())))
		return res, nil
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		res.SetError(gerror.New(public.LangCtx(ctx, "Domain not found")))
		return res, nil
	}

	_ = public.WriteLog(ctx, public.LogParams{
		Type: consts.LOGTYPE.Domain,
		Log:  fmt.Sprintf("Set receive_enabled=%d for domain %s successfully", req.ReceiveEnabled, domain),
	})

	res.SetSuccess(public.LangCtx(ctx, "Success"))
	return res, nil
}
