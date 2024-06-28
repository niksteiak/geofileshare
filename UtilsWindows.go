//go:build windows

package main

import (
	"golang.org/x/sys/windows"
)

func GetAvailableDiskSpace() (uint64, error) {
	var free uint64
	var total uint64
	var available uint64

	uploadDir := GFSConfig.UploadDirectory
	path, err := windows.UTF16PtrFromString(uploadDir)
	if err != nil {
		return 0, err
	}

	err = windows.GetDiskFreeSpaceEx(path, &free, &total, &available)
	return available, nil
}
