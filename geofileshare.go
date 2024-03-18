package main

import (
    "fmt"
    "html/template"
    "net/http"
)

type PageData struct {
    Title string
    Greeting string
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