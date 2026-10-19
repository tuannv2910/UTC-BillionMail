package content_review

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func issues(sevs ...Severity) []Issue {
	out := make([]Issue, 0, len(sevs))
	for _, s := range sevs {
		out = append(out, Issue{Severity: s})
	}
	return out
}

func TestScore(t *testing.T) {
	cases := []struct {
		name    string
		issues  []Issue
		score   float64
		verdict Verdict
	}{
		{"clean", nil, 10, VerdictPass},
		{"one low", issues(SeverityLow), 9.5, VerdictPass},
		{"lows are capped", issues(SeverityLow, SeverityLow, SeverityLow, SeverityLow, SeverityLow, SeverityLow), 8, VerdictPass},
		{"one medium", issues(SeverityMedium), 8.5, VerdictPass},
		{"two medium", issues(SeverityMedium, SeverityMedium), 7, VerdictReview},
		{"one high", issues(SeverityHigh), 7, VerdictReview},
		{"high and medium", issues(SeverityHigh, SeverityMedium), 5.5, VerdictReview},
		{"two high", issues(SeverityHigh, SeverityHigh), 4, VerdictFail},
		{"critical caps score", issues(SeverityCritical), 2, VerdictFail},
		{"floor at zero", issues(SeverityCritical, SeverityHigh, SeverityHigh, SeverityHigh), 0, VerdictFail},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, v := score(tc.issues)
			assert.Equal(t, tc.score, s)
			assert.Equal(t, tc.verdict, v)
		})
	}
}

func TestParseModelReview(t *testing.T) {
	valid := []string{
		`{"summary":"ok","issues":[]}`,
		"```json\n{\"summary\":\"ok\",\"issues\":[]}\n```",
		`Here is the review: {"summary":"ok","issues":[{"category":"other","severity":"low","excerpt":"x","reason":"r"}]} Done.`,
	}
	for _, reply := range valid {
		r, err := parseModelReview(reply)
		require.NoError(t, err, reply)
		assert.Equal(t, "ok", r.Summary)
	}

	invalid := []string{"", "no json here", `{"summary": "missing issues"}`, `{"issues": [}`, `}{`}
	for _, reply := range invalid {
		_, err := parseModelReview(reply)
		assert.ErrorIs(t, err, ErrInvalidModelOutput, reply)
	}
}

func TestBuildResult_VerifiesExcerpts(t *testing.T) {
	content := `<p>Earn &#36;5000 a week, GUARANTEED!</p><img alt="Free iPhone"><a href="https://bit.ly/x">Claim</a>`
	p := prepareEmail("Re: your invoice", content)

	r, err := parseModelReview(`{
		"summary": "Risky",
		"issues": [
			{"category":"deceptive_subject","severity":"high","location":"body","excerpt":"Re: your invoice","reason":"fake reply","suggestion":"use a real subject"},
			{"category":"misleading_claim","severity":"high","location":"body","excerpt":"earn $5000 a week, guaranteed","reason":"income claim","suggestion":"remove"},
			{"category":"link_risk","severity":"medium","location":"body","excerpt":"https://bit.ly/x","reason":"shortener","suggestion":"full url"},
			{"category":"spam_trigger","severity":"low","location":"body","excerpt":"Free iPhone","reason":"in alt text","suggestion":"remove"},
			{"category":"compliance","severity":"high","location":"general","excerpt":"","reason":"no unsubscribe","suggestion":"add link"},
			{"category":"spam_trigger","severity":"medium","location":"body","excerpt":"Act now before it is gone","reason":"invented","suggestion":"-"},
			{"category":"misleading_claim","severity":"high","location":"body","excerpt":"Earn $5000 a week, GUARANTEED","reason":"duplicate","suggestion":"-"},
			{"category":"weird","severity":"extreme","location":"body","excerpt":"Claim","reason":"unknown labels","suggestion":"-"},
			{"category":"other","severity":"low","location":"body","excerpt":"Claim","reason":"","suggestion":"no reason"}
		]
	}`)
	require.NoError(t, err)

	res := buildResult(r, p, content)

	require.Len(t, res.Issues, 6)
	assert.Equal(t, LocationSubject, res.Issues[0].Location, "location comes from where the excerpt was found")
	assert.Equal(t, LocationBody, res.Issues[1].Location, "entity-encoded text and case differences still match")
	assert.Equal(t, LocationLink, res.Issues[2].Location)
	assert.Equal(t, LocationBody, res.Issues[3].Location, "alt text is matched through the raw content")
	assert.Equal(t, LocationGeneral, res.Issues[4].Location)
	assert.Equal(t, CategoryOther, res.Issues[5].Category, "unknown category falls back to other")
	assert.Equal(t, SeverityMedium, res.Issues[5].Severity, "unknown severity falls back to medium")
	assert.Equal(t, 2, res.DroppedIssues, "invented excerpt and empty reason are dropped; duplicates are merged silently")

	// high x3 = 9, medium x2 = 3, low = 0.5 -> 10 - 12.5 floored at 0
	assert.Equal(t, 0.0, res.Score)
	assert.Equal(t, VerdictFail, res.Verdict)
	assert.Equal(t, MaxScore, res.MaxScore)
	assert.Equal(t, "Risky", res.Summary)
}

func TestBuildResult_GeneralIssuesWithDifferentReasonsAreKept(t *testing.T) {
	raw := &modelReview{Issues: []modelIssue{
		{Category: "compliance", Severity: "high", Reason: "No unsubscribe link"},
		{Category: "compliance", Severity: "medium", Reason: "No postal address"},
		{Category: "compliance", Severity: "high", Reason: "no unsubscribe  link"},
	}}
	res := buildResult(raw, prepareEmail("s", "b"), "b")

	require.Len(t, res.Issues, 2, "same reason is merged, different reasons are kept")
	assert.Equal(t, "No unsubscribe link", res.Issues[0].Reason)
	assert.Equal(t, "No postal address", res.Issues[1].Reason)
	assert.Equal(t, 5.5, res.Score)
}

func TestBuildResult_LimitsIssueCount(t *testing.T) {
	raw := &modelReview{}
	for i := 0; i < maxIssues+3; i++ {
		raw.Issues = append(raw.Issues, modelIssue{Category: "other", Severity: "low", Reason: fmt.Sprintf("reason %d", i)})
	}
	res := buildResult(raw, prepareEmail("s", "b"), "b")
	assert.Len(t, res.Issues, maxIssues)
	assert.Equal(t, 3, res.DroppedIssues)
}
