package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
const GFSBaseUrl = "http://localhost:85"

var GFSConfig Config

func main() {
	GFSConfig = LoadConfiguration("config/config.json")

	router := http.NewServeMux()
	tmpl := make(map[string]*template.Template)

	fs := http.FileServer(http.Dir("./static/"))
	router.Handle("GET /static/", http.StripPrefix("/static/", fs))

	router.HandleFunc("GET /greeting", func(w http.ResponseWriter, r *http.Request) {
		tmpl["greeting.html"] = template.Must(template.ParseFiles("templates/greeting.html", "templates/_base.html"))

		data := PageData{
			Title:    "Welcome to Geofileshare",
			Greeting: fmt.Sprintf("Hello, I see you are vistiting the page on %v\n", r.URL.Path),
		}

		tmpl["greeting.html"].ExecuteTemplate(w, "base", data)
	})

	router.HandleFunc("GET /users", func(w http.ResponseWriter, r *http.Request) {
		tmpl["dbinfo.html"] = template.Must(template.ParseFiles("templates/dbinfo.html", "templates/_base.html"))

		dbUsers := ReadDatabaseUsers()

		data := PageData{
			Title:    "Registered Users",
			Greeting: "The users that have access to Geofileshare are:",
			Users:    dbUsers,
		}
		tmpl["dbinfo.html"].ExecuteTemplate(w, "base", data)

	})

	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:    "Welcome to Geofileshare",
			Greeting: "This is home... ...page :)",
		}
		tmpl["index.html"] = template.Must(template.ParseFiles("templates/index.html", "templates/_base.html"))
		tmpl["index.html"].ExecuteTemplate(w, "base", data)
	})

	router.HandleFunc("GET /auth/google/login", oauthGoogleLogin)
	router.HandleFunc("GET /auth/google/callback", oauthGoogleCallback)

	http.ListenAndServe(":85", router)
}

func LoadConfiguration(file string) Config {
	var config Config
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		log.Println("error opening configuration file: %s", err.Error())
	}

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

func oauthGoogleLogin(w http.ResponseWriter, r *http.Request) {
	oauthstate := generateStateOauthCookie(w)
	var googleOauthConfig = getOauthConfig()

	u := googleOauthConfig.AuthCodeURL(oauthstate)
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}

func oauthGoogleCallback(w http.ResponseWriter, r *http.Request) {
	oauthState, _ := r.Cookie("oauthstate")

	if r.FormValue("state") != oauthState.Value {
		log.Println("invalid oauth google state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	data, err := getUserDataFromGoogle(r.FormValue("code"), r)
	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}

	fmt.Fprintf(w, "UserInfo: %s\n", data)
}

func getUserDataFromGoogle(code string, r *http.Request) ([]byte, error) {
	var googleOauthConfig = getOauthConfig()

	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange was wrong: %s", err.Error())
	}

	response, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("Failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response: %s", err.Error())
	}

	return contents, nil
}

func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(365 * 24 * time.Hour)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}

func getOauthConfig() oauth2.Config {
	var oauthConfig = &oauth2.Config{
		RedirectURL:  GFSBaseUrl + "/auth/google/callback",
		ClientID:     GFSConfig.AuthInfo.ClientId,
		ClientSecret: GFSConfig.AuthInfo.ClientSecret,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}

	return *oauthConfig
}
