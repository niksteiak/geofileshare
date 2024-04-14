package main

import (
	"fmt"
	"net/http"

	"gopkg.in/gomail.v2"
)

func SendMail(r *http.Request, fileInfo FileUploadInfo, uploader User) error {
	fileDescriptor := ExtractDescriptor(fileInfo.StoredFilename)
	protocol := GFSConfig.Protocol;

	downloadUrl := fmt.Sprintf("%v://%v/download/%d/%v", protocol,
		r.Host, fileInfo.RecordId, fileDescriptor)
	messageSubject := fmt.Sprintf("GEOFILESHARE - Upload Completed for %v",
		fileInfo.OriginalFilename)

	msg := gomail.NewMessage()
	msg.SetHeader("From", fmt.Sprintf("Geosysta File Share <%v>", GFSConfig.SMTP.SenderAddress))
	msg.SetHeader("To", uploader.Email)
	msg.SetHeader("Subject", messageSubject)

	messageBody := fmt.Sprintf("Download URL for file %v<br/><br/>", fileInfo.OriginalFilename)
	messageBody = fmt.Sprintf("%v<a href=\"%v\">%v</a>", messageBody, downloadUrl, downloadUrl)
	msg.SetBody("text/html", messageBody)

	smtpClient := gomail.NewDialer(GFSConfig.SMTP.SmtpServer, GFSConfig.SMTP.Port,
		GFSConfig.SMTP.Username, GFSConfig.SMTP.Password)

	err := smtpClient.DialAndSend(msg);
	return err
}
