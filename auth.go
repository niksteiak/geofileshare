package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/gorilla/sessions"
)

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
const GFSBaseUrl = "http://localhost:85"
const SessionCookie = "geofilesession"

var store sessions.Store

func oauthGoogleLogin(w http.ResponseWriter, r *http.Request) {
	oauthstate := generateStateOauthCookie(w, r)
	var googleOauthConfig = getOauthConfig()

	u := googleOauthConfig.AuthCodeURL(oauthstate)
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}

func oauthGoogleCallback(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, SessionCookie)
	oauthstate := session.Values["oauthstate"]

	if r.FormValue("state") != oauthstate {
		log.Println("invalid oauth google state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	data, err := getUserDataFromGoogle(r.FormValue("code"))
	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}

	var userAuthInfo = GoogleUserAuth{}
	err = json.Unmarshal(data, &userAuthInfo)

	user, err := GetUser(userAuthInfo.Email)
	if err != nil {
		log.Printf("User %s not registered or record not found. %s\n", userAuthInfo.Email, err.Error())
		http.Redirect(w, r, "/usererror", http.StatusTemporaryRedirect)
	}

	loginUser(user, w, r)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func loginUser(user User, w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, SessionCookie)
	session.Values["authenticated"] = true
	session.Values["user-id"] = user.Id
	session.Values["username"] = user.Username

	session.Save(r, w)
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, SessionCookie)
	session.Values["authenticated"] = false
	session.Values["user-id"] = ""
	session.Values["username"] = ""

	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func AuthorizationCheck(w http.ResponseWriter, r *http.Request) bool {
	session, _ := store.Get(r, SessionCookie)

	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		return false
	}

	return true
}

func getUserDataFromGoogle(code string) ([]byte, error) {
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

func generateStateOauthCookie(w http.ResponseWriter, r *http.Request) string {

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	session, _ := store.Get(r, SessionCookie)
	session.Values["oauthstate"] = state
	session.Save(r, w)

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
