package content_gen

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const systemPromptTemplate = `You are a senior email marketer and HTML email developer. You write marketing emails that render
correctly in Gmail, Outlook, Apple Mail and on mobile, and that pass spam filters.

## HTML rules
- Output one complete HTML document: <!DOCTYPE html>, <html>, <head> with <meta charset="utf-8"> and a
  viewport meta tag, and <body>.
- Lay out with tables (role="presentation"), centered, at most 600px wide. Put CSS inline on elements;
  a <style> block in <head> is allowed only for media queries.
- No JavaScript, forms, iframes, video, external stylesheets or web fonts. Use web-safe font stacks.
- Use images only when the brief gives their URLs; never invent image URLs. Every image needs alt text
  and a width. Keep most of the message as live text.
- Use link URLs from the brief. When a call-to-action URL is missing, use "#".
- Buttons are bulletproof table buttons, not images.
- Start <body> with a hidden preheader: <div style="display:none;max-height:0;overflow:hidden;">PREHEADER</div>.
- End with a footer containing the sender name when known, one line on why the recipient gets this email,
  and an unsubscribe link whose href is exactly {{ UnsubscribeURL . }}
- Template variables are filled in at send time. You may use {{.Subscriber.Email}}. Do not invent other
  variables. Keep any {{ ... }} variable or {option A|option B} spintax that already exists in an email
  you revise.

## Deliverability rules
- Honest subject of at most 60 characters that matches the body. Never start it with "Re:" or "Fwd:".
- No spam phrases ("100% free", "act now", "risk-free", "guaranteed", "click here", "$$$"), no ALL CAPS
  words, at most one exclamation mark in the whole email, no fake urgency, no exaggerated claims.
- One clear call to action.

## Language
Write the subject, preheader and email text in %LANGUAGE%.

## Output format
Reply with exactly this and nothing else:
SUBJECT: <subject line>
PREHEADER: <preview text, at most 100 characters>
` + "```html" + `
<the complete HTML document>
` + "```"

const repairPrompt = "Your reply did not follow the output format. Reply again with exactly SUBJECT, PREHEADER and the complete HTML document in a ```html fenced block, nothing else."

var languageNames = map[string]string{
	"en": "English",
	"vi": "Vietnamese",
	"zh": "Simplified Chinese",
	"ja": "Japanese",
}

// languageName maps a code to a name; unknown values must look like a plain language name.
func languageName(lang string) string {
	lang = strings.TrimSpace(lang)
	if name, ok := languageNames[strings.ToLower(lang)]; ok {
		return name
	}
	if lang == "" || len(lang) > 32 {
		return languageNames["en"]
	}
	for _, c := range lang {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == ' ' || c == '-') {
			return languageNames["en"]
		}
	}
	return lang
}

func buildSystemPrompt(language string) string {
	return strings.ReplaceAll(systemPromptTemplate, "%LANGUAGE%", languageName(language))
}

func buildUserPrompt(in Input) string {
	var sb strings.Builder
	if in.CurrentHTML != "" {
		sb.WriteString("Revise the email below following the instruction. Keep its structure, branding, links, ")
		sb.WriteString("images and template variables unless the instruction says otherwise.\n\n")
		fmt.Fprintf(&sb, "<instruction>\n%s\n</instruction>\n\n", truncateRunes(in.Brief, maxBriefRunes))
		if s := strings.TrimSpace(in.Subject); s != "" {
			fmt.Fprintf(&sb, "<current_subject>%s</current_subject>\n\n", s)
		}
		fmt.Fprintf(&sb, "<current_email>\n%s\n</current_email>", truncateRunes(in.CurrentHTML, maxCurrentRunes))
	} else {
		sb.WriteString("Write a new marketing email.\n\n")
		fmt.Fprintf(&sb, "<brief>\n%s\n</brief>", truncateRunes(in.Brief, maxBriefRunes))
		if s := strings.TrimSpace(in.Subject); s != "" {
			fmt.Fprintf(&sb, "\n\nUse this subject or improve it slightly: %s", s)
		}
	}
	if t := strings.TrimSpace(in.Tone); t != "" {
		fmt.Fprintf(&sb, "\n\nTone: %s", truncateRunes(t, 50))
	}
	return sb.String()
}

func truncateRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n])
}
