package main
import (
    "fmt"
    "time"
)

type PageData struct {
    Title        string
    Greeting     string
    Names        []string
    Users        []User
    User         User
    UserAuthenticated bool
    ErrorMessage string
    ResponseMessage string
    Files        []UploadedFile
}

type User struct {
    Id          int
    Username    string
    Email       string
    Active      bool
    FirstName   string
    LastName    string
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
    SessionKey string `json:"SessionKey"`
    UploadDirectory string `json:"UploadDirectory"`
}

type FileUploadInfo struct {
    OriginalFilename    string
    StoredFilename      string
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
}
