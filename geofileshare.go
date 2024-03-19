package main

import (
    "fmt"
    "log"
    "html/template"
    "net/http"
    "database/sql"

    _ "github.com/go-sql-driver/mysql"
    "github.com/gookit/ini/v2"
)

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

func main() {
    router := http.NewServeMux()
    tmpl := make(map[string]*template.Template)

    fs := http.FileServer(http.Dir("./static/"))
    router.Handle("GET /static/", http.StripPrefix("/static/", fs))

    router.HandleFunc("GET /greeting", func(w http.ResponseWriter, r *http.Request) {
        tmpl["greeting.html"] = template.Must(template.ParseFiles("templates/greeting.html", "templates/_base.html"))

        data := PageData {
            Title: "Welcome to Geofileshare",
            Greeting: fmt.Sprintf("Hello, I see you are vistiting the page on %v\n", r.URL.Path),
        }

        tmpl["greeting.html"].ExecuteTemplate(w, "base", data)
    })

    router.HandleFunc("GET /users", func(w http.ResponseWriter, r *http.Request) {
        tmpl["dbinfo.html"] = template.Must(template.ParseFiles("templates/dbinfo.html", "templates/_base.html"))

        dbUsers := ReadDatabaseUsers()

        data := PageData {
            Title: "Registered Users",
            Greeting: "The users that have access to Geofileshare are:",
            Users: dbUsers,
        }
        tmpl["dbinfo.html"].ExecuteTemplate(w, "base", data)

    })

    router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
        data := PageData {
            Title: "Welcome to Geofileshare",
            Greeting: "This is home... ...page :)",
        }
        tmpl["index.html"] = template.Must(template.ParseFiles("templates/index.html", "templates/_base.html"))
        tmpl["index.html"].ExecuteTemplate(w, "base", data)
    })


    http.ListenAndServe(":85", router)
}

func ReadConnectionInfo() string {
    err := ini.LoadFiles("config/database.ini")
    if err != nil {
        panic(err)
    }

    dbConnectionInfo := &DatabaseConnection{}
    ini.MapStruct(ini.DefSection(), dbConnectionInfo)

    connectionString := fmt.Sprintf("%v:%v@(%v:3306)/%v?parseTime=true",
        dbConnectionInfo.Username, dbConnectionInfo.Password,
        dbConnectionInfo.Server, dbConnectionInfo.Database)
    return connectionString
}

func ReadDatabaseUsers() []User {
    connectionString := ReadConnectionInfo()

    db, err := sql.Open("mysql", connectionString)
    if err != nil {
        log.Fatal(err)
    }

    if err := db.Ping(); err != nil {
        log.Fatal(err)
    }

    var retUsers []User

    rows, err := db.Query("SELECT id, username, active, first_name, last_name FROM user")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    for rows.Next() {
        var u User
        err := rows.Scan(&u.Id, &u.Username, &u.Active, &u.FirstName, &u.LastName)
        if err != nil {
            log.Fatal(err)
        }

        retUsers = append(retUsers, u)
    }

    return retUsers
}
