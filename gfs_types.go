package main
import (
    "fmt"
)

type PageData struct {
    Title        string
    Greeting     string
    Names        []string
    Users        []User
    User         User
    UserAuthenticated bool
    UserAdministrator bool
    ErrorMessage string
    StatusCode   int
    ResponseMessage string
    Files        *[]UploadedFile
    DownloadBaseUrl string
    AllowedFileTypes string
}

type User struct {
    Id          int
    Username    string
    Email       string
    Active      bool
    FirstName   string
    LastName    string
    Administrator bool
}

func (u *User) FullName() string {
    return fmt.Sprintf("%v %v", u.FirstName, u.LastName)
}

type GoogleUserAuth struct {
    Id      string `json:"id"`
    Email       string `json:"email"`
    VerifiedEmail       bool `json:"verified_email"`
    Picture     string `json:"picture"`
    HD      string `json:"hd"`
}

type Config struct {
    Database struct {
        Server   string `json:"Server"`
        Database string `json:"Database"`
        Username string `json:"Username"`
        Password string `json:"Password"`
        TimeZone string `json:"TimeZone"`
    } `json:"Database"`
    AuthInfo struct {
        ClientId     string `json:"ClientId"`
        ClientSecret string `json:"ClientSecret"`
        AuthURI      string `json:"AuthURI"`
        TokenURI     string `json:"TokenURI"`
    } `json:"AuthInfo"`
    SMTP struct {
        SenderAddress   string  `json:"SenderAddress"`
        SenderName      string  `json:"SenderName"`
        SmtpServer      string  `json:"SmtpServer"`
        Port            int     `josn:"Port"`
        UseTLS          bool    `json:"UseTLS"`
        UseSSL          bool    `json:"UseSSL"`
        Username        string  `json:"Username"`
        Password        string  `json:"Password"`
        SendNotifications   bool    `json:"SendNotifications"`
    } `json:"SMTP"`
    SessionKey string `json:"SessionKey"`
    UploadDirectory string `json:"UploadDirectory"`
    AllowedFileTypes string `json:"AllowedFileTypes"`
    Protocol string `json:"Protocol"`
}

type AjaxResponse struct {
    Status      string
    Message     string
}
