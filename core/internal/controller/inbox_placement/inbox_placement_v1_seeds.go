package inbox_placement

import (
	"context"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	v1 "billionmail-core/api/inbox_placement/v1"
	"billionmail-core/internal/service/inbox_placement"
	"billionmail-core/internal/service/public"
)

func (c *ControllerV1) ListSeeds(ctx context.Context, req *v1.ListSeedsReq) (res *v1.ListSeedsRes, err error) {
	res = &v1.ListSeedsRes{}
	list, err := inbox_placement.Default().ListSeeds(ctx)
	if err != nil {
		fail(ctx, &res.StandardRes, err)
		return res, nil
	}
	res.Data.Total = len(list)
	res.Data.List = list
	res.SetSuccess(public.LangCtx(ctx, "Successful"))
	return res, nil
}

func (c *ControllerV1) AddSeed(ctx context.Context, req *v1.AddSeedReq) (res *v1.AddSeedRes, err error) {
	res = &v1.AddSeedRes{}
	id, err := inbox_placement.Default().AddSeed(ctx, inbox_placement.AddSeedInput{
		Email:        strings.TrimSpace(req.Email),
		RefreshToken: strings.TrimSpace(req.RefreshToken),
		SeedGroup:    strings.TrimSpace(req.SeedGroup),
		OAuthClient:  strings.TrimSpace(req.OAuthClient),
	})
	if err != nil {
		fail(ctx, &res.StandardRes, err)
		return res, nil
	}
	res.Data.Id = id
	res.SetSuccess(public.LangCtx(ctx, "Seed mailbox saved"))
	return res, nil
}

func (c *ControllerV1) SeedOAuth(ctx context.Context, req *v1.SeedOAuthReq) (res *v1.SeedOAuthRes, err error) {
	res = &v1.SeedOAuthRes{}
	u, err := inbox_placement.Default().SeedOAuthURL(ctx, strings.TrimSpace(req.SeedGroup), strings.TrimSpace(req.OAuthClient))
	if err != nil {
		fail(ctx, &res.StandardRes, err)
		return res, nil
	}
	res.Data.Url = u
	res.SetSuccess(public.LangCtx(ctx, "Successful"))
	return res, nil
}

func (c *ControllerV1) UpdateSeed(ctx context.Context, req *v1.UpdateSeedReq) (res *v1.UpdateSeedRes, err error) {
	res = &v1.UpdateSeedRes{}
	if err := inbox_placement.Default().UpdateSeed(ctx, req.Id, inbox_placement.SeedPatch{
		IsActive: req.IsActive, SeedGroup: req.SeedGroup,
	}); err != nil {
		fail(ctx, &res.StandardRes, err)
		return res, nil
	}
	res.SetSuccess(public.LangCtx(ctx, "Seed mailbox updated"))
	return res, nil
}

func (c *ControllerV1) CheckSeed(ctx context.Context, req *v1.CheckSeedReq) (res *v1.CheckSeedRes, err error) {
	res = &v1.CheckSeedRes{}
	seed, err := inbox_placement.Default().CheckSeed(ctx, req.Id)
	if err != nil {
		fail(ctx, &res.StandardRes, err)
		return res, nil
	}
	res.Data = seed
	res.SetSuccess(public.LangCtx(ctx, "Health check finished"))
	return res, nil
}

// ServeOAuthCallback receives Google's consent redirect for a new seed and
// forwards it to the placement service, which validates the signed state.
// Public route (Google cannot send the BillionMail JWT); configure
// https://<billionmail>/inbox-placement/oauth/callback as redirect_url.
func ServeOAuthCallback(r *ghttp.Request) {
	status, body, err := inbox_placement.Default().OAuthCallback(r.Context(), r.URL.RawQuery)
	r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err != nil {
		r.Response.WriteStatus(http.StatusBadGateway, "<h3>Inbox placement service is unavailable</h3>")
		return
	}
	// Close the consent popup and let the seed list refresh itself.
	page := `<!DOCTYPE html><html><head><meta charset="utf-8"><title>Seed mailbox</title>
<style>body{font-family:-apple-system,"Segoe UI",Roboto,sans-serif;padding:40px;color:#333}</style></head><body>` +
		string(body) +
		`<script>try{window.opener&&window.opener.postMessage({type:"inbox-placement-seed",ok:` + boolJS(status == http.StatusOK) +
		`},window.location.origin)}catch(e){}</script></body></html>`
	r.Response.WriteStatus(status, page)
}

func boolJS(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
