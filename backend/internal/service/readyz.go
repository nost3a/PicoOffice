package service

import "syscall"

// MinFreeDiskBytes minimum free disk for readiness; below = unhealthy
const MinFreeDiskBytes int64 = 100 * 1024 * 1024 // 100MB

// DiskFreeBytes returns free bytes on path's filesystem (f_bavail).
// extracted so readyz handler and tests share one logic.
func DiskFreeBytes(path string) (int64, error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0, err
	}
	// Bsize is uint32/int64 per platform; cast to int64 to avoid overflow
	avail := int64(st.Bavail) * int64(st.Bsize)
	return avail, nil
}

// DiskHealthy reports whether path's disk has enough free space
func DiskHealthy(path string) (bool, error) {
	avail, err := DiskFreeBytes(path)
	if err != nil {
		return false, err
	}
	return avail >= MinFreeDiskBytes, nil
}
