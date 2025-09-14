package utils

import (
	"path/filepath"
	"syscall"
)

// IsOnNetworkStorage() performs a syscall.Statfs on the parent directory and
// checks known magic constants for NFS/SMB-like filesystems. This is a heuristic
// intended for operators who might care about fsync semantics on network mounts..
func IsOnNetworkStorage(path string) (bool, error) {
	if path == "" {
		return false, nil
	}
	var st syscall.Statfs_t
	if err := syscall.Statfs(filepath.Dir(path), &st); err != nil {
		return false, err
	}
	const (
		NFS_SUPER_MAGIC  = 0x6969
		CIFS_SUPER_MAGIC = 0xFF534D42
		SMB2_MAGIC       = 0xFE534D42
	)
	switch uint64(st.Type) {
	case NFS_SUPER_MAGIC, CIFS_SUPER_MAGIC, SMB2_MAGIC:
		return true, nil
	default:
		return false, nil
	}
}
