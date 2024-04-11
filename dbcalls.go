package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func ReadConnectionInfo() string {
	connectionString := fmt.Sprintf("%v:%v@(%v:3306)/%v?parseTime=true",
		GFSConfig.Database.Username, GFSConfig.Database.Password,
		GFSConfig.Database.Server, GFSConfig.Database.Database)
	return connectionString
}

func ReadDatabaseUsers() []User {
	connectionString := ReadConnectionInfo()

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	var retUsers []User

	rows, err := db.Query("SELECT id, username, email, active, first_name, last_name FROM user")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var u User
		err := rows.Scan(&u.Id, &u.Username, &u.Email, &u.Active, &u.FirstName, &u.LastName)
		if err != nil {
			log.Fatal(err)
		}

		retUsers = append(retUsers, u)
	}

	return retUsers
}

func GetUser(email string) (User, error) {
	connectionString := ReadConnectionInfo()

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	var user User
	query := "SELECT id, username, email, active, first_name, last_name FROM user where email = ?"
	err = db.QueryRow(query, email).Scan(&user.Id, &user.Username, &user.Email, &user.Active, &user.FirstName, &user.LastName)
	if err != nil {
		log.Fatal(err.Error())
		return user, err
	}

	return user, nil
}

func AddUploadRecord(uploadInfo FileUploadInfo, byUser User) (int64, error) {
	connectionString := ReadConnectionInfo()

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return -1, err
	}

	result, err := db.Exec("INSERT INTO files (added_on, added_by_id, stored_filename, original_filename) VALUES (?, ?, ?, ?)",
			time.Now(), byUser.Id, uploadInfo.StoredFilename, uploadInfo.OriginalFilename)
	if err != nil {
		return -1, err
	}

	uploadId, err := result.LastInsertId()
	return uploadId, err
}
