package email_template

import (
	"billionmail-core/internal/service/askai"
	"billionmail-core/internal/service/content_gen"
	"billionmail-core/internal/service/content_review"
	"billionmail-core/internal/service/public"
	"context"
	"errors"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"billionmail-core/api/email_template/v1"
)

const maxAiContentBytes = 512 * 1024

func (c *ControllerV1) AiReviewContent(ctx context.Context, req *v1.AiReviewContentReq) (res *v1.AiReviewContentRes, err error) {
	res = &v1.AiReviewContentRes{}
	if len(req.Content) > maxAiContentBytes {
		res.SetError(gerror.New(public.LangCtx(ctx, "Email content is too large for AI processing")))
		return
	}

	provider, model, err := askai.NewLLMProvider(req.SupplierName, req.ModelId)
	if err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, err.Error())))
		return res, nil
	}

	result, err := content_review.NewReviewer(provider, content_review.WithModel(model.ModelId)).Review(ctx, content_review.Input{
		Subject:            req.Subject,
		Content:            req.Content,
		Language:           req.Language,
		UnsubscribeManaged: req.UnsubscribeManaged,
	})
	if err != nil {
		res.SetError(gerror.New(aiErrorMessage(ctx, err)))
		return res, nil
	}

	res.Data = result
	res.SetSuccess(public.LangCtx(ctx, "Review completed"))
	return
}

func (c *ControllerV1) AiGenerateContent(ctx context.Context, req *v1.AiGenerateContentReq) (res *v1.AiGenerateContentRes, err error) {
	res = &v1.AiGenerateContentRes{}
	if len(req.CurrentHtml) > maxAiContentBytes {
		res.SetError(gerror.New(public.LangCtx(ctx, "Email content is too large for AI processing")))
		return
	}

	provider, model, err := askai.NewLLMProvider(req.SupplierName, req.ModelId)
	if err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, err.Error())))
		return res, nil
	}

	opts := []content_gen.Option{content_gen.WithModel(model.ModelId)}
	if model.MaxTokens > 0 {
		opts = append(opts, content_gen.WithMaxTokens(model.MaxTokens))
	}
	out, err := content_gen.NewGenerator(provider, opts...).Generate(ctx, content_gen.Input{
		Brief:       req.Brief,
		Subject:     req.Subject,
		CurrentHTML: req.CurrentHtml,
		Tone:        req.Tone,
		Language:    req.Language,
	})
	if err != nil {
		res.SetError(gerror.New(aiErrorMessage(ctx, err)))
		return res, nil
	}

	res.Data = out
	res.SetSuccess(public.LangCtx(ctx, "Content generated"))
	return
}

// aiErrorMessage turns provider and parsing errors into a message the user can act on.
func aiErrorMessage(ctx context.Context, err error) string {
	g.Log().Warning(ctx, "AI content request failed:", err)

	switch {
	case errors.Is(err, content_review.ErrEmptyInput):
		return public.LangCtx(ctx, "Subject and content are both empty")
	case errors.Is(err, content_gen.ErrEmptyBrief):
		return public.LangCtx(ctx, "Please describe the email you want")
	case errors.Is(err, content_review.ErrInvalidModelOutput), errors.Is(err, content_gen.ErrInvalidModelOutput):
		return public.LangCtx(ctx, "The AI model returned an unexpected answer, please try again or choose another model")
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		return public.LangCtx(ctx, "The AI model took too long to respond, please try again")
	}
	return public.LangCtx(ctx, "AI request failed: {}", err.Error())
}
