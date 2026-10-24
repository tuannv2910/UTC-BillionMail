package files

import (
	"path/filepath"
	"testing"
)

func TestResolvePathWithin(t *testing.T) {
	base := t.TempDir()

	ok := []string{"a.txt", "dir/b.txt", "dir/../c.txt", "./d.txt"}
	for _, p := range ok {
		got, err := resolvePathWithin(base, p)
		if err != nil {
			t.Errorf("%q should be allowed: %v", p, err)
			continue
		}
		if rel, _ := filepath.Rel(base, got); rel == ".." || filepath.IsAbs(rel) {
			t.Errorf("%q escaped base: %s", p, got)
		}
	}

	bad := []string{"", "/etc/passwd", "../secret", "a/../../etc/passwd", "..", "dir/\x00x"}
	for _, p := range bad {
		if _, err := resolvePathWithin(base, p); err == nil {
			t.Errorf("%q should be rejected", p)
		}
	}
}
