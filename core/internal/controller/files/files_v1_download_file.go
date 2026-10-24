package files

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"billionmail-core/api/files/v1"
)

// DownloadFile is intentionally not implemented (see ReadFile, #172/#371).
func (c *ControllerV1) DownloadFile(ctx context.Context, req *v1.DownloadFileReq) (res *v1.DownloadFileRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotSupported, "Not yet achieved")
}
