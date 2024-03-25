package main

import (
	"fmt"
	"log"
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"io"
    "time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
const GFSBaseUrl = "http://localhost:85"

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
