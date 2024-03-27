package main

type PageData struct {
    Title string
    Greeting string
    Names []string
    Users []User
}

type User struct {
    Id          int
    Username    string
    Email       string
    Active      bool
    FirstName   string
    LastName    string
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
        Server string `json:"Server"`
        Database string `json:"Database"`
        Username string `json:"Username"`
        Password string `json:"Password"`
    } `json:"Database"`
    AuthInfo struct {
        ClientId string `json:"ClientId"`
        ClientSecret string `json:"ClientSecret"`
        AuthURI string `json:"AuthURI"`
        TokenURI string `json:"TokenURI"`
    } `json:"AuthInfo"`
    SessionKey string `json:"SessionKey"`
}


