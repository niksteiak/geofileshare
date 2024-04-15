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
	"strconv"
	"net/mail"

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
		data.Greeting = fmt.Sprintf("Hello, I see you are vistiting the page on %v://%v%v\n",
			GFSConfig.Protocol, r.Host, r.URL.Path)

		tmpl["greeting.html"].ExecuteTemplate(w, "base", data)
	})

	router.HandleFunc("GET /users", Authorize(true, func(w http.ResponseWriter, r *http.Request) {
		tmpl["dbinfo.html"] = template.Must(template.ParseFiles("templates/dbinfo.html", "templates/_base.html"))

		data := getSessionData(r)
		data.Title ="Registered Users"
		data.Greeting = "The users that have access to Geofileshare are:"
		data.Users = ReadDatabaseUsers()

		tmpl["dbinfo.html"].ExecuteTemplate(w, "base", data)
	}))

	router.HandleFunc("POST /users", Authorize(true, func(w http.ResponseWriter, r *http.Request) {
		tmpl["dbinfo.html"] = template.Must(template.ParseFiles("templates/dbinfo.html", "templates/_base.html"))
		data := getSessionData(r)
		data.Title ="Registered Users"
		data.Greeting = "The users that have access to Geofileshare are:"

		userEmail		:= r.FormValue("email")
		_, err := mail.ParseAddress(userEmail)
		if err != nil {
			data.Users = ReadDatabaseUsers()
			data.ErrorMessage = fmt.Sprintf("Invalid email address. %v", err.Error())
			tmpl["dbinfo.html"].ExecuteTemplate(w, "base", data)
			return
		}

		userFirstName	:= r.FormValue("first_name")
		userLastName	:= r.FormValue("last_name")
		if !(ContainsOnlyLetters(userFirstName) && ContainsOnlyLetters(userLastName)) {
			data.Users = ReadDatabaseUsers()
			data.ErrorMessage = "User First and Last name must contain only letters"
			tmpl["dbinfo.html"].ExecuteTemplate(w, "base", data)
			return
		}
		isAdminValue := r.FormValue("administrator")
		isAdmin := isAdminValue == "on"

		_, err = AddUser(userEmail, userFirstName, userLastName, isAdmin)
		if err != nil {
			data.Users = ReadDatabaseUsers()
			data.ErrorMessage = err.Error()
			tmpl["dbinfo.html"].ExecuteTemplate(w, "base", data)
			return
		}

		http.Redirect(w, r, "/users", http.StatusSeeOther)
	}))

	router.HandleFunc("GET /deleteuser/{id}", Authorize(true, func(w http.ResponseWriter, r *http.Request) {
		tmpl["user.html"] = template.Must(template.ParseFiles("templates/user.html", "templates/_base.html"))
		data := getSessionData(r)
		data.Title ="Delete User"
		data.Greeting = "Are you sure you want to delete this user?"

		id_arg		:= r.PathValue("id")
		userId, err		:= strconv.Atoi(id_arg)
		if err != nil {
			errorMessage := fmt.Sprintf("Error finding user: %s\n", err.Error())
			data.ErrorMessage = errorMessage
			tmpl["user.html"].ExecuteTemplate(w, "base", data)
			return
		}

		userRecord, err := GetUserById(userId)
		data.Users = []User{userRecord}

		tmpl["user.html"].ExecuteTemplate(w, "base", data)
	}))

	router.HandleFunc("POST /deleteuser/{id}", Authorize(true, func(w http.ResponseWriter, r *http.Request) {
		tmpl["user.html"] = template.Must(template.ParseFiles("templates/user.html", "templates/_base.html"))
		data := getSessionData(r)
		data.Title ="Delete User"
		data.Greeting = "Are you sure you want to delete this user?"

		id_arg		:= r.PathValue("id")
		userId, err		:= strconv.Atoi(id_arg)
		if err != nil {
			errorMessage := fmt.Sprintf("Error finding user: %s\n", err.Error())
			data.ErrorMessage = errorMessage
			tmpl["user.html"].ExecuteTemplate(w, "base", data)
			return
		}

		err = DeleteUser(userId)
		if err != nil {
			errorMessage := fmt.Sprintf("Error deleting user: %s\n", err.Error())
			data.ErrorMessage = errorMessage
			tmpl["user.html"].ExecuteTemplate(w, "base", data)
			return
		}
		http.Redirect(w, r, "/users", http.StatusSeeOther)
	}))

	router.HandleFunc("GET /edituser/{id}", Authorize(true, func(w http.ResponseWriter, r *http.Request) {
		tmpl["edituser.html"] = template.Must(template.ParseFiles("templates/edituser.html", "templates/_base.html"))
		data := getSessionData(r)
		data.Title ="Edit User"
		data.Greeting = ""

		id_arg		:= r.PathValue("id")
		userId, err		:= strconv.Atoi(id_arg)
		if err != nil {
			errorMessage := fmt.Sprintf("Error finding user: %s\n", err.Error())
			data.ErrorMessage = errorMessage
			tmpl["edituser.html"].ExecuteTemplate(w, "base", data)
			return
		}
		userRecord, err := GetUserById(userId)
		data.Users = []User{userRecord}

		tmpl["edituser.html"].ExecuteTemplate(w, "base", data)
	}))

	router.HandleFunc("POST /edituser/{id}", Authorize(true, func(w http.ResponseWriter, r *http.Request) {
		tmpl["edituser.html"] = template.Must(template.ParseFiles("templates/edituser.html", "templates/_base.html"))
		data := getSessionData(r)
		data.Title ="Edit User"
		data.Greeting = ""

		id_arg		:= r.PathValue("id")
		userId, err		:= strconv.Atoi(id_arg)
		if err != nil {
			errorMessage := fmt.Sprintf("Error finding user: %s\n", err.Error())
			data.ErrorMessage = errorMessage
			tmpl["edituser.html"].ExecuteTemplate(w, "base", data)
			return
		}
		userRecord, err := GetUserById(userId)
		data.Users = []User{userRecord}

		userRecord.FirstName = r.FormValue("first_name")
		userRecord.LastName = r.FormValue("last_name")
		if !(ContainsOnlyLetters(userRecord.FirstName) && ContainsOnlyLetters(userRecord.LastName)) {
			errorMessage := "User First and Last name must contain only letters"
			data.ErrorMessage = errorMessage
			tmpl["edituser.html"].ExecuteTemplate(w, "base", data)
			return
		}

		isAdminValue := r.FormValue("administrator")
		userRecord.Administrator = isAdminValue == "on"
		isActiveValue := r.FormValue("user_active")
		userRecord.Active = isActiveValue == "on"

		err = UpdateUser(userRecord)
		if err != nil {
			errorMessage := fmt.Sprintf("Error updating user: %s\n", err.Error())
			data.ErrorMessage = errorMessage
			tmpl["edituser.html"].ExecuteTemplate(w, "base", data)
			return
		}
		http.Redirect(w, r, "/users", http.StatusSeeOther)
	}))

	router.HandleFunc("GET /upload", Authorize(false, func(w http.ResponseWriter, r *http.Request) {
		tmpl["upload.html"] = template.Must(template.ParseFiles("templates/upload.html", "templates/_base.html"))

		data := getSessionData(r)
		data.Title = "File Upload"
		data.Greeting = "Upload new File for Sharing"
		data.AllowedFileTypes = GFSConfig.AllowedFileTypes

		tmpl["upload.html"].ExecuteTemplate(w, "base", data)
	}))

	router.HandleFunc("POST /upload", Authorize(false, func(w http.ResponseWriter, r *http.Request) {
		// This will return a json response to indicate to the asychronous uploader whether
		// the upload succeeded or failed and an error message if required
		var response AjaxResponse
		data := getSessionData(r)

		fileInfo, err := uploadFile(r)
		if err != nil {
			response.Status  = "ERROR"
			response.Message = fmt.Sprintf("Upload Error: %s", err.Error())

			json.NewEncoder(w).Encode(response)
			return
		}

		// Save the database record
		fileInfo.RecordId, err = AddUploadRecord(fileInfo, data.User)
		if err != nil {
			response.Status  = "ERROR"
			response.Message = fmt.Sprintf("Error saving upload file record: %s", err.Error())

			json.NewEncoder(w).Encode(response)
			return
		}

		if GFSConfig.SMTP.SendNotifications {
			err = SendMail(r, fileInfo, data.User)
		}

		response.Status = "OK"
		response.Message = fmt.Sprintf("Uploaded file: %v as %v", fileInfo.OriginalFilename, fileInfo.StoredFilename)
		json.NewEncoder(w).Encode(response)
	}))

	router.HandleFunc("GET /files", Authorize(false, func(w http.ResponseWriter, r *http.Request) {
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

		data.Files = &files
		tmpl["files.html"].ExecuteTemplate(w, "base", data)
	}))

	router.HandleFunc("GET /download/{id}/{descriptor}", func(w http.ResponseWriter, r *http.Request) {
		tmpl["file.html"] = template.Must(template.ParseFiles("templates/file.html", "templates/_base.html"))
		data := getSessionData(r)

		id_arg		:= r.PathValue("id")
		descriptor  := r.PathValue("descriptor")
		id, err		:= strconv.Atoi(id_arg)
		if err != nil {
			errorMessage := fmt.Sprintf("Error finding file: %s\n", err.Error())
			data.ErrorMessage = errorMessage
			tmpl["file.html"].ExecuteTemplate(w, "base", data)
			return
		}

		fileInfo, err := GetFileRecord(id, descriptor)
		if err != nil {
			errorMessage := fmt.Sprintf("Error finding file: %s\n", err.Error())
			data.ErrorMessage = errorMessage
			tmpl["file.html"].ExecuteTemplate(w, "base", data)
			return
		}
		storedFilename := filepath.Join(GFSConfig.UploadDirectory,
			fileInfo.StoredFilename)

		// Update the Times Requested Count for the file
		err = UpdateFileRequestedCount(id)
		if err != nil {
			errorMessage := fmt.Sprintf("Error finding file: %s\n", err.Error())
			data.ErrorMessage = errorMessage
			tmpl["file.html"].ExecuteTemplate(w, "base", data)
			return
		}

		// Serve the actual file to the client
		downloadFile, err := os.Open(storedFilename)
		defer downloadFile.Close()
		if err != nil {
			errorMessage := fmt.Sprintf("Error finding file: %s\n", err.Error())
			data.ErrorMessage = errorMessage
			tmpl["file.html"].ExecuteTemplate(w, "base", data)
			return
		}

		contentBuffer := make([]byte, 512)
		downloadFile.Read(contentBuffer)
		fileContentType := http.DetectContentType(contentBuffer)
		fileStat, _ := downloadFile.Stat()
		fileSize := strconv.FormatInt(fileStat.Size(), 10)

		w.Header().Set("Content-Type", fileContentType)
		w.Header().Set("Content-Length", fileSize)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", fileInfo.OriginalFilename))
		downloadFile.Seek(0, 0)
		io.Copy(w, downloadFile)
	})

	router.HandleFunc("GET /delete/{id}/{descriptor}", Authorize(false, func(w http.ResponseWriter, r *http.Request) {
		tmpl["delete.html"] = template.Must(template.ParseFiles("templates/delete.html", "templates/_base.html"))
		data := getSessionData(r)

		id_arg		:= r.PathValue("id")
		descriptor  := r.PathValue("descriptor")
		id, err		:= strconv.Atoi(id_arg)
		if err != nil {
			errorMessage := fmt.Sprintf("Error finding file: %s\n", err.Error())
			data.ErrorMessage = errorMessage
			tmpl["delete.html"].ExecuteTemplate(w, "base", data)
			return
		}

		data.Title = "Delete File"
		data.Greeting = "Are you sure you want to delete this file?"

		fileInfo, err := GetFileRecord(id, descriptor)
		if err != nil {
			errorMessage := fmt.Sprintf("Error finding file: %s\n", err.Error())
			data.ErrorMessage = errorMessage
			tmpl["delete.html"].ExecuteTemplate(w, "base", data)
			return
		}

		data.Files = &[]UploadedFile { fileInfo }
		tmpl["delete.html"].ExecuteTemplate(w, "base", data)

	}))

	router.HandleFunc("POST /delete/{id}/{descriptor}", Authorize(false, func(w http.ResponseWriter, r *http.Request) {
		tmpl["delete.html"] = template.Must(template.ParseFiles("templates/delete.html", "templates/_base.html"))
		data := getSessionData(r)

		id_arg		:= r.PathValue("id")
		descriptor  := r.PathValue("descriptor")
		id, err		:= strconv.Atoi(id_arg)
		if err != nil {
			errorMessage := fmt.Sprintf("Error finding file: %s\n", err.Error())
			data.ErrorMessage = errorMessage
			tmpl["delete.html"].ExecuteTemplate(w, "base", data)
			return
		}

		data.Title = "Delete File"

		fileInfo, err := GetFileRecord(id, descriptor)
		if err != nil {
			errorMessage := fmt.Sprintf("Error finding file: %s\n", err.Error())
			data.ErrorMessage = errorMessage
			tmpl["delete.html"].ExecuteTemplate(w, "base", data)
			return
		}

		err = deleteFile(fileInfo.StoredFilename)
		if err != nil {
			errorMessage := fmt.Sprintf("Error deleting file: %s\n", err.Error())
			data.ErrorMessage = errorMessage
			tmpl["delete.html"].ExecuteTemplate(w, "base", data)
			return
		}

		err = DeleteFileRecord(id)
		if err != nil {
			errorMessage := fmt.Sprintf("Error deleting file record: %s\n", err.Error())
			data.ErrorMessage = errorMessage
			tmpl["delete.html"].ExecuteTemplate(w, "base", data)
			return
		}

		http.Redirect(w, r, "/files", http.StatusSeeOther)
	}))

	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		 if r.URL.Path != "/" {
			errorHandler(w, r, http.StatusNotFound)
			return
		}

		data := getSessionData(r)
		data.Title ="Welcome to Geofileshare"
		data.Greeting = ""

		tmpl["index.html"] = template.Must(template.ParseFiles("templates/index.html", "templates/_base.html"))
		tmpl["index.html"].ExecuteTemplate(w, "base", data)
	})

	log.Fatal(http.ListenAndServe(GFSConfig.ApplicationPort, router))
}

