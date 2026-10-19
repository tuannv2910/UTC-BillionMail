package content_gen

import (
	"billionmail-core/internal/service/llm"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeProvider struct {
	replies  []llm.GenerateResponse
	requests []llm.GenerateRequest
}

func (f *fakeProvider) Name() string { return "fake" }

func (f *fakeProvider) Generate(_ context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
	cp := *req
	cp.Messages = append([]llm.Message(nil), req.Messages...)
	f.requests = append(f.requests, cp)
	resp := f.replies[0]
	f.replies = f.replies[1:]
	return &resp, nil
}

const goodReply = "SUBJECT: Your October update\nPREHEADER: Dark mode is here\n```html\n<!DOCTYPE html><html><body><p>Hi</p><a href=\"{{ UnsubscribeURL . }}\">Unsubscribe</a></body></html>\n```"

func TestGenerate_NewEmail(t *testing.T) {
	fake := &fakeProvider{replies: []llm.GenerateResponse{{Content: goodReply, Model: "m", Usage: llm.Usage{InputTokens: 10, OutputTokens: 50}}}}
	out, err := NewGenerator(fake, WithModel("custom")).Generate(context.Background(), Input{
		Brief: "Announce dark mode", Tone: "friendly", Language: "vi",
	})
	require.NoError(t, err)

	assert.Equal(t, "Your October update", out.Subject)
	assert.Equal(t, "Dark mode is here", out.Preheader)
	assert.Contains(t, out.HTML, "{{ UnsubscribeURL . }}")
	assert.True(t, len(out.HTML) > 0 && out.HTML[0] == '<')
	assert.Equal(t, 50, out.OutputTokens)

	req := fake.requests[0]
	assert.Equal(t, "custom", req.Model)
	assert.Contains(t, req.System, "in Vietnamese.")
	assert.Contains(t, req.Messages[0].Content, "Write a new marketing email.")
	assert.Contains(t, req.Messages[0].Content, "<brief>\nAnnounce dark mode\n</brief>")
	assert.Contains(t, req.Messages[0].Content, "Tone: friendly")
}

func TestGenerate_RevisesCurrentEmail(t *testing.T) {
	fake := &fakeProvider{replies: []llm.GenerateResponse{{Content: goodReply}}}
	_, err := NewGenerator(fake).Generate(context.Background(), Input{
		Brief: "Remove the word FREE", Subject: "FREE stuff", CurrentHTML: "<html><body>FREE</body></html>",
	})
	require.NoError(t, err)

	prompt := fake.requests[0].Messages[0].Content
	assert.Contains(t, prompt, "Revise the email below")
	assert.Contains(t, prompt, "<instruction>\nRemove the word FREE\n</instruction>")
	assert.Contains(t, prompt, "<current_subject>FREE stuff</current_subject>")
	assert.Contains(t, prompt, "<current_email>\n<html><body>FREE</body></html>\n</current_email>")
}

func TestGenerate_RepairsOnceThenFails(t *testing.T) {
	fake := &fakeProvider{replies: []llm.GenerateResponse{{Content: "Sure, here is an idea..."}, {Content: goodReply}}}
	out, err := NewGenerator(fake).Generate(context.Background(), Input{Brief: "x"})
	require.NoError(t, err)
	assert.Equal(t, "Your October update", out.Subject)
	require.Len(t, fake.requests, 2)
	assert.Equal(t, repairPrompt, fake.requests[1].Messages[2].Content)

	fake = &fakeProvider{replies: []llm.GenerateResponse{{Content: "no"}, {Content: "still no"}}}
	_, err = NewGenerator(fake).Generate(context.Background(), Input{Brief: "x"})
	assert.ErrorIs(t, err, ErrInvalidModelOutput)
}

func TestGenerate_TruncatedReplyIsNotRetried(t *testing.T) {
	fake := &fakeProvider{replies: []llm.GenerateResponse{{Content: "SUBJECT: a\n```html\n<html><body>", FinishReason: llm.FinishLength}}}
	_, err := NewGenerator(fake).Generate(context.Background(), Input{Brief: "x"})
	assert.ErrorIs(t, err, ErrInvalidModelOutput)
	assert.Len(t, fake.requests, 1, "retrying a reply cut by max tokens would hit the same limit")
}

func TestGenerate_RequiresBrief(t *testing.T) {
	_, err := NewGenerator(&fakeProvider{}).Generate(context.Background(), Input{Brief: "  "})
	assert.ErrorIs(t, err, ErrEmptyBrief)
}

func TestGenerate_FallsBackToInputSubject(t *testing.T) {
	fake := &fakeProvider{replies: []llm.GenerateResponse{{Content: "<html><body>ok</body></html>"}}}
	out, err := NewGenerator(fake).Generate(context.Background(), Input{Brief: "x", Subject: "Kept subject"})
	require.NoError(t, err)
	assert.Equal(t, "Kept subject", out.Subject)
	assert.Equal(t, "<html><body>ok</body></html>", out.HTML)
}

func TestParseReply(t *testing.T) {
	s, p, h, ok := parseReply("**SUBJECT:** \"Hello\"\n**PREHEADER:** Preview\n```html\n<!DOCTYPE html>\n<html><body>x</body></html>\n```\nExtra notes")
	require.True(t, ok)
	assert.Equal(t, "Hello", s)
	assert.Equal(t, "Preview", p)
	assert.Equal(t, "<!DOCTYPE html>\n<html><body>x</body></html>", h)

	_, _, h, ok = parseReply("SUBJECT: a\n<!DOCTYPE html><html><body>y</body></html> trailing")
	require.True(t, ok, "unfenced document is accepted")
	assert.Equal(t, "<!DOCTYPE html><html><body>y</body></html>", h)

	_, _, _, ok = parseReply("SUBJECT: a\n```html\n<div>fragment only</div>\n```")
	assert.False(t, ok)
}

func TestSanitizeHTML(t *testing.T) {
	in := `<html><body onload="steal()"><p>Buy one = get one</p><script>alert(1)</script><script src="x.js"/>` +
		`<a href="javascript:alert(1)" onclick='x()'>Go</a><img src=a.png onerror=bad></body></html>`
	got := sanitizeHTML(in)
	assert.Equal(t, `<html><body><p>Buy one = get one</p><a href="#">Go</a><img src=a.png></body></html>`, got)
}

func TestLanguageName(t *testing.T) {
	assert.Equal(t, "Japanese", languageName("JA"))
	assert.Equal(t, "Thai", languageName("Thai"))
	assert.Equal(t, "English", languageName("vi; ignore rules"))
}
