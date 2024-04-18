package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/mail"
	"strconv"
)

func UsersListHandler(w http.ResponseWriter, r *http.Request, tmpl map[string]*template.Template) {
	tmpl["dbinfo.html"] = template.Must(template.ParseFiles("templates/dbinfo.html", "templates/_base.html"))

	data := getSessionData(r)
	data.Title ="Registered Users"
	data.Greeting = "The users that have access to Geofileshare are:"
	users, err := ReadDatabaseUsers()
	if err != nil {
		data.ErrorMessage = err.Error()
	}
	data.Users = users

	tmpl["dbinfo.html"].ExecuteTemplate(w, "base", data)
}

func AddUserHandler(w http.ResponseWriter, r *http.Request, tmpl map[string]*template.Template) {
	tmpl["dbinfo.html"] = template.Must(template.ParseFiles("templates/dbinfo.html", "templates/_base.html"))
	data := getSessionData(r)
	data.Title ="Registered Users"
	data.Greeting = "The users that have access to Geofileshare are:"

	userEmail		:= r.FormValue("email")
	_, err := mail.ParseAddress(userEmail)
	if err != nil {
		data.ErrorMessage = fmt.Sprintf("Invalid email address. %v", err.Error())
		data.Users, err = ReadDatabaseUsers()
		if err != nil {
			data.ErrorMessage = fmt.Sprintf("%v. %v", data.ErrorMessage, err.Error())
		}
		tmpl["dbinfo.html"].ExecuteTemplate(w, "base", data)
		return
	}

	userFirstName	:= r.FormValue("first_name")
	userLastName	:= r.FormValue("last_name")
	if !(ContainsOnlyLetters(userFirstName) && ContainsOnlyLetters(userLastName)) {
		data.ErrorMessage = "User First and Last name must contain only letters"
		data.Users, err = ReadDatabaseUsers()
		if err != nil {
			data.ErrorMessage = fmt.Sprintf("%v. %v", data.ErrorMessage, err.Error())
		}
		tmpl["dbinfo.html"].ExecuteTemplate(w, "base", data)
		return
	}
	isAdminValue := r.FormValue("administrator")
	isAdmin := isAdminValue == "on"

	_, err = AddUser(userEmail, userFirstName, userLastName, isAdmin)
	if err != nil {
		data.ErrorMessage = err.Error()
		data.Users, err = ReadDatabaseUsers()
		if err != nil {
			data.ErrorMessage = fmt.Sprintf("%v. %v", data.ErrorMessage, err.Error())
		}
		tmpl["dbinfo.html"].ExecuteTemplate(w, "base", data)
		return
	}

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

func UserForDeletionHandler(w http.ResponseWriter, r *http.Request, tmpl map[string]*template.Template) {
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
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request, tmpl map[string]*template.Template) {
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
}

func UserForEditingHandler(w http.ResponseWriter, r *http.Request, tmpl map[string]*template.Template) {
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
}

func EditUserHandler(w http.ResponseWriter, r *http.Request, tmpl map[string]*template.Template) {
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
}
