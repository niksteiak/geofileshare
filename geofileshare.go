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
	"time"

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
		data.Title = "File Upload"

		fileInfo, err := uploadFile(r)
		if err != nil {
			errorMessage := fmt.Sprintf("Error reading upload file: %s\n", err.Error())
			data.ErrorMessage = errorMessage
			tmpl["upload.html"].ExecuteTemplate(w, "base", data)
			return
		}

		// Save the database record
		_, err = AddUploadRecord(fileInfo, data.User)
		if err != nil {
			errorMessage := fmt.Sprintf("Error saving upload file record: %s\n", err.Error())
			data.ErrorMessage = errorMessage
			tmpl["upload.html"].ExecuteTemplate(w, "base", data)
			return
		}

		data.Greeting = "File uploaded.Select new file for Sharing..."

		data.ResponseMessage = fmt.Sprintf("Uploaded file: %v as %v", fileInfo.OriginalFilename, fileInfo.StoredFilename)
		tmpl["upload.html"].ExecuteTemplate(w, "base", data)
	}))

	router.HandleFunc("GET /files", Authorize(func(w http.ResponseWriter, r *http.Request) {
		tmpl["files.html"] = template.Must(template.ParseFiles("templates/files.html", "templates/_base.html"))
		data := getSessionData(r)
		data.Title = "Files Uploaded"

		files, err := UploadedFiles()
		if err != nil {
			errorMessage := fmt.Sprintf("Error reading upload file: %s\n", err.Error())
			data.ErrorMessage = errorMessage
			tmpl["files.html"].ExecuteTemplate(w, "base", data)
			return
		}

		data.Files = files
		tmpl["files.html"].ExecuteTemplate(w, "base", data)
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

func uploadFile(r *http.Request) (FileUploadInfo, error) {
	r.ParseMultipartForm(10 << 20)  // TODO: Check if this works with large files

	var uploadInfo = FileUploadInfo{}

	file, handler, err := r.FormFile("file_upload")
	if err != nil {
		return uploadInfo, err
	}
	defer file.Close()

	filename	  := handler.Filename
	fileExtension := filepath.Ext(filename)
	prefix		  := time.Now().Format("20060102150405")

	uploadInfo.OriginalFilename = filename

	tempFile, err := os.CreateTemp(GFSConfig.UploadDirectory, fmt.Sprintf("gfs_%v_*%v", prefix, fileExtension))
	if err != nil {
		return uploadInfo, err
	}
	defer tempFile.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return uploadInfo, err
	}

	tempFile.Write(fileBytes)

	uploadInfo.StoredFilename = filepath.Base(tempFile.Name())
	return uploadInfo, nil
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

