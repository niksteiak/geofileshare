package main

import(
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
    LastName		string
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
    return FormatFileSize(uint64(f.FileSize))
}
