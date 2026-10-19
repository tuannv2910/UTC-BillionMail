package v1

import (
	"billionmail-core/internal/service/content_gen"
	"billionmail-core/internal/service/content_review"
	"billionmail-core/utility/types/api_v1"

	"github.com/gogf/gf/v2/frame/g"
)

type AiReviewContentReq struct {
	g.Meta             `path:"/email_template/ai_review" method:"post" tags:"EmailTemplate" summary:"Review email content with AI"`
	Authorization      string `json:"authorization" dc:"Authorization" in:"header"`
	Subject            string `json:"subject" dc:"Email subject"`
	Content            string `json:"content" dc:"Email HTML content"`
	Language           string `json:"language" dc:"Language of reasons and suggestions (en, vi, zh, ja)"`
	UnsubscribeManaged bool   `json:"unsubscribe_managed" dc:"Whether the platform appends the unsubscribe link"`
	SupplierName       string `json:"supplier_name" dc:"AI supplier; empty uses the first enabled model"`
	ModelId            string `json:"model_id" dc:"AI model; empty uses the first enabled model"`
}

type AiReviewContentRes struct {
	api_v1.StandardRes
	Data *content_review.Result `json:"data"`
}

type AiGenerateContentReq struct {
	g.Meta        `path:"/email_template/ai_generate" method:"post" tags:"EmailTemplate" summary:"Write or revise an email with AI"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	Brief         string `json:"brief" v:"required" dc:"What the email is about, or how to change current_html"`
	Subject       string `json:"subject" dc:"Current or desired subject"`
	CurrentHtml   string `json:"current_html" dc:"Email to revise; empty writes a new one"`
	Tone          string `json:"tone" dc:"Tone, e.g. friendly or professional"`
	Language      string `json:"language" dc:"Language of the email (en, vi, zh, ja)"`
	SupplierName  string `json:"supplier_name" dc:"AI supplier; empty uses the first enabled model"`
	ModelId       string `json:"model_id" dc:"AI model; empty uses the first enabled model"`
}

type AiGenerateContentRes struct {
	api_v1.StandardRes
	Data *content_gen.Output `json:"data"`
}
