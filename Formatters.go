package main

import(
	"fmt"
)

func FormatFileSize(rawFileSize uint64) string {
	var result string

	fileSize := float64(rawFileSize)

	switch  {
	case fileSize > 1073741824:
		gbSize := fileSize / 1073741824
		result = fmt.Sprintf("%.2f GB", gbSize)
	case fileSize > 1048576:
		mbSize := fileSize / 1048576
		result = fmt.Sprintf("%.2f MB", mbSize)
	case fileSize > 1024:
		kbSize := fileSize / 1024
		result = fmt.Sprintf("%.2f KB", kbSize)
	default:
		result = fmt.Sprintf("%.2f B", fileSize)
	}

	return result
}
