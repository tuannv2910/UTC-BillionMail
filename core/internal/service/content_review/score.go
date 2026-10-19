package content_review

import (
	"encoding/json"
	"html"
	"math"
	"strings"
)

const maxIssues = 30

// Score deductions per severity. Low-severity deductions are capped so a pile of nitpicks cannot
// sink an otherwise clean email; any critical issue caps the score.
const (
	deductLow        = 0.5
	deductMedium     = 1.5
	deductHigh       = 3.0
	deductCritical   = 5.0
	maxLowDeduction  = 2.0
	criticalScoreCap = 2.0

	passScore   = 8.0
	reviewScore = 5.0
)

// modelReview is the JSON shape the prompt asks the model for.
type modelReview struct {
	Summary string       `json:"summary"`
	Issues  []modelIssue `json:"issues"`
}

// modelIssue is one issue as the model reported it, before validation. The model's location is
// ignored: the platform derives it from where the excerpt is actually found.
type modelIssue struct {
	Category   string `json:"category"`
	Severity   string `json:"severity"`
	Location   string `json:"location"`
	Excerpt    string `json:"excerpt"`
	Reason     string `json:"reason"`
	Suggestion string `json:"suggestion"`
}

// parseModelReview extracts the JSON object from the reply, tolerating markdown fences or
// stray text around it.
func parseModelReview(reply string) (*modelReview, error) {
	start, end := strings.Index(reply, "{"), strings.LastIndex(reply, "}")
	if start < 0 || end <= start {
		return nil, ErrInvalidModelOutput
	}
	var r modelReview
	if err := json.Unmarshal([]byte(reply[start:end+1]), &r); err != nil {
		return nil, ErrInvalidModelOutput
	}
	if r.Issues == nil && !strings.Contains(reply[start:end+1], `"issues"`) {
		return nil, ErrInvalidModelOutput
	}
	return &r, nil
}

// buildResult validates the model's issues against the email and scores what remains.
// Issues quoting text that is not in the email are dropped, so the model cannot invent problems.
func buildResult(r *modelReview, p preparedEmail, rawContent string) *Result {
	idx := newExcerptIndex(p, rawContent)
	res := &Result{MaxScore: MaxScore, Summary: strings.TrimSpace(r.Summary), Issues: make([]Issue, 0, len(r.Issues))}
	seen := make(map[string]bool)

	for _, raw := range r.Issues {
		issue := Issue{
			Category:   normalizeCategory(raw.Category),
			Severity:   normalizeSeverity(raw.Severity),
			Excerpt:    strings.TrimSpace(raw.Excerpt),
			Reason:     strings.TrimSpace(raw.Reason),
			Suggestion: strings.TrimSpace(raw.Suggestion),
		}
		if issue.Reason == "" {
			res.DroppedIssues++
			continue
		}

		if issue.Excerpt == "" {
			issue.Location = LocationGeneral
		} else if loc, ok := idx.locate(issue.Excerpt); ok {
			issue.Location = loc
		} else {
			res.DroppedIssues++
			continue
		}

		// General issues have no excerpt, so their reason tells them apart.
		key := string(issue.Category) + "|" + normalizeForMatch(issue.Excerpt)
		if issue.Excerpt == "" {
			key += "|" + normalizeForMatch(issue.Reason)
		}
		if seen[key] {
			continue
		}
		seen[key] = true

		if len(res.Issues) == maxIssues {
			res.DroppedIssues++
			continue
		}
		res.Issues = append(res.Issues, issue)
	}

	res.Score, res.Verdict = score(res.Issues)
	return res
}

// score turns issues into a 0-10 score (rounded to 0.5) and a verdict.
func score(issues []Issue) (float64, Verdict) {
	var low, other float64
	hasCritical := false
	for _, i := range issues {
		switch i.Severity {
		case SeverityLow:
			low += deductLow
		case SeverityMedium:
			other += deductMedium
		case SeverityHigh:
			other += deductHigh
		case SeverityCritical:
			other += deductCritical
			hasCritical = true
		}
	}

	s := MaxScore - math.Min(low, maxLowDeduction) - other
	if hasCritical {
		s = math.Min(s, criticalScoreCap)
	}
	s = math.Round(math.Max(s, 0)*2) / 2

	switch {
	case hasCritical || s < reviewScore:
		return s, VerdictFail
	case s < passScore:
		return s, VerdictReview
	default:
		return s, VerdictPass
	}
}

func normalizeCategory(c string) Category {
	cat := Category(strings.ToLower(strings.TrimSpace(c)))
	if validCategories[cat] {
		return cat
	}
	return CategoryOther
}

// normalizeSeverity maps unknown values to medium rather than discarding the issue.
func normalizeSeverity(s string) Severity {
	switch sev := Severity(strings.ToLower(strings.TrimSpace(s))); sev {
	case SeverityLow, SeverityMedium, SeverityHigh, SeverityCritical:
		return sev
	default:
		return SeverityMedium
	}
}

// excerptIndex holds normalized copies of each part of the email for excerpt lookups.
type excerptIndex struct {
	subject string
	links   string
	body    string // visible text plus raw content, so quotes from alt text or attributes match
}

func newExcerptIndex(p preparedEmail, rawContent string) excerptIndex {
	var links strings.Builder
	for _, l := range p.Links {
		links.WriteString(l.Text + "\n" + l.URL + "\n")
	}
	return excerptIndex{
		subject: normalizeForMatch(p.Subject),
		links:   normalizeForMatch(links.String()),
		body:    normalizeForMatch(p.Text + "\n" + rawContent),
	}
}

func (x excerptIndex) locate(excerpt string) (Location, bool) {
	needle := normalizeForMatch(excerpt)
	switch {
	case needle == "":
		return "", false
	case strings.Contains(x.subject, needle):
		return LocationSubject, true
	case strings.Contains(x.links, needle):
		return LocationLink, true
	case strings.Contains(x.body, needle):
		return LocationBody, true
	default:
		return "", false
	}
}

var quoteReplacer = strings.NewReplacer(
	"‘", "'", "’", "'", "“", `"`, "”", `"`, " ", " ", "…", "...",
)

// normalizeForMatch makes excerpt matching tolerant of case, entities, smart quotes and whitespace.
func normalizeForMatch(s string) string {
	s = quoteReplacer.Replace(html.UnescapeString(s))
	return strings.ToLower(collapseSpaces(s))
}
