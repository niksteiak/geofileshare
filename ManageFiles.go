package main

import (
	"fmt"
	"errors"
	"path/filepath"
	"os"
	"io"
	"net/http"
	"time"
	"strings"
)


func uploadFile(r *http.Request) (FileUploadInfo, error) {
	r.ParseMultipartForm(10 << 20)  // TODO: Check if this works with large files

	var uploadInfo = FileUploadInfo{}

	file, handler, err := r.FormFile("file_upload")
	if err != nil {
		return uploadInfo, err
	}
	defer file.Close()

	filename	  := handler.Filename
	fileExtension := filepath.Ext(filename)
	if !fileTypeAllowed(fileExtension) {
		errorMessage := fmt.Sprintf("File type not allowed. Only %v file types allowed.", GFSConfig.AllowedFileTypes)
		return uploadInfo, errors.New(errorMessage)
	}

	prefix		  := time.Now().Format("20060102150405")
	uploadInfo.OriginalFilename = filename
	uploadInfo.FileSize = handler.Size

	tempFile, err := os.CreateTemp(GFSConfig.UploadDirectory, fmt.Sprintf("gfs_%v_*%v", prefix, fileExtension))
	if err != nil {
		return uploadInfo, err
	}
	defer tempFile.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return uploadInfo, err
	}

	tempFile.Write(fileBytes)

	uploadInfo.StoredFilename = filepath.Base(tempFile.Name())
	return uploadInfo, nil
}

func deleteFile(filename string) error {
	storedFilename := filepath.Join(GFSConfig.UploadDirectory, filename)
	err := os.Remove(storedFilename)
	return err
}

func fileTypeAllowed(fileExtension string) bool {
	typeAllowed := false
	allowedFileTypes := strings.Split(GFSConfig.AllowedFileTypes, ",")

	for i := range len(allowedFileTypes) {
		if fileExtension == allowedFileTypes[i] {
			typeAllowed = true
			break
		}
	}

	return typeAllowed
}
