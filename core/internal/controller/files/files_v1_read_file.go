package files

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"billionmail-core/api/files/v1"
)

// ReadFile is intentionally not implemented: an unrestricted path parameter
// allowed reading arbitrary server files (#172). A plain GoFrame error replaces
// the previous gRPC status, removing the grpc dependency (#371).
func (c *ControllerV1) ReadFile(ctx context.Context, req *v1.ReadFileReq) (res *v1.ReadFileRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotSupported, "Not yet achieved")
}
