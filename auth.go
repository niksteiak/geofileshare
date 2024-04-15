package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/gorilla/sessions"
)

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
const SessionCookie = "geofilesession"

var store sessions.Store

func LoggedInUser(r *http.Request) (User, error) {
	session, _ := store.Get(r, SessionCookie)

	var user User
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		return user, errors.New("No autheticated User found")
	}

	userEmail := session.Values["email"].(string)
	user, err := GetUser(userEmail)
	if err != nil {
		return user, err
	}

	return user, err
}

func AuthorizationCheck(w http.ResponseWriter, r *http.Request, mustBeAdmin bool) bool {
	if mustBeAdmin {
		user, err := LoggedInUser(r)
		if err != nil {
			log.Println(err.Error())
			return false
		}

		if !user.Administrator {
			return false
		}

	} else {
		session, _ := store.Get(r, SessionCookie)

		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			return false
		}
	}

	return true
}

func loginUser(user User, w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, SessionCookie)
	session.Values["authenticated"] = true
	session.Values["user-id"] = user.Id
	session.Values["username"] = user.Username
	session.Values["email"] = user.Email
	session.Values["admin"] = user.Administrator

	session.Save(r, w)
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, SessionCookie)
	session.Values["authenticated"] = false
	session.Values["user-id"] = ""
	session.Values["username"] = ""
	session.Values["email"] = ""

	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}


func oauthGoogleLogin(w http.ResponseWriter, r *http.Request) {
	oauthstate := generateStateOauthCookie(w, r)
	var googleOauthConfig = getOauthConfig(r)

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

	data, err := getUserDataFromGoogle(r)
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

func getUserDataFromGoogle(r *http.Request) ([]byte, error) {
	var googleOauthConfig = getOauthConfig(r)
	code := r.FormValue("code")

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

func getOauthConfig(r *http.Request) oauth2.Config {
	var oauthConfig = &oauth2.Config{
		RedirectURL:  GFSConfig.AuthInfo.CallbackUrl,
		ClientID:     GFSConfig.AuthInfo.ClientId,
		ClientSecret: GFSConfig.AuthInfo.ClientSecret,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}

	return *oauthConfig
}
