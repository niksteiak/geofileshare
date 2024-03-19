package main

import (
    "fmt"
    "log"
    "database/sql"

    _ "github.com/go-sql-driver/mysql"
    "github.com/gookit/ini/v2"
)

func ReadConnectionInfo() string {
    err := ini.LoadFiles("config/database.ini")
    if err != nil {
        panic(err)
    }

    dbConnectionInfo := &DatabaseConnection{}
    ini.MapStruct(ini.DefSection(), dbConnectionInfo)

    connectionString := fmt.Sprintf("%v:%v@(%v:3306)/%v?parseTime=true",
        dbConnectionInfo.Username, dbConnectionInfo.Password,
        dbConnectionInfo.Server, dbConnectionInfo.Database)
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
