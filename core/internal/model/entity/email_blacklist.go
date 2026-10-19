package entity

// EmailBlacklist is a suppressed recipient (scope 1) or recipient domain (scope 2).
// Enum values are defined in internal/service/blacklist.
type EmailBlacklist struct {
	Id           int64  `json:"id"            dc:"ID"`
	Scope        int    `json:"scope"         dc:"Scope (1: email, 2: domain)"`
	Value        string `json:"value"         dc:"Normalized email address or domain"`
	Domain       string `json:"domain"        dc:"Recipient domain"`
	Reason       int    `json:"reason"        dc:"Suppression reason"`
	Source       int    `json:"source"        dc:"Where the entry came from"`
	Status       int    `json:"status"        dc:"Status (0: observing, 1: active, 2: released)"`
	HitCount     int    `json:"hit_count"     dc:"Number of bounces/violations seen"`
	Dsn          string `json:"dsn"           dc:"Last DSN code, e.g. 5.1.1"`
	LastResponse string `json:"last_response" dc:"Last SMTP response"`
	LastTaskId   int    `json:"last_task_id"  dc:"Last task that triggered this entry"`
	ExpiresAt    int64  `json:"expires_at"    dc:"Unix time the block ends, 0 means permanent"`
	Note         string `json:"note"          dc:"Note"`
	FirstSeenAt  int64  `json:"first_seen_at" dc:"First seen time"`
	LastSeenAt   int64  `json:"last_seen_at"  dc:"Last seen time"`
	CreateTime   int64  `json:"create_time"   dc:"Create time"`
	UpdateTime   int64  `json:"update_time"   dc:"Update time"`
}
