package database_initialization

import (
	"billionmail-core/internal/service/blacklist"
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

func init() {
	registerHandler(func() {
		for _, sql := range blacklist.SchemaSQL {
			if _, err := g.DB().Exec(context.Background(), sql); err != nil {
				g.Log().Error(context.Background(), "Failed to create email blacklist table:", err)
				return
			}
		}
	})
}
