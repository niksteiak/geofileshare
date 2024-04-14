package main

import(
	"fmt"
	"time"
	"path/filepath"
	"strings"
)

type FileUploadInfo struct {
    OriginalFilename    string
    StoredFilename      string
    FileSize            int64
    RecordId		int64
}

type UploadedFile struct {
    Id          int
    OriginalFilename    string
    StoredFilename      string
    UploadedBy          string
    UploadedById        int
    UploadedOn          time.Time
    Available           bool
    TimesRequested      int
    LastRequested       time.Time
    FileSize            int
}

func (f *UploadedFile) GetDescriptor() string {
    filename := f.StoredFilename
    fileDescriptor := ExtractDescriptor(filename)
    return fileDescriptor
}

func (f *UploadedFile) HasDescriptor(descriptor string) bool {
    fileDescriptor := f.GetDescriptor()

    hasDescriptor := fileDescriptor == descriptor
    return hasDescriptor
}

func ExtractDescriptor(filename string) string {
    fileExtension := filepath.Ext(filename)

    filename      = strings.Replace(filename, fileExtension, "", -1)
    filenameAttrs := strings.Split(filename, "_")
    fileDescriptor := filenameAttrs[len(filenameAttrs)-1]
    return fileDescriptor
}

func (f *UploadedFile) FormattedSize() string {
	var result string

	fileSize := float64(f.FileSize)

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
