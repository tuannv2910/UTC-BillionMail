package settings

import (
	"billionmail-core/api/settings/v1"
	"billionmail-core/internal/consts"
	"billionmail-core/internal/service/domains"
	"billionmail-core/internal/service/public"
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) GetAlertWebhook(ctx context.Context, req *v1.GetAlertWebhookReq) (res *v1.GetAlertWebhookRes, err error) {
	res = &v1.GetAlertWebhookRes{}
	res.Data = domains.LoadAlertWebhookSettings(ctx)
	res.SetSuccess(public.LangCtx(ctx, "Success"))
	return res, nil
}

func (c *ControllerV1) SetAlertWebhook(ctx context.Context, req *v1.SetAlertWebhookReq) (res *v1.SetAlertWebhookRes, err error) {
	res = &v1.SetAlertWebhookRes{}

	settings := domains.AlertWebhookSettings{
		Enabled: req.Enabled,
		URL:     req.URL,
		Format:  req.Format,
	}

	if err := domains.ValidateAlertWebhook(settings); err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, "Parameter validation failed: {}", err.Error())))
		return res, nil
	}

	if req.SendTest && settings.URL != "" {
		ev := domains.AlertEvent{
			Event:   "test",
			Title:   "BillionMail alert webhook test",
			Message: "Your webhook alert channel is configured correctly.",
			Time:    time.Now().Unix(),
		}
		if err := domains.SendAlertWebhook(ctx, settings, ev); err != nil {
			res.SetError(gerror.New(public.LangCtx(ctx, "Webhook test failed: {}", err.Error())))
			return res, nil
		}
	}

	if err := domains.SaveAlertWebhookSettings(ctx, settings); err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to save alert settings: {}", err.Error())))
		return res, nil
	}

	_ = public.WriteLog(ctx, public.LogParams{
		Type: consts.LOGTYPE.Settings,
		Log:  "Updated alert webhook settings successfully",
	})

	res.SetSuccess(public.LangCtx(ctx, "Alert settings saved successfully"))
	return res, nil
}
