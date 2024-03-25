package main

import (
    "fmt"
    "log"
    "database/sql"

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

    rows, err := db.Query("SELECT id, username, active, first_name, last_name FROM user")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    for rows.Next() {
        var u User
        err := rows.Scan(&u.Id, &u.Username, &u.Active, &u.FirstName, &u.LastName)
        if err != nil {
            log.Fatal(err)
        }

        retUsers = append(retUsers, u)
    }

    return retUsers
}
