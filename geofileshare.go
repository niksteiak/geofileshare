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
}

type DatabaseConnection struct {
    Server   string
    Database string
    Username string
    Password string
}

func main() {
    router := http.NewServeMux()
    tmpl := make(map[string]*template.Template)

    router.HandleFunc("GET /greeting", func(w http.ResponseWriter, r *http.Request) {
        tmpl["greeting.html"] = template.Must(template.ParseFiles("templates/greeting.html", "templates/_base.html"))

        data := PageData {
            Title: "Welcome to Geofileshare",
            Greeting: fmt.Sprintf("Hello, I see you are vistiting the page on %v\n", r.URL.Path),
        }

        tmpl["greeting.html"].ExecuteTemplate(w, "base", data)
    })

    router.HandleFunc("POST /greeting", func(w http.ResponseWriter, r *http.Request) {
        username := r.FormValue("username")

        data := PageData {
            Title: "Welcome to Geofileshare",
            Greeting: fmt.Sprintf("Hello, %v, good to see you can post here", username),
        }
        tmpl["greeting.html"].ExecuteTemplate(w, "base", data)
    })

    router.HandleFunc("GET /database", func(w http.ResponseWriter, r *http.Request) {

        tmpl["dbinfo.html"] = template.Must(template.ParseFiles("templates/dbinfo.html", "templates/_base.html"))

        dbNames := ReadDatabaseNames()

        data := PageData {
            Title: "Database Information",
            Greeting: "The Names from the database are the following",
            Names: dbNames,
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


    fs := http.FileServer(http.Dir("static/"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))

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

func ReadDatabaseNames() []string {
    connectionString := ReadConnectionInfo()

    db, err := sql.Open("mysql", connectionString)
    if err != nil {
        log.Fatal(err)
    }

    if err := db.Ping(); err != nil {
        log.Fatal(err)
    }

    var retNames []string

    rows, err := db.Query("SELECT `name` FROM project")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    for rows.Next() {
        var curProject string
        err := rows.Scan(&curProject)
        if err != nil {
            log.Fatal(err)
        }

        retNames = append(retNames, curProject)
    }

    return retNames
}
