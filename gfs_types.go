package main

type PageData struct {
    Title string
    Greeting string
    Names []string
    Users []User
}

type DatabaseConnection struct {
    Server   string
    Database string
    Username string
    Password string
}

type User struct {
    Id         int
    Username   string
    Active     bool
    FirstName string
    LastName  string
}

