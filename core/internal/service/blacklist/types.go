// Package blacklist manages the email suppression list: recipients and domains that
// must not receive mail because sending to them only hurts sender reputation.
package blacklist

import (
	"errors"
	"time"
)

// TableName is the suppression list table.
const TableName = "bm_email_blacklist"

// Scope says whether an entry blocks one address or a whole domain.
type Scope int

const (
	ScopeEmail  Scope = 1
	ScopeDomain Scope = 2
)

func (s Scope) Valid() bool { return s == ScopeEmail || s == ScopeDomain }

// Reason is why a recipient is suppressed. Values are persisted; never renumber them.
type Reason int

const (
	ReasonInvalidMailbox  Reason = 1  // mailbox does not exist (5.1.1)
	ReasonInvalidDomain   Reason = 2  // domain has no mail host (5.1.2)
	ReasonMailboxDisabled Reason = 3  // account disabled (5.2.1)
	ReasonMailboxFull     Reason = 4  // over quota (5.2.2 / 4.2.2)
	ReasonBlocked         Reason = 5  // rejected by policy (5.7.x); often our IP, not the recipient
	ReasonComplaint       Reason = 6  // recipient reported the mail as spam (FBL)
	ReasonSpamTrap        Reason = 7  // known spam trap address
	ReasonDisposable      Reason = 8  // throwaway mailbox provider
	ReasonRoleAccount     Reason = 9  // shared role address (abuse@, postmaster@ ...)
	ReasonManual          Reason = 10 // added by an operator without a specific cause
	ReasonLegacy          Reason = 11 // migrated from abnormal_recipient without a known cause
)

// Policy controls when a reason starts blocking and for how long.
type Policy struct {
	Threshold int           // hits needed before the entry becomes active
	TTL       time.Duration // block duration once active; 0 means permanent
}

var reasonPolicies = map[Reason]Policy{
	ReasonInvalidMailbox:  {Threshold: 1},
	ReasonInvalidDomain:   {Threshold: 1},
	ReasonMailboxDisabled: {Threshold: 1},
	ReasonMailboxFull:     {Threshold: 3, TTL: 7 * 24 * time.Hour},
	ReasonBlocked:         {Threshold: 3, TTL: 3 * 24 * time.Hour},
	ReasonComplaint:       {Threshold: 1},
	ReasonSpamTrap:        {Threshold: 1},
	ReasonDisposable:      {Threshold: 1},
	ReasonRoleAccount:     {Threshold: 1},
	ReasonManual:          {Threshold: 1},
	ReasonLegacy:          {Threshold: 1},
}

var reasonNames = map[Reason]string{
	ReasonInvalidMailbox:  "invalid_mailbox",
	ReasonInvalidDomain:   "invalid_domain",
	ReasonMailboxDisabled: "mailbox_disabled",
	ReasonMailboxFull:     "mailbox_full",
	ReasonBlocked:         "blocked",
	ReasonComplaint:       "complaint",
	ReasonSpamTrap:        "spam_trap",
	ReasonDisposable:      "disposable",
	ReasonRoleAccount:     "role_account",
	ReasonManual:          "manual",
	ReasonLegacy:          "legacy",
}

func (r Reason) Valid() bool {
	_, ok := reasonPolicies[r]
	return ok
}

func (r Reason) String() string {
	if name, ok := reasonNames[r]; ok {
		return name
	}
	return "unknown"
}

// Policy returns the default threshold and TTL for r.
func (r Reason) Policy() Policy { return reasonPolicies[r] }

// Source is where an entry came from.
type Source int

const (
	SourceBounce    Source = 1
	SourceComplaint Source = 2
	SourceManual    Source = 3
	SourceImport    Source = 4
	SourceAPI       Source = 5
	SourceLegacy    Source = 6
)

func (s Source) Valid() bool { return s >= SourceBounce && s <= SourceLegacy }

// Status is the lifecycle state of an entry.
type Status int

const (
	StatusObserving Status = 0 // seen, below threshold, not blocking
	StatusActive    Status = 1 // blocking
	StatusReleased  Status = 2 // unblocked by an operator, kept for history
)

func (s Status) Valid() bool { return s >= StatusObserving && s <= StatusReleased }

var (
	ErrNotFound      = errors.New("blacklist: entry not found")
	ErrInvalidScope  = errors.New("blacklist: invalid scope")
	ErrInvalidReason = errors.New("blacklist: invalid reason")
	ErrInvalidSource = errors.New("blacklist: invalid source")
	ErrInvalidStatus = errors.New("blacklist: invalid status")
	ErrInvalidEmail  = errors.New("blacklist: invalid email address")
	ErrInvalidDomain = errors.New("blacklist: invalid domain")
	ErrCacheSync     = errors.New("blacklist: saved to database but cache sync failed")
)
