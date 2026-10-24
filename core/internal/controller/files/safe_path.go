package files

import (
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

// resolvePathWithin resolves userPath relative to baseDir and guarantees the
// result stays inside baseDir. Absolute paths, ".." traversal and symlink-free
// escapes such as "a/../../etc/passwd" are rejected (#172). Any future file
// read/download handler must go through this instead of using the raw path.
func resolvePathWithin(baseDir, userPath string) (string, error) {
	if strings.TrimSpace(userPath) == "" {
		return "", gerror.New("Invalid path parameter")
	}
	if strings.ContainsRune(userPath, 0) {
		return "", gerror.New("Invalid path parameter")
	}
	if filepath.IsAbs(userPath) || strings.HasPrefix(userPath, "/") || strings.HasPrefix(userPath, `\`) {
		return "", gerror.New("Absolute paths are not allowed")
	}

	base, err := filepath.Abs(baseDir)
	if err != nil {
		return "", gerror.New("Invalid base directory")
	}

	target := filepath.Join(base, filepath.Clean(userPath))
	rel, err := filepath.Rel(base, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", gerror.New("Path is outside the allowed directory")
	}
	return target, nil
}
