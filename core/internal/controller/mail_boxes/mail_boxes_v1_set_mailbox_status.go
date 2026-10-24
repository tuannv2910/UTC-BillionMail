package mail_boxes

import (
	"billionmail-core/api/mail_boxes/v1"
	"billionmail-core/internal/consts"
	"billionmail-core/internal/service/mail_boxes"
	"billionmail-core/internal/service/public"
	"context"
	"fmt"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) SetMailboxStatus(ctx context.Context, req *v1.SetMailboxStatusReq) (res *v1.SetMailboxStatusRes, err error) {
	res = &v1.SetMailboxStatusRes{}

	affected, err := mail_boxes.SetActive(ctx, req.Username, req.Active)
	if err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to update mailbox status: {}", err.Error())))
		return res, nil
	}

	if affected == 0 {
		res.SetError(gerror.New(public.LangCtx(ctx, "Mailbox not found")))
		return res, nil
	}

	_ = public.WriteLog(ctx, public.LogParams{
		Type: consts.LOGTYPE.Mailboxes,
		Log:  fmt.Sprintf("Set mailbox %s status to %d successfully", req.Username, req.Active),
	})

	res.SetSuccess(public.LangCtx(ctx, "Mailbox updated successfully"))
	return res, nil
}
