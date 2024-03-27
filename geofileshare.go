package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gorilla/sessions"
)

var GFSConfig Config

func main() {
	GFSConfig = LoadConfiguration("config/config.json")
	key := []byte(GFSConfig.SessionKey)
	store = sessions.NewCookieStore(key)

	router := http.NewServeMux()
	tmpl := make(map[string]*template.Template)

	fs := http.FileServer(http.Dir("./static/"))
	router.Handle("GET /static/", http.StripPrefix("/static/", fs))

	router.HandleFunc("GET /auth/google/login", oauthGoogleLogin)
	router.HandleFunc("GET /auth/google/callback", oauthGoogleCallback)
	router.HandleFunc("GET /logout", logout)

	router.HandleFunc("GET /greeting", func(w http.ResponseWriter, r *http.Request) {
		tmpl["greeting.html"] = template.Must(template.ParseFiles("templates/greeting.html", "templates/_base.html"))

		data := PageData{
			Title:    "Welcome to Geofileshare",
			Greeting: fmt.Sprintf("Hello, I see you are vistiting the page on %v\n", r.URL.Path),
		}

		tmpl["greeting.html"].ExecuteTemplate(w, "base", data)
	})

	router.HandleFunc("GET /users", Authorize(func(w http.ResponseWriter, r *http.Request) {
		tmpl["dbinfo.html"] = template.Must(template.ParseFiles("templates/dbinfo.html", "templates/_base.html"))

		dbUsers := ReadDatabaseUsers()

		data := PageData{
			Title:    "Registered Users",
			Greeting: "The users that have access to Geofileshare are:",
			Users:    dbUsers,
		}
		tmpl["dbinfo.html"].ExecuteTemplate(w, "base", data)

	}))

	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:    "Welcome to Geofileshare",
			Greeting: "This is home... ...page :)",
		}
		tmpl["index.html"] = template.Must(template.ParseFiles("templates/index.html", "templates/_base.html"))
		tmpl["index.html"].ExecuteTemplate(w, "base", data)
	})


	http.ListenAndServe(":85", router)
}

func Authorize(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authorized := AuthorizationCheck(w, r)
		if !authorized {
			http.Error(w, "Not Authorized", http.StatusForbidden)
			return
		}

		f(w, r)
	}
}

func LoadConfiguration(file string) Config {
	var config Config
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		log.Printf("error opening configuration file: %s\n", err.Error())
	}

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

