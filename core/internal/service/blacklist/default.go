package blacklist

import (
	"sync"

	"github.com/gogf/gf/v2/frame/g"
)

var (
	defaultOnce    sync.Once
	defaultManager *Manager
)

// Default returns the process-wide Manager backed by the default database.
func Default() *Manager {
	defaultOnce.Do(func() {
		defaultManager = NewManager(NewRepository(g.DB()), nil)
	})
	return defaultManager
}
