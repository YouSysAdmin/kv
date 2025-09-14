package utils_test

import (
	"path/filepath"
	"runtime"
	"syscall"
	"testing"

	"github.com/yousysadmin/kv/internal/utils"
)

// helper: create a fresh temp directory + file path for a store.
func newTempStorePath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "test.txt")
}

func TestIsOnNetworkStorage(t *testing.T) {
	// check os, windows not sopported.
	if runtime.GOOS == "windows" {
		t.Skip("syscall.Statfs not supported on Windows in this test")
	}
	path := newTempStorePath(t)

	onNet, err := utils.IsOnNetworkStorage(path)
	if err != nil {
		t.Fatalf("IsOnNetworkStorage() = %v", err)
	}
	_ = onNet

	// Statfs type is readable.
	var st syscall.Statfs_t
	if err := syscall.Statfs(filepath.Dir(path), &st); err != nil {
		t.Fatalf("Statfs(dir) = %v", err)
	}
}
