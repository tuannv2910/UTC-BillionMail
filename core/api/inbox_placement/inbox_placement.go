// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package inbox_placement

import (
	"context"

	"billionmail-core/api/inbox_placement/v1"
)

type IInboxPlacementV1 interface {
	Status(ctx context.Context, req *v1.StatusReq) (res *v1.StatusRes, err error)
	CreateTest(ctx context.Context, req *v1.CreateTestReq) (res *v1.CreateTestRes, err error)
	ListTests(ctx context.Context, req *v1.ListTestsReq) (res *v1.ListTestsRes, err error)
	GetTest(ctx context.Context, req *v1.GetTestReq) (res *v1.GetTestRes, err error)
	CancelTest(ctx context.Context, req *v1.CancelTestReq) (res *v1.CancelTestRes, err error)
	ListSeeds(ctx context.Context, req *v1.ListSeedsReq) (res *v1.ListSeedsRes, err error)
	AddSeed(ctx context.Context, req *v1.AddSeedReq) (res *v1.AddSeedRes, err error)
	SeedOAuth(ctx context.Context, req *v1.SeedOAuthReq) (res *v1.SeedOAuthRes, err error)
	UpdateSeed(ctx context.Context, req *v1.UpdateSeedReq) (res *v1.UpdateSeedRes, err error)
	CheckSeed(ctx context.Context, req *v1.CheckSeedReq) (res *v1.CheckSeedRes, err error)
}
