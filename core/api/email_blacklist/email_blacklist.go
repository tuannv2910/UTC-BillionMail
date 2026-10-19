// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package email_blacklist

import (
	"context"

	"billionmail-core/api/email_blacklist/v1"
)

type IEmailBlacklistV1 interface {
	ListBlacklist(ctx context.Context, req *v1.ListBlacklistReq) (res *v1.ListBlacklistRes, err error)
	CreateBlacklist(ctx context.Context, req *v1.CreateBlacklistReq) (res *v1.CreateBlacklistRes, err error)
	ImportBlacklist(ctx context.Context, req *v1.ImportBlacklistReq) (res *v1.ImportBlacklistRes, err error)
	UpdateBlacklist(ctx context.Context, req *v1.UpdateBlacklistReq) (res *v1.UpdateBlacklistRes, err error)
	ReleaseBlacklist(ctx context.Context, req *v1.ReleaseBlacklistReq) (res *v1.ReleaseBlacklistRes, err error)
	DeleteBlacklist(ctx context.Context, req *v1.DeleteBlacklistReq) (res *v1.DeleteBlacklistRes, err error)
	StatsBlacklist(ctx context.Context, req *v1.StatsBlacklistReq) (res *v1.StatsBlacklistRes, err error)
}
