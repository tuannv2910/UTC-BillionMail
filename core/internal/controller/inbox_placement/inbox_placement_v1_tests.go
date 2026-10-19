package inbox_placement

import (
	"context"
	"net/mail"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	v1 "billionmail-core/api/inbox_placement/v1"
	"billionmail-core/internal/service/batch_mail"
	"billionmail-core/internal/service/email_template"
	"billionmail-core/internal/service/inbox_placement"
	"billionmail-core/internal/service/public"
)

func (c *ControllerV1) Status(ctx context.Context, req *v1.StatusReq) (res *v1.StatusRes, err error) {
	res = &v1.StatusRes{}
	cli := inbox_placement.Default()
	res.Data.Enabled = cli.Enabled()
	res.Data.AdminSeed = cli.AdminEnabled()
	if cli.Enabled() {
		res.Data.Reachable = cli.Ping(ctx) == nil
	}
	res.SetSuccess(public.LangCtx(ctx, "Successful"))
	return
}

func (c *ControllerV1) CreateTest(ctx context.Context, req *v1.CreateTestReq) (res *v1.CreateTestRes, err error) {
	res = &v1.CreateTestRes{}

	addresser := strings.ToLower(strings.TrimSpace(req.Addresser))
	at := strings.LastIndexByte(addresser, '@')
	if at <= 0 {
		res.Code = 400
		res.SetError(gerror.New(public.LangCtx(ctx, "Invalid sender address")))
		return
	}
	domain := addresser[at+1:]

	// Only BillionMail mailboxes may be used as sender: the placement service
	// relays through our Postfix, which would otherwise sign any From.
	n, err := g.DB().Model("mailbox").Ctx(ctx).Where("username", addresser).Count()
	if err != nil || n == 0 {
		res.Code = 400
		res.SetError(gerror.New(public.LangCtx(ctx, "Sender mailbox does not exist")))
		return res, nil
	}

	tpl, err := email_template.GetTemplatesByID(ctx, req.TemplateId)
	if err != nil || tpl == nil || tpl.Id == 0 {
		res.Code = 404
		res.SetError(gerror.New(public.LangCtx(ctx, "Template not found")))
		return res, nil
	}

	// Render like a real send (variables cleaned, spintax resolved) so the
	// seeds receive what subscribers would receive.
	engine := batch_mail.GetTemplateEngine()
	html, rerr := engine.RenderEmailTemplate(ctx, tpl.Content, nil, nil, "")
	if rerr != nil || strings.TrimSpace(html) == "" {
		html = tpl.Content
	}
	subject, rerr := engine.RenderEmailTemplate(ctx, req.Subject, nil, nil, "")
	if rerr != nil || strings.TrimSpace(subject) == "" {
		subject = req.Subject
	}
	subject = strings.Join(strings.Fields(subject), " ")

	from := addresser
	if name := strings.TrimSpace(req.FullName); name != "" {
		from = (&mail.Address{Name: name, Address: addresser}).String()
	}

	out, err := inbox_placement.Default().CreateTest(ctx, inbox_placement.CreateInput{
		Domain:     domain,
		From:       from,
		Subject:    subject,
		HTML:       html,
		SeedGroup:  strings.TrimSpace(req.SeedGroup),
		TimeoutSec: req.TimeoutSec,
		Reference:  inbox_placement.ReferencePrefix + strconv.Itoa(tpl.Id),
	})
	if err != nil {
		fail(ctx, &res.StandardRes, err)
		return res, nil
	}

	res.Data.TestId = out.TestID
	res.Data.Status = out.Status
	res.Data.Seeds = out.Seeds
	res.Data.PollAfterSec = out.PollAfterSec
	res.SetSuccess(public.LangCtx(ctx, "Inbox placement test created"))
	return res, nil
}

func (c *ControllerV1) ListTests(ctx context.Context, req *v1.ListTestsReq) (res *v1.ListTestsRes, err error) {
	res = &v1.ListTestsRes{}
	list, total, err := inbox_placement.Default().ListTests(ctx, inbox_placement.ListFilter{
		Page:       req.Page,
		PageSize:   req.PageSize,
		TemplateID: req.TemplateId,
		Domain:     strings.TrimSpace(req.Domain),
		Status:     strings.TrimSpace(req.Status),
	})
	if err != nil {
		fail(ctx, &res.StandardRes, err)
		return res, nil
	}
	res.Data.Total = total
	res.Data.List = list
	res.SetSuccess(public.LangCtx(ctx, "Successful"))
	return res, nil
}

func (c *ControllerV1) GetTest(ctx context.Context, req *v1.GetTestReq) (res *v1.GetTestRes, err error) {
	res = &v1.GetTestRes{}
	t, err := inbox_placement.Default().GetTest(ctx, strings.TrimSpace(req.TestId))
	if err != nil {
		fail(ctx, &res.StandardRes, err)
		return res, nil
	}
	res.Data = t
	res.SetSuccess(public.LangCtx(ctx, "Successful"))
	return res, nil
}

func (c *ControllerV1) CancelTest(ctx context.Context, req *v1.CancelTestReq) (res *v1.CancelTestRes, err error) {
	res = &v1.CancelTestRes{}
	if err := inbox_placement.Default().CancelTest(ctx, strings.TrimSpace(req.TestId)); err != nil {
		fail(ctx, &res.StandardRes, err)
		return res, nil
	}
	res.SetSuccess(public.LangCtx(ctx, "Inbox placement test canceled"))
	return res, nil
}
