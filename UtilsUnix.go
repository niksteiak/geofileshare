//go:build linux

package main
import (
	"golang.org/x/sys/unix"
)

func GetAvailableDiskSpace() (uint64, error) {
	var stat unix.Statfs_t

	uploadDir := GFSConfig.UploadDirectory
	unix.Statfs(uploadDir, &stat)
	availableSpace := stat.Bavail * uint64(stat.Bsize)
	return availableSpace, nil
}


