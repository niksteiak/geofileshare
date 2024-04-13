package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	"errors"

	_ "github.com/go-sql-driver/mysql"
)

func ReadConnectionInfo() string {
	connectionString := fmt.Sprintf("%v:%v@(%v:3306)/%v?parseTime=true&loc=%v",
		GFSConfig.Database.Username, GFSConfig.Database.Password,
		GFSConfig.Database.Server, GFSConfig.Database.Database,
		GFSConfig.Database.TimeZone)
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

	result, err := db.Exec("INSERT INTO files (added_on, added_by_id, stored_filename, original_filename, last_requested, file_size) VALUES (?, ?, ?, ?, ?, ?)",
			time.Now(), byUser.Id, uploadInfo.StoredFilename, uploadInfo.OriginalFilename, time.Now(), uploadInfo.FileSize)
	if err != nil {
		return -1, err
	}

	uploadId, err := result.LastInsertId()
	return uploadId, err
}

func UploadedFiles() ([]UploadedFile, error) {
	connectionString := ReadConnectionInfo()

	var retFiles []UploadedFile

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return retFiles, err
	}

	rows, err := db.Query("SELECT F.id, F.original_filename, F.stored_filename, U.id, CONCAT(U.first_name, ' ', U.last_name) as Fullname, "+
		"F.added_on, F.available, F.times_requested, F.last_requested, F.file_size  "+
		"FROM files F INNER JOIN user U on U.id = F.added_by_id")
	if err != nil {
		return retFiles, err
	}
	defer rows.Close()

	for rows.Next() {
		var f UploadedFile
		err := rows.Scan(&f.Id, &f.OriginalFilename, &f.StoredFilename, &f.UploadedById, &f.UploadedBy,
			&f.UploadedOn, &f.Available, &f.TimesRequested, &f.LastRequested, &f.FileSize)
		if err != nil {
			log.Fatal(err.Error())
			return retFiles, err
		}

		retFiles = append(retFiles, f)
	}

	return retFiles, nil
}

func GetFileRecord(id int, descriptor string) (UploadedFile, error) {
	var fileInfo UploadedFile

	connectionString := ReadConnectionInfo()
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return fileInfo, err
	}

	query := "SELECT F.id, F.original_filename, F.stored_filename, U.id, CONCAT(U.first_name, ' ', U.last_name) as Fullname, "+
		"F.added_on, F.available, F.times_requested, F.last_requested, F.file_size  "+
		"FROM files F INNER JOIN user U on U.id = F.added_by_id "+
		"WHERE F.id = ?"

	err = db.QueryRow(query, id).Scan(&fileInfo.Id, &fileInfo.OriginalFilename,
		&fileInfo.StoredFilename, &fileInfo.UploadedById, &fileInfo.UploadedBy,
		&fileInfo.UploadedOn, &fileInfo.Available, &fileInfo.TimesRequested,
		&fileInfo.LastRequested, &fileInfo.FileSize)
	if err != nil {
		return fileInfo, err
	}

	// Check the descriptor
	if !fileInfo.HasDescriptor(descriptor) {
		return fileInfo, errors.New("Requested file attributes are not valid, or file not found")
	}

	return fileInfo, nil
}

func UpdateFileRequestedCount(id int) error {
	connectionString := ReadConnectionInfo()
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return err
	}

	query := "SELECT F.id, F.original_filename, F.stored_filename, "+
		"F.added_on, F.available, F.times_requested  "+
		"FROM files F WHERE F.id = ?"

	var fileInfo UploadedFile
	err = db.QueryRow(query, id).Scan(&fileInfo.Id, &fileInfo.OriginalFilename,
		&fileInfo.StoredFilename,
		&fileInfo.UploadedOn, &fileInfo.Available, &fileInfo.TimesRequested)
	if err != nil {
		return err
	}

	updatedCount := fileInfo.TimesRequested + 1

	query = "UPDATE files SET times_requested = ?, last_requested = ? WHERE id = ?"
	_, err = db.Exec(query, updatedCount, time.Now(), fileInfo.Id)
	return err
}

func DeleteFileRecord(id int) error {
	connectionString := ReadConnectionInfo()
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return err
	}

	_, err = db.Exec("DELETE FROM files WHERE id = ?", id)
	return err
}