func errorHandler(w http.ResponseWriter, r *http.Request, statusCode int) {
	tmpl := template.Must(template.ParseFiles("templates/error.html"))
	data := getSessionData(r)
	w.WriteHeader(statusCode)

	var errorMessage string
	switch(statusCode) {
	case http.StatusForbidden:
		errorMessage = "Not Authorized to access this section"
	case http.StatusNotFound:
		errorMessage = "Section not found or not available"
	default:
		errorMessage = "Something has gone wrong..."
	}

	data.StatusCode = statusCode
	data.ErrorMessage = errorMessage

	tmpl.Execute(w, data)
}

func getSessionData(r *http.Request) PageData {
	data := PageData{}
	loggedInUser, err := LoggedInUser(r)
	if err != nil {
		// data.ErrorMessage = "User not logged in or user not found"
		data.UserAuthenticated = false
		return data
	}

	data.UserAuthenticated = true
	data.UserAdministrator = loggedInUser.Administrator
	data.User = loggedInUser
	data.ErrorMessage = ""

	data.DownloadBaseUrl = fmt.Sprintf("%v/download",
		GFSConfig.BaseUrl)
	return data
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "/static/images/favicon.ico")
}

func Authorize(mustBeAdmin bool, f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authorized := AuthorizationCheck(w, r, mustBeAdmin)
		if !authorized {
			// http.Error(w, "Not Authorized", http.StatusForbidden)
			errorHandler(w, r, http.StatusForbidden)
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
