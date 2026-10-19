// Package content_review asks an LLM to review an email's subject and body for spam triggers
// and policy violations, then scores the result deterministically on a 0-10 scale.
package content_review

import "errors"

// MaxScore is the best possible score.
const MaxScore = 10.0

// Category groups issues by the kind of risk they carry.
type Category string

const (
	CategorySpamTrigger       Category = "spam_trigger"       // phrases filters associate with spam
	CategoryDeceptiveSubject  Category = "deceptive_subject"  // subject misleads about the content
	CategoryMisleadingClaim   Category = "misleading_claim"   // guaranteed or exaggerated results
	CategoryFalseUrgency      Category = "false_urgency"      // pressure or fake scarcity
	CategoryProhibitedContent Category = "prohibited_content" // content providers or law forbid
	CategoryPhishingRisk      Category = "phishing_risk"      // credential/payment requests, impersonation
	CategoryLinkRisk          Category = "link_risk"          // shorteners, mismatched or raw-IP links
	CategoryCompliance        Category = "compliance"         // CAN-SPAM / GDPR / bulk sender rules
	CategoryFormatting        Category = "formatting"         // ALL CAPS, symbol spam, emoji floods
	CategoryOther             Category = "other"
)

var validCategories = map[Category]bool{
	CategorySpamTrigger: true, CategoryDeceptiveSubject: true, CategoryMisleadingClaim: true,
	CategoryFalseUrgency: true, CategoryProhibitedContent: true, CategoryPhishingRisk: true,
	CategoryLinkRisk: true, CategoryCompliance: true, CategoryFormatting: true, CategoryOther: true,
}

// Severity is how much an issue hurts deliverability or compliance.
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// Location is where an issue was found.
type Location string

const (
	LocationSubject Location = "subject"
	LocationBody    Location = "body"
	LocationLink    Location = "link"
	LocationGeneral Location = "general" // whole email, or something that is missing
)

// Verdict is the overall recommendation derived from the score.
type Verdict string

const (
	VerdictPass   Verdict = "pass"   // safe to send
	VerdictReview Verdict = "review" // fix the issues before sending
	VerdictFail   Verdict = "fail"   // do not send as is
)

// Input is the email to review.
type Input struct {
	Subject string
	Content string // HTML or plain text
	// Language for reasons and suggestions: a code (en, vi, zh, ja) or an English language name.
	Language string
	// UnsubscribeManaged is true when the platform appends the unsubscribe link at send time,
	// so a missing link in the content must not be reported.
	UnsubscribeManaged bool
}

// Issue is one problem the reviewer found.
type Issue struct {
	Category   Category `json:"category"`
	Severity   Severity `json:"severity"`
	Location   Location `json:"location"`
	Excerpt    string   `json:"excerpt"` // verbatim text from the email; empty for general issues
	Reason     string   `json:"reason"`
	Suggestion string   `json:"suggestion"`
}

// Result is the outcome of a review.
type Result struct {
	Score         float64 `json:"score"`     // 0-10, higher is safer
	MaxScore      float64 `json:"max_score"` // always 10
	Verdict       Verdict `json:"verdict"`
	Summary       string  `json:"summary"`
	Issues        []Issue `json:"issues"`
	DroppedIssues int     `json:"dropped_issues"` // issues discarded because their excerpt is not in the email
	Model         string  `json:"model"`
	PromptVersion string  `json:"prompt_version"`
	InputTokens   int     `json:"input_tokens"`
	OutputTokens  int     `json:"output_tokens"`
}

var (
	ErrEmptyInput         = errors.New("content_review: subject and content are both empty")
	ErrInvalidModelOutput = errors.New("content_review: model did not return valid review JSON")
)
