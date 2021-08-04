package session

import (
	"path/filepath"
	"testing"
)

func TestFileStorage(t *testing.T) {
	dir := t.TempDir()
	t.Run("Storage", testStorage(&FileStorage{
		Path: filepath.Join(dir, "session.json"),
	}))
}
