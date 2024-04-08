package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"io"
	"path/filepath"

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

	router.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/favicon.ico")
	})

	router.HandleFunc("GET /login", oauthGoogleLogin)
	router.HandleFunc("GET /auth/google/callback", oauthGoogleCallback)
	router.HandleFunc("GET /logout", logout)

	router.HandleFunc("GET /greeting", func(w http.ResponseWriter, r *http.Request) {
		tmpl["greeting.html"] = template.Must(template.ParseFiles("templates/greeting.html", "templates/_base.html"))

		data := getSessionData(r)
		data.Title ="Welcome to Geofileshare"
		data.Greeting = fmt.Sprintf("Hello, I see you are vistiting the page on %v\n", r.URL.Path)

		tmpl["greeting.html"].ExecuteTemplate(w, "base", data)
	})

	router.HandleFunc("GET /users", Authorize(func(w http.ResponseWriter, r *http.Request) {
		tmpl["dbinfo.html"] = template.Must(template.ParseFiles("templates/dbinfo.html", "templates/_base.html"))

		data := getSessionData(r)
		data.Title ="Registered Users"
		data.Greeting = "The users that have access to Geofileshare are:"
		data.Users = ReadDatabaseUsers()

		tmpl["dbinfo.html"].ExecuteTemplate(w, "base", data)
	}))

	router.HandleFunc("GET /upload", Authorize(func(w http.ResponseWriter, r *http.Request) {
		tmpl["upload.html"] = template.Must(template.ParseFiles("templates/upload.html", "templates/_base.html"))

		data := getSessionData(r)
		data.Title = "File Upload"
		data.Greeting = "Upload new File for Sharing"

		tmpl["upload.html"].ExecuteTemplate(w, "base", data)
	}))

	router.HandleFunc("POST /upload", Authorize(func(w http.ResponseWriter, r *http.Request) {
		tmpl["upload.html"] = template.Must(template.ParseFiles("templates/upload.html", "templates/_base.html"))
		data := getSessionData(r)

		r.ParseMultipartForm(10 << 20)  // TODO: Check if this works with large files

		file, handler, err := r.FormFile("file_upload")
		if err != nil {
			errorMessage := fmt.Sprintf("error reading upload file: %s\n", err.Error())
			log.Printf(errorMessage)
			data.ErrorMessage = errorMessage
			tmpl["upload.html"].ExecuteTemplate(w, "base", data)
			return
		}
		defer file.Close()

		filename	  := handler.Filename
		fileExtension := filepath.Ext(filename)
		filesize      := handler.Size
		fileheader	  := handler.Header

		data.ResponseMessage = fmt.Sprintf("Uploaded file: %v, size: %v of type %v", filename, fileheader, filesize)

		tempFile, err := os.CreateTemp(GFSConfig.UploadDirectory, fmt.Sprintf("upload-*%v", fileExtension))
		if err != nil {
			log.Printf(err.Error())
		}
		defer tempFile.Close()

		fileBytes, err := io.ReadAll(file)
		if err != nil {
			log.Printf(err.Error())
		}

		tempFile.Write(fileBytes)
		tmpl["upload.html"].ExecuteTemplate(w, "base", data)
	}))

	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		data := getSessionData(r)
		data.Title ="Welcome to Geofileshare"
		data.Greeting = "This is home... ...page :)"

		tmpl["index.html"] = template.Must(template.ParseFiles("templates/index.html", "templates/_base.html"))
		tmpl["index.html"].ExecuteTemplate(w, "base", data)
	})


	http.ListenAndServe(":85", router)
}

func getSessionData(r *http.Request) PageData {
	data := PageData{}
	loggedInUser, err := LoggedInUser(r)
	if err != nil {
		data.ErrorMessage = "User not logged in or user not found"
		data.UserAuthenticated = false
		return data
	}

	data.UserAuthenticated = true
	data.User = loggedInUser
	data.ErrorMessage = ""
	return data
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "/static/images/favicon.ico")
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

