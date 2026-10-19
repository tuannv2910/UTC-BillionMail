package v1

import (
	"billionmail-core/internal/model/entity"
	"billionmail-core/utility/types/api_v1"

	"github.com/gogf/gf/v2/frame/g"
)

// Enum values (scope, reason, source, status) are documented in internal/service/blacklist/types.go.

type ListBlacklistReq struct {
	g.Meta        `path:"/email_blacklist/list" method:"get" tags:"Email Blacklist" summary:"List blacklist entries"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	Page          int    `json:"page" v:"required|min:1" dc:"Page number"`
	PageSize      int    `json:"page_size" v:"required|min:1|max:500" dc:"Page size"`
	Keyword       string `json:"keyword" dc:"Search in email or domain"`
	Scope         int    `json:"scope" dc:"0: all, 1: email, 2: domain"`
	Reason        int    `json:"reason" dc:"0: all"`
	Source        int    `json:"source" dc:"0: all"`
	Status        int    `json:"status" d:"-1" dc:"-1: all, 0: observing, 1: active, 2: released"`
}

type ListBlacklistRes struct {
	api_v1.StandardRes
	Data struct {
		Total int                      `json:"total"`
		List  []*entity.EmailBlacklist `json:"list"`
	} `json:"data"`
}

type CreateBlacklistReq struct {
	g.Meta        `path:"/email_blacklist/create" method:"post" tags:"Email Blacklist" summary:"Block an email or domain"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	Scope         int    `json:"scope" v:"required|in:1,2" dc:"1: email, 2: domain"`
	Value         string `json:"value" v:"required" dc:"Email address or domain"`
	Reason        int    `json:"reason" v:"required" dc:"Reason"`
	Note          string `json:"note" dc:"Note"`
	TtlDays       int    `json:"ttl_days" d:"-1" v:"min:-1|max:3650" dc:"-1: reason default, 0: permanent, >0: days"`
}

type CreateBlacklistRes struct {
	api_v1.StandardRes
	Data *entity.EmailBlacklist `json:"data"`
}

type ImportBlacklistReq struct {
	g.Meta        `path:"/email_blacklist/import" method:"post" tags:"Email Blacklist" summary:"Block many emails or domains"`
	Authorization string   `json:"authorization" dc:"Authorization" in:"header"`
	Scope         int      `json:"scope" v:"required|in:1,2" dc:"1: email, 2: domain"`
	Values        []string `json:"values" v:"required" dc:"Emails or domains, at most 10000"`
	Reason        int      `json:"reason" v:"required" dc:"Reason"`
	Note          string   `json:"note" dc:"Note"`
	TtlDays       int      `json:"ttl_days" d:"-1" v:"min:-1|max:3650" dc:"-1: reason default, 0: permanent, >0: days"`
}

type InvalidValue struct {
	Value string `json:"value"`
	Error string `json:"error"`
}

type ImportBlacklistRes struct {
	api_v1.StandardRes
	Data struct {
		Inserted int            `json:"inserted"`
		Updated  int            `json:"updated"`
		Invalid  []InvalidValue `json:"invalid"`
	} `json:"data"`
}

type UpdateBlacklistReq struct {
	g.Meta        `path:"/email_blacklist/update" method:"post" tags:"Email Blacklist" summary:"Edit a blacklist entry"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	Id            int64  `json:"id" v:"required|min:1" dc:"ID"`
	Reason        int    `json:"reason" v:"required" dc:"Reason"`
	Note          string `json:"note" dc:"Note"`
	ExpireMode    string `json:"expire_mode" d:"keep" v:"in:keep,permanent,until" dc:"keep: unchanged, permanent, until: use expires_at"`
	ExpiresAt     int64  `json:"expires_at" dc:"Unix time the block ends, used with expire_mode=until"`
}

type UpdateBlacklistRes struct {
	api_v1.StandardRes
	Data *entity.EmailBlacklist `json:"data"`
}

type ReleaseBlacklistReq struct {
	g.Meta        `path:"/email_blacklist/release" method:"post" tags:"Email Blacklist" summary:"Unblock entries and keep their history"`
	Authorization string  `json:"authorization" dc:"Authorization" in:"header"`
	Ids           []int64 `json:"ids" v:"required" dc:"Entry IDs"`
}

type ReleaseBlacklistRes struct {
	api_v1.StandardRes
	Data struct {
		Count int `json:"count"`
	} `json:"data"`
}

type DeleteBlacklistReq struct {
	g.Meta        `path:"/email_blacklist/delete" method:"post" tags:"Email Blacklist" summary:"Delete entries permanently"`
	Authorization string  `json:"authorization" dc:"Authorization" in:"header"`
	Ids           []int64 `json:"ids" v:"required" dc:"Entry IDs"`
}

type DeleteBlacklistRes struct {
	api_v1.StandardRes
	Data struct {
		Count int `json:"count"`
	} `json:"data"`
}

type StatsBlacklistReq struct {
	g.Meta        `path:"/email_blacklist/stats" method:"get" tags:"Email Blacklist" summary:"Blacklist counts by status and reason"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
}

type ReasonCount struct {
	Reason int `json:"reason"`
	Total  int `json:"total"`
}

type StatsBlacklistRes struct {
	api_v1.StandardRes
	Data struct {
		Active    int           `json:"active"`
		Observing int           `json:"observing"`
		Released  int           `json:"released"`
		ByReason  []ReasonCount `json:"by_reason" dc:"Active entries per reason"`
	} `json:"data"`
}
