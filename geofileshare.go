package main

import (
    "fmt"
    "html/template"
    "net/http"

    _ "github.com/go-sql-driver/mysql"
)

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

